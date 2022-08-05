// pachage auth provides functionality to enable authentication middleware for auth0
package auth

import (
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Auth is the Auth service struct
type Auth struct {
	Domain   string
	Audience []string

	Middleware *jwtmiddleware.JWTMiddleware
	logger     zerolog.Logger
}

// New creates a new Auth service
func New(domain string, audience []string) *Auth {
	return &Auth{
		Domain:   domain,
		Audience: audience,
		logger:   log.With().Str("component", "auth").Logger(),
	}
}
