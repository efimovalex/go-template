package static

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/iconimpact/replaceme/config"
	"github.com/iconimpact/replaceme/internal/mongodb"
	"github.com/iconimpact/replaceme/internal/redisdb"
	"github.com/iconimpact/replaceme/internal/sqldb"
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

	return &Server{
		cfg: cfg,
	}, nil
}

func (s *Server) Start() int {
	// start static file server
	st := &Static{srv: &http.Server{Addr: ":" + s.cfg.Static.Port}, logger: s.logger}
	go st.Start()

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

				st.Stop()

				return 0
			}

			// break loop
			break
		}
	}
}
