package rest

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	auth "github.com/efimovalex/replaceme/internal/auth0"
)

type CtxKey string

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
		rest.logger.Error().Err(err).Msg("failed to validate the jwt")
		c := rest.Router.NewContext(rq, w)
		_ = rest.JSONError(c, http.StatusUnauthorized, err)
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return middleware, nil
}
