package rest

import (
	"context"
	"net/http"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	auth "github.com/efimovalex/replaceme/internal/auth0"
	"github.com/efimovalex/replaceme/internal/mongodb"
	"github.com/efimovalex/replaceme/internal/redisdb"
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

	rest.srv = &http.Server{Addr: ":" + port, Handler: rest.Router}

	return rest, nil
}

func (rest *R) Start() {
	rest.logger.Info().Msgf("Starting REST service %s", rest.srv.Addr)

	if err := rest.srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		rest.logger.Fatal().Msgf("healthcheck server error: %v", err)
	}
}

func (rest *R) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := rest.srv.Shutdown(ctx); err != nil {
		// Error from closing listeners, or context timeout:
		rest.logger.Error().Err(err).Msg("healthcheck server shutdown error")
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
