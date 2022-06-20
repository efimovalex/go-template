package rest

import (
	"context"
	"net/http"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/iconimpact/go-core/errors"
	auth "github.com/iconimpact/replaceme/internal/auth0"
	"github.com/iconimpact/replaceme/internal/mongodb"
	"github.com/iconimpact/replaceme/internal/redisdb"
	"github.com/iconimpact/replaceme/internal/sqldb"
	"go.uber.org/zap"
)

type REST interface {
	Start()
	Stop()
}

type R struct {
	Router *chi.Mux
	srv    *http.Server

	DB    *sqldb.Client
	Mongo *mongodb.Client
	Redis *redisdb.Client

	AuthMiddleware *jwtmiddleware.JWTMiddleware

	logger *zap.SugaredLogger
}

func New(DB *sqldb.Client, Mongo *mongodb.Client, redis *redisdb.Client, a *auth.Auth, port string, logger *zap.SugaredLogger) (REST, error) {
	rest := &R{
		DB:     DB,
		Mongo:  Mongo,
		Redis:  redis,
		logger: logger,
	}
	var err error
	rest.AuthMiddleware, err = rest.AuthMiddlewareSetup(a)
	if err != nil {
		return nil, errors.E(err, errors.Internal, "unable to setup jwt middleware")
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
	rest.logger.Infow("Starting REST service", "addr", rest.srv.Addr)

	if err := rest.srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		rest.logger.Fatalf("healthcheck server error: %v", err)
	}
}

func (rest *R) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := rest.srv.Shutdown(ctx); err != nil {
		// Error from closing listeners, or context timeout:
		rest.logger.Errorf("healthcheck server shutdown error: %v", err)
	}
}
