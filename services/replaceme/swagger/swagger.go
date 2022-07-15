package swagger

import (
	"context"
	"net/http"
	"time"

	_ "github.com/efimovalex/replaceme/docs/swagger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Swagger interface {
	Start()
	Stop()

	Check(w http.ResponseWriter, r *http.Request)
}
type Ping interface {
	Ping() error
}

type S struct {
	logger zerolog.Logger
	srv    *http.Server

	// Add your depending services that matter to Swagger here
	DB    Ping
	Mongo Ping
	Redis Ping
}

type SwaggerResponse struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
}

func New(port string) *S {
	h := &S{
		srv: &http.Server{Addr: "0.0.0.0:" + port},

		logger: log.With().Str("component", "Swagger").Logger(),
	}

	h.srv.Handler = httpSwagger.Handler(
		httpSwagger.URL(":" + port + "/swagger/doc.json"), //The url pointing to API definition
	)

	return h
}

func (h *S) Start() {
	h.logger.Info().Msgf("Starting swagger service http://%s", h.srv.Addr)
	if err := h.srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		h.logger.Fatal().Msgf("Swagger server error: %v", err)
	}
}

func (h *S) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := h.srv.Shutdown(ctx); err != nil {
		// Error from closing listeners, or context timeout:
		h.logger.Error().Err(err).Msgf("Swagger server shutdown error")
	}
}
