package swagger

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag/example/basic/docs"
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
		srv: &http.Server{Addr: ":" + port},

		logger: log.With().Str("component", "Swagger").Logger(),
	}

	r := gin.Default()
	// programmatically set swagger info
	docs.SwaggerInfo.Title = "replaceme API"
	docs.SwaggerInfo.Description = "This is a replaceme API server."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/v2"
	docs.SwaggerInfo.Schemes = []string{"http"}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	h.srv.Handler = r

	return h
}

func (h *S) Start() {
	h.logger.Info().Msgf("starting ealthcheck service %s", h.srv.Addr)
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
