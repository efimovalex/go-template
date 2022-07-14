package static

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/iconimpact/replaceme/config"
	"github.com/rs/zerolog/log"
)

type Server struct {
	cfg *config.Config

	Static *Static
}

func New(cfg *config.Config, logger *zap.SugaredLogger) (*Server, error) {
	return &Server{
		cfg: cfg,
	}, nil
}

func (s *Server) Start() {
	// start static file server
	s.Static = &Static{srv: &http.Server{Addr: ":" + s.cfg.Static.Port}}
	go s.Static.Start()

	// trap signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// logging
	log.Info().Msg("Entering run loop")
	// run until signal or error

	for sig := range sigChan {
		// log signal
		log.Info().Msgf("Received signal: %d (%s)", sig, sig)

		if sig == syscall.SIGINT || sig == syscall.SIGKILL || sig == syscall.SIGTERM {

			s.Static.Stop()

			return
		}
	}
}
