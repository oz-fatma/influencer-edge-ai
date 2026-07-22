package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/model"
)

// PolicyRepository defines the interface for endpoint policy persistence.
type PolicyRepository interface {
	Create(ctx context.Context, policy *model.EndpointPolicy) error
	GetByEndpointID(ctx context.Context, endpointID uuid.UUID) (*model.EndpointPolicy, error)
	Update(ctx context.Context, policy *model.EndpointPolicy) error
	Delete(ctx context.Context, id uuid.UUID) error
}
