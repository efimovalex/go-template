package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/efimovalex/replaceme/adapters/mongodb"
	"github.com/efimovalex/replaceme/adapters/postgres"
	"github.com/efimovalex/replaceme/adapters/redisdb"
	"github.com/efimovalex/replaceme/config"
	auth "github.com/efimovalex/replaceme/internal/auth0"
	"github.com/efimovalex/replaceme/services/apis/healthcheck"
	"github.com/efimovalex/replaceme/services/apis/rest"
	"github.com/efimovalex/replaceme/services/apis/swagger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context)
}
type Server struct {
	cfg *config.Config

	DB *postgres.Client

	REST        Service
	HealthCheck Service
	Swagger     Service

	SwaggerUI Service

	logger zerolog.Logger

	sigChan chan os.Signal
}

func New(cfg *config.Config) (*Server, error) {
	db, err := postgres.New(cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Name, cfg.Postgres.SSLMode)
	if err != nil {
		return nil, err
	}

	mongodb, err := mongodb.New(cfg.Mongo.Host, cfg.Mongo.Port, cfg.Mongo.User, cfg.Mongo.Password, cfg.Mongo.Name, cfg.Mongo.SSLMode)
	if err != nil {
		return nil, err
	}

	redis, err := redisdb.New(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.Database)
	if err != nil {
		return nil, err
	}
	claims := auth.New(cfg.Auth.Domain, []string{cfg.Auth.Audience})

	rest, err := rest.New(db, mongodb, redis, claims, cfg.REST.Pretty, cfg.REST.Port)
	if err != nil {
		return nil, err
	}

	return &Server{
		cfg:         cfg,
		DB:          db,
		REST:        rest,
		HealthCheck: healthcheck.New(db, mongodb, redis, cfg.HealthCheck.Port),
		Swagger:     swagger.New(cfg.Swagger.Port, cfg.REST.Port),
		sigChan:     make(chan os.Signal, 1),
		logger:      log.With().Str("component", "server").Logger(),
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)

	defer stop()
	errWg, errCtx := errgroup.WithContext(ctx)

	if s.cfg.Swagger.Enable {
		errWg.Go(func() error {
			return s.Swagger.Start(errCtx)
		})
	}
	// start health check server
	errWg.Go(func() error {
		return s.HealthCheck.Start(errCtx)
	})
	errWg.Go(func() error {
		return s.REST.Start(errCtx)
	})

	errWg.Go(func() error {
		<-ctx.Done()
		s.logger.Info().Msg("stopping server")
		stop()
		s.REST.Stop(ctx)
		s.HealthCheck.Stop(ctx)
		if s.cfg.Swagger.Enable {
			s.Swagger.Stop(ctx)
		}
		return nil
	})

	err := errWg.Wait()

	if err == context.Canceled || err == nil {
		s.logger.Info().Msg("gracefully quit server")

		return nil
	} else if err != nil {
		s.logger.Error().Err(err).Msg("server quit with error")

		return err
	}

	return nil
}
