package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/model"
)

// EndpointRepository defines the interface for endpoint persistence.
type EndpointRepository interface {
	Create(ctx context.Context, endpoint *model.Endpoint) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Endpoint, error)
	GetByMethodPath(ctx context.Context, appID uuid.UUID, method, path, version string) (*model.Endpoint, error)
	Update(ctx context.Context, endpoint *model.Endpoint) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByApp(ctx context.Context, appID uuid.UUID, offset, limit int) ([]*model.Endpoint, int, error)
}
