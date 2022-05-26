package healthcheck

import (
	"context"
	"net/http"
	"time"

	"github.com/iconimpact/go-core/respond"
	"github.com/iconimpact/replaceme/internal/mongodb"
	"github.com/iconimpact/replaceme/internal/redisdb"
	"github.com/iconimpact/replaceme/internal/sqldb"
	"go.uber.org/zap"
)

type Health struct {
	logger *zap.SugaredLogger
	srv    *http.Server

	// Add your depending services that matter to healthcheck here
	DB    *sqldb.Client
	Mongo *mongodb.Client
	Redis *redisdb.Client
}

func New(DB *sqldb.Client, Mongo *mongodb.Client, Redis *redisdb.Client, logger *zap.SugaredLogger, port string) *Health {
	h := &Health{
		DB:    DB,
		Mongo: Mongo,
		Redis: Redis,
		srv:   &http.Server{Addr: ":" + port},

		logger: logger,
	}

	h.srv.Handler = http.HandlerFunc(h.Check)

	return h
}

func (h *Health) Start() {
	h.logger.Infow("starting ealthcheck service", "addr", h.srv.Addr)
	if err := h.srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		h.logger.Fatalf("healthcheck server error: %v", err)
	}
}

func (h *Health) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := h.srv.Shutdown(ctx); err != nil {
		// Error from closing listeners, or context timeout:
		h.logger.Errorf("healthcheck server shutdown error: %v", err)
	}
}

func (h *Health) Check(w http.ResponseWriter, r *http.Request) {
	// Add your depending services that matter to healthcheck here
	var extendErrs []string
	errPostgres := h.DB.Ping()
	if errPostgres != nil {
		h.logger.Errorw("Unable to ping postgres", "error", errPostgres)
		extendErrs = append(extendErrs, "Unable to ping postgres")
	}
	errMongo := h.Mongo.Ping()
	if errMongo != nil {
		h.logger.Errorw("Unable to ping mongo", "error", errMongo)
		extendErrs = append(extendErrs, "Unable to ping mongo")
	}
	errRedis := h.Redis.Ping()
	if errRedis != nil {
		h.logger.Errorw("Unable to ping redis", "error", errRedis)
		extendErrs = append(extendErrs, "Unable to ping redis")
	}

	if len(extendErrs) > 0 {
		respond.JSON(w, h.logger.Desugar(), http.StatusInternalServerError, map[string]interface{}{"errors": extendErrs})
		return
	}

	respond.JSON(w, h.logger.Desugar(), http.StatusOK, map[string]interface{}{"message": "OK"})
}
