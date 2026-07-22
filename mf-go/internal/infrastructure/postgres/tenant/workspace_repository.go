package tenant

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/repository"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
)

var _ repository.WorkspaceRepository = (*WorkspaceRepository)(nil)

// WorkspaceRepository implements the workspace repository interface.
type WorkspaceRepository struct {
	db *pgxpool.Pool
}

// NewWorkspaceRepository creates a new workspace repository.
func NewWorkspaceRepository(db *pgxpool.Pool) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

// Create creates a new workspace.
func (r *WorkspaceRepository) Create(ctx context.Context, workspace *model.Workspace) error {
	query := `
		INSERT INTO workspaces (id, organization_id, name, slug, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	now := time.Now()
	workspace.CreatedAt = now
	workspace.UpdatedAt = now
	if workspace.Status == "" {
		workspace.Status = model.WorkspaceStatusActive
	}

	_, err := r.db.Exec(ctx, query,
		workspace.ID,
		workspace.OrganizationID,
		workspace.Name,
		workspace.Slug,
		workspace.Description,
		workspace.Status,
		workspace.CreatedAt,
		workspace.UpdatedAt,
	)
	return err
}

// GetByID retrieves a workspace by ID.
func (r *WorkspaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Workspace, error) {
	query := `
		SELECT id, organization_id, name, slug, description, status, created_at, updated_at
		FROM workspaces
		WHERE id = $1
	`
	var w model.Workspace
	err := r.db.QueryRow(ctx, query, id).Scan(
		&w.ID,
		&w.OrganizationID,
		&w.Name,
		&w.Slug,
		&w.Description,
		&w.Status,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "workspace not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get workspace", err)
	}
	return &w, nil
}

// GetBySlug retrieves a workspace by organization ID and slug.
func (r *WorkspaceRepository) GetBySlug(ctx context.Context, orgID uuid.UUID, slug string) (*model.Workspace, error) {
	query := `
		SELECT id, organization_id, name, slug, description, status, created_at, updated_at
		FROM workspaces
		WHERE organization_id = $1 AND slug = $2
	`
	var w model.Workspace
	err := r.db.QueryRow(ctx, query, orgID, slug).Scan(
		&w.ID,
		&w.OrganizationID,
		&w.Name,
		&w.Slug,
		&w.Description,
		&w.Status,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "workspace not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get workspace", err)
	}
	return &w, nil
}

// ListByOrganization lists all workspaces for an organization.
func (r *WorkspaceRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*model.Workspace, error) {
	query := `
		SELECT id, organization_id, name, slug, description, status, created_at, updated_at
		FROM workspaces
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []*model.Workspace
	for rows.Next() {
		var w model.Workspace
		if err := rows.Scan(
			&w.ID,
			&w.OrganizationID,
			&w.Name,
			&w.Slug,
			&w.Description,
			&w.Status,
			&w.CreatedAt,
			&w.UpdatedAt,
		); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, &w)
	}

	return workspaces, rows.Err()
}

// Update updates an existing workspace.
func (r *WorkspaceRepository) Update(ctx context.Context, workspace *model.Workspace) error {
	query := `
		UPDATE workspaces
		SET name = $2, slug = $3, description = $4, status = $5, updated_at = $6
		WHERE id = $1
	`
	workspace.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx, query,
		workspace.ID,
		workspace.Name,
		workspace.Slug,
		workspace.Description,
		workspace.Status,
		workspace.UpdatedAt,
	)
	return err
}

// Delete soft-deletes a workspace (sets status to archived).
func (r *WorkspaceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE workspaces
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, model.WorkspaceStatusArchived, time.Now(), id)
	return err
}
