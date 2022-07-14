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
	auth "github.com/efimovalex/replaceme/internal/auth0"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

		log.Debug().
			Dict("metadata", zerolog.Dict().
				Str("Method", req.Method).
				Str("URL", req.URL.String()).
				Str("UserAgent", req.UserAgent()).
				Str("Referrer", req.Referer()).
				Str("RemoteIP", req.RemoteAddr).
				Str("RequestData", string(bodyBytes)).
				Int64("RequestSize", req.ContentLength).
				Int("ResponseStatus", recorder.Status).
				Str("ResponseData", recorder.Response.String()).
				Int64("ResponseSize", recorder.Count).
				Str("Latency", fmt.Sprintf("%.6fs", time.Since(start).Seconds())),
			).Msg("request received")
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
			rest.JSONError(w, http.StatusUnauthorized, err)

			return
		}

		ctx := auth.WithUserID(r.Context(), cc.GetUserID())
		ctx = auth.WithUserEmail(ctx, cc.GetUserEmail())
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (rest *R) AuthMiddlewareSetup(a *auth.Auth) (*jwtmiddleware.JWTMiddleware, error) {
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
		rest.JSONError(w, http.StatusUnauthorized, err)
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return middleware, nil
}
