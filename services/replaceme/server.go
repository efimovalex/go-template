package server

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/efimovalex/replaceme/config"
	auth "github.com/efimovalex/replaceme/internal/auth0"
	"github.com/efimovalex/replaceme/internal/mongodb"
	"github.com/efimovalex/replaceme/internal/redisdb"
	"github.com/efimovalex/replaceme/internal/sqldb"
	"github.com/efimovalex/replaceme/services/replaceme/healthcheck"
	"github.com/efimovalex/replaceme/services/replaceme/rest"
	"github.com/efimovalex/replaceme/services/replaceme/swagger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Service interface {
	Start()
	Stop()
}
type Server struct {
	cfg *config.Config

	DB *sqldb.Client

	REST        Service
	HealthCheck Service

	SwaggerUI Service

	logger zerolog.Logger

	sigChan chan os.Signal
}

func New(cfg *config.Config) (*Server, error) {
	db, err := sqldb.New(cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Name, cfg.Postgres.SSLMode)
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
	rest, err := rest.New(db, mongodb, redis, claims, cfg.REST.Port)
	if err != nil {
		return nil, err
	}

	return &Server{
		cfg:         cfg,
		DB:          db,
		REST:        rest,
		HealthCheck: healthcheck.New(db, mongodb, redis, cfg.HealthCheck.Port),

		sigChan: make(chan os.Signal, 1),
		logger:  log.With().Str("component", "server").Logger(),
	}, nil
}

func (s *Server) Start() {
	if s.cfg.Swagger.Enable {
		swagger := swagger.New(s.cfg.Swagger.Port)
		defer swagger.Stop()
		go swagger.Start()
	}
	// start health check server
	go s.HealthCheck.Start()
	go s.REST.Start()
	s.checkSignal()
}

func (s *Server) checkSignal() {
	// trap signals
	signal.Notify(s.sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// logging
	s.logger.Info().Msg("Entering run loop")
	// run until signal or error

	for sig := range s.sigChan {
		// log signal
		s.logger.Info().Msgf("Received signal: %d (%s)", sig, sig)

		if sig == syscall.SIGINT || sig == syscall.SIGKILL || sig == syscall.SIGTERM {

			s.HealthCheck.Stop()
			s.REST.Stop()

			return
		}
	}
}
