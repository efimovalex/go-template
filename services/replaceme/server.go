package server

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/iconimpact/replaceme/config"
	"github.com/iconimpact/replaceme/internal/mongodb"
	"github.com/iconimpact/replaceme/internal/redisdb"
	"github.com/iconimpact/replaceme/internal/sqldb"
	"github.com/iconimpact/replaceme/services/replaceme/healthcheck"
	"github.com/iconimpact/replaceme/services/replaceme/rest"
)

type Server struct {
	cfg *config.Config

	DB    *sqldb.Client
	Mongo *mongodb.Client
	Redis *redisdb.Client

	REST *chi.Mux

	logger *zap.SugaredLogger
}

func New(cfg *config.Config, logger *zap.SugaredLogger) (*Server, error) {
	db, err := sqldb.New(cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Name, cfg.Postgres.SSLMode, logger)
	if err != nil {
		return nil, err
	}

	mdb, err := mongodb.New(cfg.Mongo.Host, cfg.Mongo.Port, cfg.Mongo.User, cfg.Mongo.Password, cfg.Mongo.Name, cfg.Mongo.SSL, logger)
	if err != nil {
		return nil, err
	}

	rdb, err := redisdb.New(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.Database, logger)
	if err != nil {
		return nil, err
	}

	return &Server{
		cfg:    cfg,
		DB:     db,
		Mongo:  mdb,
		Redis:  rdb,
		logger: logger,
	}, nil
}

func (s *Server) Start() int {
	// start health check server
	hc := healthcheck.New(s.DB, s.Mongo, s.Redis, s.logger, s.cfg.HealthCheck.Port)
	go hc.Start()

	r := rest.New(s.DB, s.Mongo, s.Redis, s.cfg.REST.Port, s.logger)
	go r.Start()

	// trap signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// logging
	s.logger.Info("Entering run loop")
	// run until signal or error

	for {
		select {
		case sig := <-sigChan:
			// log signal
			s.logger.Infof("Received signal: %d (%s)", sig, sig)

			if sig == syscall.SIGINT || sig == syscall.SIGKILL || sig == syscall.SIGTERM {

				hc.Stop()
				r.Stop()

				return 0
			}

			// break loop
			break
		}
	}
}
