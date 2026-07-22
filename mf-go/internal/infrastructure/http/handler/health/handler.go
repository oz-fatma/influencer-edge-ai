// Package health provides liveness and readiness HTTP probes.
package health

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/masterfabric-go/masterfabric/internal/shared/response"
	"github.com/redis/go-redis/v9"
)

// Handler provides health check endpoints.
type Handler struct {
	db    dbPinger
	redis redisPinger
}

type dbPinger interface {
	Ping(ctx context.Context) error
}

type redisPinger interface {
	Ping(ctx context.Context) *redis.StatusCmd
}

// NewHandler creates a new health handler.
func NewHandler(db *pgxpool.Pool, redis *redis.Client) *Handler {
	var dbP dbPinger
	if db != nil {
		dbP = db
	}
	var redisP redisPinger
	if redis != nil {
		redisP = redis
	}
	return &Handler{db: dbP, redis: redisP}
}

// HealthResponse is the JSON structure for health checks.
type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

// Liveness returns 200 if the server is alive.
func (h *Handler) Liveness(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

// Readiness checks the database and cache connectivity.
func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	services := make(map[string]string)
	healthy := true

	// Check Postgres
	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			slog.Error("readiness check failed", "service", "postgres", "error", err)
			services["postgres"] = "unhealthy"
			healthy = false
		} else {
			services["postgres"] = "healthy"
		}
	}

	// Check Redis
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			slog.Error("readiness check failed", "service", "redis", "error", err)
			services["redis"] = "unhealthy"
			healthy = false
		} else {
			services["redis"] = "healthy"
		}
	}

	status := "ready"
	code := http.StatusOK
	if !healthy {
		status = "not ready"
		code = http.StatusServiceUnavailable
	}

	response.JSON(w, code, HealthResponse{
		Status:   status,
		Services: services,
	})
}
