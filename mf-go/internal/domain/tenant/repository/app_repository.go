package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/model"
)

// AppRepository defines the interface for app persistence.
type AppRepository interface {
	Create(ctx context.Context, app *model.App) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.App, error)
	GetBySlug(ctx context.Context, orgID uuid.UUID, slug string) (*model.App, error)
	Update(ctx context.Context, app *model.App) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByOrg(ctx context.Context, orgID uuid.UUID, offset, limit int) ([]*model.App, int, error)
}
