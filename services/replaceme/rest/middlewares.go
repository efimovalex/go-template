package rest

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/iconimpact/go-core/errors"
	"github.com/iconimpact/go-core/respond"
	auth "github.com/iconimpact/replaceme/internal/auth0"
)

type CtxKey string

const CtxKeyTime = CtxKey("time")

type StatusRecorder struct {
	http.ResponseWriter
	Status   int
	Count    int64
	Response bytes.Buffer
}

func (r *StatusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}
func (r *StatusRecorder) Write(p []byte) (int, error) {
	r.Count += int64(len(p))
	r.Response.Write(p)
	return r.ResponseWriter.Write(p)
}

// LogRequestMiddleware defines a http middleware logs every requests
func (r *R) LogRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()

		recorder := &StatusRecorder{w, 0, 0, bytes.Buffer{}}
		bodyBytes, _ := ioutil.ReadAll(req.Body)
		req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		next.ServeHTTP(recorder, req)

		r.logger.Debugw("Request received",
			"Method", req.Method,
			"URL", req.URL.String(),
			"UserAgent", req.UserAgent(),
			"Referrer", req.Referer(),
			"RemoteIP", req.RemoteAddr,
			"RequestData", string(bodyBytes),
			"RequestSize", req.ContentLength,
			"ResponseStatus", recorder.Status,
			"ResponseData", recorder.Response.String(),
			"ResponseSize", recorder.Count,
			"Latency", fmt.Sprintf("%.6fs", time.Since(start).Seconds()),
		)
	})
}

// adds the current time to the time context value
// should be added first to the Middleware chain
func addTimeContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), CtxKeyTime, time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CORSMiddleware adds the headers to accept CORS requests from everywhere
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// see https://goo.gl/7Qke6k to learn more
		allowedHeaders := "Accept, Authorization, Accept-Language, Accept-Encoding, " +
			"Content-Type, Content-Length, X-CSRF-Token, Session-UUID, Device-UUID"

		// just add CORS headers once
		if r.Header.Get("Access-Control-Allow-Origin") != "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

// UserCtx is an Auth0-based middleware that adds the user ID (from previously
// validated JWT claims)  to the request context.
func (rest *R) UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cc, err := auth.ClaimsValue(r.Context())
		if err != nil {
			err := errors.E(err, errors.Unauthorized, "invalid auth claims")
			respond.JSONError(w, rest.logger.Desugar(), err)

			return
		}

		ctx := auth.WithUserID(r.Context(), cc.GetUserID())
		ctx = auth.WithUserEmail(ctx, cc.GetUserEmail())
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (r *R) AuthMiddlewareSetup(a *auth.Auth) (*jwtmiddleware.JWTMiddleware, error) {
	issuerURL, err := url.Parse("https://" + a.Domain + "/")
	if err != nil {
		return nil, fmt.Errorf("failed to parse the issuer url: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)
	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		a.Audience,
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &auth.CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set up the jwt validator: %v", err)
	}

	errorHandler := func(w http.ResponseWriter, rq *http.Request, err error) {
		err = errors.E(err, errors.Unauthorized, "invalid token")
		respond.JSONError(w, r.logger.Desugar(), err)
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return middleware, nil
}
