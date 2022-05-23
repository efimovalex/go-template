package static

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Static struct {
	srv  *http.Server
	port string

	logger *zap.SugaredLogger
}

func (s *Static) Start() {
	s.srv.Handler = http.FileServer(http.Dir("./static"))
	s.logger.Infow("Starting static service", "addr", s.srv.Addr)

	if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		s.logger.Fatalf("healthcheck server error: %v", err)
	}
}

func (s *Static) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		// Error from closing listeners, or context timeout:
		s.logger.Errorf("healthcheck server shutdown error: %v", err)
	}
}
