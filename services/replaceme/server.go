package server

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/iconimpact/replaceme/config"
	auth "github.com/iconimpact/replaceme/internal/auth0"
	"github.com/iconimpact/replaceme/internal/mongodb"
	"github.com/iconimpact/replaceme/internal/redisdb"
	"github.com/iconimpact/replaceme/internal/sqldb"
	"github.com/iconimpact/replaceme/services/replaceme/healthcheck"
	"github.com/iconimpact/replaceme/services/replaceme/rest"
)

type Server struct {
	cfg *config.Config

	DB *sqldb.Client

	REST        rest.REST
	HealthCheck healthcheck.HealthCheck

	logger *zap.SugaredLogger

	sigChan chan os.Signal
}

func New(cfg *config.Config, logger *zap.SugaredLogger) (*Server, error) {
	db, err := sqldb.New(cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Name, cfg.Postgres.SSLMode, logger)
	if err != nil {
		return nil, err
	}

	mongodb, err := mongodb.New(cfg.Mongo.Host, cfg.Mongo.Port, cfg.Mongo.User, cfg.Mongo.Password, cfg.Mongo.Name, cfg.Mongo.SSLMode, logger)
	if err != nil {
		return nil, err
	}

	redis, err := redisdb.New(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.Database, logger)
	if err != nil {
		return nil, err
	}
	claims := auth.New(cfg.Auth.Domain, []string{cfg.Auth.Audience}, logger)
	rest, err := rest.New(db, mongodb, redis, claims, cfg.REST.Port, logger)
	if err != nil {
		return nil, err
	}

	return &Server{
		cfg:         cfg,
		DB:          db,
		REST:        rest,
		HealthCheck: healthcheck.New(db, mongodb, redis, logger, cfg.HealthCheck.Port),

		sigChan: make(chan os.Signal, 1),
		logger:  logger,
	}, nil
}

func (s *Server) Start() {
	// start health check server
	go s.HealthCheck.Start()
	go s.REST.Start()

	s.checkSignal()
}

func (s *Server) checkSignal() {
	// trap signals
	signal.Notify(s.sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// logging
	s.logger.Info("Entering run loop")
	// run until signal or error

	for sig := range s.sigChan {
		// log signal
		s.logger.Infof("Received signal: %d (%s)", sig, sig)

		if sig == syscall.SIGINT || sig == syscall.SIGKILL || sig == syscall.SIGTERM {

			s.HealthCheck.Stop()
			s.REST.Stop()

			return
		}
	}
}
