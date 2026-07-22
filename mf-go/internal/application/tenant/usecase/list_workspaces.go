package usecase

import (
	"context"

	"github.com/masterfabric-go/masterfabric/internal/application/tenant/dto"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/repository"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
	"github.com/masterfabric-go/masterfabric/internal/shared/middleware"
)

// ListWorkspacesUseCase handles listing workspaces for an organization.
type ListWorkspacesUseCase struct {
	workspaceRepo repository.WorkspaceRepository
}

// NewListWorkspacesUseCase creates a new ListWorkspacesUseCase.
func NewListWorkspacesUseCase(workspaceRepo repository.WorkspaceRepository) *ListWorkspacesUseCase {
	return &ListWorkspacesUseCase{workspaceRepo: workspaceRepo}
}

// Execute lists all workspaces for the current organization.
func (uc *ListWorkspacesUseCase) Execute(ctx context.Context) ([]*dto.WorkspaceInfo, error) {
	// Get organization from context
	orgID, ok := middleware.OrgIDFromContext(ctx)
	if !ok {
		return nil, domainErr.New(domainErr.ErrUnauthorized, "organization context required", nil)
	}

	workspaces, err := uc.workspaceRepo.ListByOrganization(ctx, orgID)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "failed to list workspaces", err)
	}

	result := make([]*dto.WorkspaceInfo, 0, len(workspaces))
	for _, ws := range workspaces {
		result = append(result, &dto.WorkspaceInfo{
			ID:             ws.ID,
			OrganizationID: ws.OrganizationID,
			Name:           ws.Name,
			Slug:           ws.Slug,
			Description:    ws.Description,
			Status:         string(ws.Status),
			CreatedAt:      ws.CreatedAt,
			UpdatedAt:      ws.UpdatedAt,
		})
	}

	return result, nil
}
