package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	adminModel "github.com/masterfabric-go/masterfabric/internal/domain/admin/model"
	"github.com/masterfabric-go/masterfabric/internal/shared/database"
)

// LLMRequestRepository persists outbound LLM (Ollama) call metrics for Grafana.
type LLMRequestRepository struct {
	db               *pgxpool.Pool
	llmRequestsTable string
}

// NewLLMRequestRepository creates a new LLMRequestRepository.
func NewLLMRequestRepository(db *pgxpool.Pool, schema string) *LLMRequestRepository {
	return &LLMRequestRepository{
		db:               db,
		llmRequestsTable: database.QualifyTable(schema, "llm_requests"),
	}
}

// Insert stores a single LLM request log entry.
func (r *LLMRequestRepository) Insert(
	ctx context.Context,
	modelName string,
	promptLength int,
	durationMs int64,
	success bool,
) error {
	_, err := r.db.Exec(ctx,
		fmt.Sprintf(`INSERT INTO %s (id, model_name, prompt_length, duration_ms, success, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`, r.llmRequestsTable),
		uuid.New(),
		modelName,
		promptLength,
		durationMs,
		success,
		time.Now().UTC(),
	)
	return err
}

// ListRecent returns the most recent LLM request log entries.
func (r *LLMRequestRepository) ListRecent(ctx context.Context, limit int) ([]adminModel.LLMRequestLog, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(ctx,
		fmt.Sprintf(`SELECT model_name, duration_ms, success, created_at
		 FROM %s ORDER BY created_at DESC LIMIT $1`, r.llmRequestsTable),
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []adminModel.LLMRequestLog
	for rows.Next() {
		var item adminModel.LLMRequestLog
		if err := rows.Scan(&item.ModelName, &item.DurationMs, &item.Success, &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
