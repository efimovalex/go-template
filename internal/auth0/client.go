package auth

import (
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Auth struct {
	Domain   string
	Audience []string

	Middleware *jwtmiddleware.JWTMiddleware
	logger     zerolog.Logger
}

func New(domain string, audience []string) *Auth {
	return &Auth{
		Domain:   domain,
		Audience: audience,
		logger:   log.With().Str("component", "auth").Logger(),
	}
}
