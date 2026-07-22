package observability

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RequestLogRepository persists HTTP request logs for Grafana dashboards.
type RequestLogRepository struct {
	db *pgxpool.Pool
}

// NewRequestLogRepository creates a new RequestLogRepository.
func NewRequestLogRepository(db *pgxpool.Pool) *RequestLogRepository {
	return &RequestLogRepository{db: db}
}

// Insert stores a single HTTP request log entry.
func (r *RequestLogRepository) Insert(
	ctx context.Context,
	method, path string,
	statusCode int,
	durationMs int64,
) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO request_logs (id, method, path, status_code, duration_ms, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(),
		method,
		path,
		statusCode,
		durationMs,
		time.Now().UTC(),
	)
	return err
}
