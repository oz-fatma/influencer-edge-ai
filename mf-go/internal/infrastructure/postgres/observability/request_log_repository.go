package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/masterfabric-go/masterfabric/internal/shared/database"
)

// RequestLogRepository persists HTTP request logs for Grafana dashboards.
type RequestLogRepository struct {
	db              *pgxpool.Pool
	requestLogsTable string
}

// NewRequestLogRepository creates a new RequestLogRepository.
func NewRequestLogRepository(db *pgxpool.Pool, schema string) *RequestLogRepository {
	return &RequestLogRepository{
		db:               db,
		requestLogsTable: database.QualifyTable(schema, "request_logs"),
	}
}

// Insert stores a single HTTP request log entry.
func (r *RequestLogRepository) Insert(
	ctx context.Context,
	method, path string,
	statusCode int,
	durationMs int64,
) error {
	_, err := r.db.Exec(ctx,
		fmt.Sprintf(`INSERT INTO %s (id, method, path, status_code, duration_ms, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`, r.requestLogsTable),
		uuid.New(),
		method,
		path,
		statusCode,
		durationMs,
		time.Now().UTC(),
	)
	return err
}
