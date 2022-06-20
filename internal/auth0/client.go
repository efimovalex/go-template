package auth

import (
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"go.uber.org/zap"
)

type Auth struct {
	Domain   string
	Audience []string

	Middleware *jwtmiddleware.JWTMiddleware

	logger *zap.SugaredLogger
}

func New(domain string, audience []string, logger *zap.SugaredLogger) *Auth {
	return &Auth{
		Domain:   domain,
		Audience: audience,
		logger:   logger,
	}
}
