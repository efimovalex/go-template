package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	auth "github.com/efimovalex/replaceme/internal/auth0"
	"github.com/efimovalex/replaceme/internal/mongodb"
	"github.com/efimovalex/replaceme/internal/redisdb"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type REST interface {
	Start()
	Stop()
}

type DB interface {
	Ping() error
}

type R struct {
	Router *chi.Mux
	srv    *http.Server

	DB    DB
	Mongo *mongodb.Client
	Redis *redisdb.Client

	AuthMiddleware *jwtmiddleware.JWTMiddleware

	logger zerolog.Logger
}

func New(DB DB, Mongo *mongodb.Client, redis *redisdb.Client, a *auth.Auth, port string) (REST, error) {
	rest := &R{
		DB:     DB,
		Mongo:  Mongo,
		Redis:  redis,
		logger: log.With().Str("component", "rest").Logger(),
	}
	var err error
	rest.AuthMiddleware, err = rest.AuthMiddlewareSetup(a)
	if err != nil {
		return nil, err
	}

	mux := chi.NewRouter()

	// Add middlewares
	mux.Use(rest.LogRequestMiddleware)
	mux.Use(addTimeContextMiddleware) // used for request-time and action-time headers
	//r.Use(timeTrackingMiddleware)
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Timeout(180 * time.Second))

	mux.Use(CORSMiddleware)

	rest.Router = mux
	rest.AddRoutes()

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
func (rest *R) JSON(w http.ResponseWriter, status int, v interface{}) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		rest.logger.Error().Err(err).Msg("error json encoding response")

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)

	_, err = w.Write(jsonBytes)
	if err != nil {
		rest.logger.Error().Err(err).Msg("error writing response")
	}
}

type errorResponse struct {
	Message string `json:"message"`
}

// JSONError returns an HTTP response as JSON message with status code
// base on app err Kind, Msg from app err HTTPMessage.
// Logs the error if l is not nil.
func (rest *R) JSONError(w http.ResponseWriter, status int, err error) {
	rest.JSON(w, status, errorResponse{Message: err.Error()})
}
