package static

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/iconimpact/replaceme/config"
)

type Server struct {
	cfg    *config.Config
	logger *zap.SugaredLogger

	Static *Static
}

func New(cfg *config.Config, logger *zap.SugaredLogger) (*Server, error) {
	return &Server{
		cfg:    cfg,
		logger: logger,
	}, nil
}

func (s *Server) Start() {
	// start static file server
	s.Static = &Static{srv: &http.Server{Addr: ":" + s.cfg.Static.Port}, logger: s.logger}
	go s.Static.Start()

	// trap signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// logging
	s.logger.Info("Entering run loop")
	// run until signal or error

	for sig := range sigChan {
		// log signal
		s.logger.Infof("Received signal: %d (%s)", sig, sig)

		if sig == syscall.SIGINT || sig == syscall.SIGKILL || sig == syscall.SIGTERM {

			s.Static.Stop()

			return
		}
	}
}
