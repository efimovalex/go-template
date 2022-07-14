package static

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type Static struct {
	srv *http.Server
}

func (s *Static) Start() {
	s.srv.Handler = http.FileServer(http.Dir("./static"))
	log.Info().Msgf("Starting static service: %s", s.srv.Addr)

	if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatal().Msgf("healthcheck server error: %v", err)
	}
}

func (s *Static) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		// Error from closing listeners, or context timeout:
		log.Error().Err(err).Msgf("healthcheck server shutdown error")
	}
}
