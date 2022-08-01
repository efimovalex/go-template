package healthcheck

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type HealthCheck interface {
	Start()
	Stop()

	Check(w http.ResponseWriter, r *http.Request)
}
type Ping interface {
	Ping() error
}

type Health struct {
	logger zerolog.Logger
	srv    *http.Server

	// Add your depending services that matter to healthcheck here
	DB    Ping
	Mongo Ping
	Redis Ping
}

type HealthCheckResponse struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
}

func New(DB Ping, Mongo Ping, Redis Ping, port string) *Health {
	h := &Health{
		DB:    DB,
		Mongo: Mongo,
		Redis: Redis,
		srv:   &http.Server{Addr: ":" + port},

		logger: log.With().Str("component", "healthcheck").Logger(),
	}

	h.srv.Handler = http.HandlerFunc(h.Check)

	return h
}

func (h *Health) Start(ctx context.Context) error {
	h.logger.Info().Msgf("Starting healthcheck service %s", h.srv.Addr)
	lc := net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", h.srv.Addr)
	if err != nil {
		return err
	}
	if err := h.srv.Serve(ln); err != http.ErrServerClosed {
		// Error starting or closing listener:
		h.logger.Fatal().Msgf("healthcheck server error: %v", err)
		return err
	}

	return nil
}

func (h *Health) Stop(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := h.srv.Shutdown(ctx); err != nil {
		if err == context.Canceled {
			h.logger.Info().Msg("Healthcheck server shutdown gracefully")
		} else {
			h.logger.Error().Msgf("Swagger server error: %v", err)
		}
	}
}

func (h *Health) Check(w http.ResponseWriter, r *http.Request) {
	// Add your depending services that matter to healthcheck here
	var extendErrs []string
	errPostgres := h.DB.Ping()
	if errPostgres != nil {
		h.logger.Error().Err(errPostgres).Msgf("Unable to ping postgres")
		extendErrs = append(extendErrs, "Unable to ping postgres")
	}
	errMongo := h.Mongo.Ping()
	if errMongo != nil {
		h.logger.Error().Err(errMongo).Msgf("Unable to ping mongo")
		extendErrs = append(extendErrs, "Unable to ping mongo")
	}
	errRedis := h.Redis.Ping()
	if errRedis != nil {
		h.logger.Error().Err(errRedis).Msgf("Unable to ping redis")
		extendErrs = append(extendErrs, "Unable to ping redis")
	}

	if len(extendErrs) > 0 {
		h.JSON(w, http.StatusInternalServerError, HealthCheckResponse{Message: "healthcheck failed", Errors: extendErrs})
		return
	}

	h.JSON(w, http.StatusOK, HealthCheckResponse{Message: "OK"})
}

func (h *Health) JSON(w http.ResponseWriter, status int, v interface{}) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		h.logger.Error().Err(err).Msgf("Unable to marshal response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)

	_, err = w.Write(jsonBytes)
	if err != nil {
		h.logger.Error().Err(err).Msg("error writing response")
	}
}
