package rest

import (
	"context"
	"net"
	"net/http"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/efimovalex/replaceme/adapters/mongodb"
	"github.com/efimovalex/replaceme/adapters/redisdb"
	auth "github.com/efimovalex/replaceme/internal/auth0"
	"github.com/labstack/echo/v4"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type DB interface {
	Ping() error
}

type Router interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	NewContext(r *http.Request, w http.ResponseWriter) echo.Context
}

type R struct {
	Router Router
	srv    *http.Server

	DB    DB
	Mongo *mongodb.Client
	Redis *redisdb.Client

	AuthMiddleware *jwtmiddleware.JWTMiddleware

	logger         zerolog.Logger
	prettyResponse bool
}

func New(DB DB, Mongo *mongodb.Client, redis *redisdb.Client, a *auth.Auth, prettyResponse bool, port string) (*R, error) {
	rest := &R{
		DB:             DB,
		Mongo:          Mongo,
		Redis:          redis,
		logger:         log.With().Str("component", "rest").Logger(),
		prettyResponse: prettyResponse,
	}
	var err error
	rest.AuthMiddleware, err = rest.AuthMiddlewareSetup(a)
	if err != nil {
		return nil, err
	}

	rest.SetupRouter()

	rest.srv = &http.Server{Addr: "0.0.0.0:" + port, Handler: rest.Router}

	return rest, nil
}

func (rest *R) Start(ctx context.Context) error {
	rest.logger.Info().Msgf("Starting REST service %s", rest.srv.Addr)
	lc := net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", rest.srv.Addr)
	if err != nil {
		return err
	}
	if err := rest.srv.Serve(ln); err != http.ErrServerClosed {
		// Error starting or closing listener:
		rest.logger.Fatal().Msgf("healthcheck server error: %v", err)
		return err
	}
	return nil
}

func (rest *R) Stop(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := rest.srv.Shutdown(ctx); err != nil {
		if err == context.Canceled {
			rest.logger.Info().Msg("REST server shutdown gracefully")
		} else {
			rest.logger.Error().Msgf("Swagger server error: %v", err)
		}
	}
}

// JSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json" and
// X-Content-Type-Options as "nosniff".
// Logs the status and v if l is not nil.
func (rest *R) JSON(c echo.Context, status int, v interface{}) error {
	var err error
	if rest.prettyResponse {
		err = c.JSONPretty(status, v, "\t")
	} else {
		err = c.JSON(status, v)
	}
	if err != nil {
		rest.logger.Error().Err(err).Msg("error json encoding response")

		return err
	}
	return nil
}

type errorResponse struct {
	Message string `json:"message"`
}

// JSONError returns an HTTP response as JSON message with status code
// base on app err Kind, Msg from app err HTTPMessage.
// Logs the error if l is not nil.
func (rest *R) JSONError(c echo.Context, status int, err error) error {
	return rest.JSON(c, status, errorResponse{Message: err.Error()})
}
