package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/tenant/dto"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/repository"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
	"github.com/masterfabric-go/masterfabric/internal/shared/middleware"
)

// UpdateWorkspaceUseCase handles workspace updates.
type UpdateWorkspaceUseCase struct {
	workspaceRepo repository.WorkspaceRepository
}

// NewUpdateWorkspaceUseCase creates a new UpdateWorkspaceUseCase.
func NewUpdateWorkspaceUseCase(workspaceRepo repository.WorkspaceRepository) *UpdateWorkspaceUseCase {
	return &UpdateWorkspaceUseCase{workspaceRepo: workspaceRepo}
}

// Execute updates an existing workspace.
func (uc *UpdateWorkspaceUseCase) Execute(ctx context.Context, workspaceID uuid.UUID, req dto.UpdateWorkspaceRequest) (*dto.WorkspaceInfo, error) {
	// Get organization from context
	orgID, ok := middleware.OrgIDFromContext(ctx)
	if !ok {
		return nil, domainErr.New(domainErr.ErrUnauthorized, "organization context required", nil)
	}

	// Get existing workspace
	workspace, err := uc.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Verify workspace belongs to organization
	if workspace.OrganizationID != orgID {
		return nil, domainErr.New(domainErr.ErrForbidden, "workspace does not belong to your organization", nil)
	}

	// Update fields if provided
	if req.Name != "" {
		workspace.Name = strings.TrimSpace(req.Name)
	}
	if req.Slug != "" {
		slug := strings.ToLower(strings.TrimSpace(req.Slug))
		// Check if new slug is taken (excluding current workspace)
		existing, _ := uc.workspaceRepo.GetBySlug(ctx, orgID, slug)
		if existing != nil && existing.ID != workspaceID {
			return nil, domainErr.New(domainErr.ErrAlreadyExists, "workspace slug already taken", nil)
		}
		workspace.Slug = slug
	}
	if req.Description != "" {
		workspace.Description = strings.TrimSpace(req.Description)
	}
	if req.Status != "" {
		workspace.Status = model.WorkspaceStatus(req.Status)
	}

	if err := uc.workspaceRepo.Update(ctx, workspace); err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "failed to update workspace", err)
	}

	return &dto.WorkspaceInfo{
		ID:             workspace.ID,
		OrganizationID: workspace.OrganizationID,
		Name:           workspace.Name,
		Slug:           workspace.Slug,
		Description:    workspace.Description,
		Status:         string(workspace.Status),
		CreatedAt:      workspace.CreatedAt,
		UpdatedAt:      workspace.UpdatedAt,
	}, nil
}
