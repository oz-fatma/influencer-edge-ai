package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/model"
)

// WorkspaceRepository defines the interface for workspace persistence.
type WorkspaceRepository interface {
	// Create creates a new workspace.
	Create(ctx context.Context, workspace *model.Workspace) error

	// GetByID retrieves a workspace by ID.
	GetByID(ctx context.Context, id uuid.UUID) (*model.Workspace, error)

	// GetBySlug retrieves a workspace by organization ID and slug.
	GetBySlug(ctx context.Context, orgID uuid.UUID, slug string) (*model.Workspace, error)

	// ListByOrganization lists all workspaces for an organization.
	ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*model.Workspace, error)

	// Update updates an existing workspace.
	Update(ctx context.Context, workspace *model.Workspace) error

	// Delete soft-deletes a workspace (sets status to archived).
	Delete(ctx context.Context, id uuid.UUID) error
}
