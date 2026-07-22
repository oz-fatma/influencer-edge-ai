package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/model"
)

// OrgRepository defines the interface for organization persistence.
type OrgRepository interface {
	Create(ctx context.Context, org *model.Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error)
	GetBySlug(ctx context.Context, slug string) (*model.Organization, error)
	Update(ctx context.Context, org *model.Organization) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*model.Organization, int, error)
}
