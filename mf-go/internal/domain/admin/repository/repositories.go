package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/admin/model"
)

type LLMConfigRepository interface {
	Get(ctx context.Context) (*model.LLMConfig, error)
	Update(ctx context.Context, cfg *model.LLMConfig) (*model.LLMConfig, error)
}

type AdminRepository interface {
	IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
}

type LLMRequestLogRepository interface {
	ListRecent(ctx context.Context, limit int) ([]model.LLMRequestLog, error)
}
