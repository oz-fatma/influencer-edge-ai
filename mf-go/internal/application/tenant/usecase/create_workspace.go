package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/tenant/dto"
	tenantEvent "github.com/masterfabric-go/masterfabric/internal/domain/tenant/event"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/repository"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
	"github.com/masterfabric-go/masterfabric/internal/shared/middleware"
)

// CreateWorkspaceUseCase handles workspace creation.
type CreateWorkspaceUseCase struct {
	workspaceRepo repository.WorkspaceRepository
	orgRepo       repository.OrgRepository
	eventBus      events.EventBus
}

// NewCreateWorkspaceUseCase creates a new CreateWorkspaceUseCase.
func NewCreateWorkspaceUseCase(
	workspaceRepo repository.WorkspaceRepository,
	orgRepo repository.OrgRepository,
	eventBus events.EventBus,
) *CreateWorkspaceUseCase {
	return &CreateWorkspaceUseCase{
		workspaceRepo: workspaceRepo,
		orgRepo:       orgRepo,
		eventBus:      eventBus,
	}
}

// Execute creates a new workspace.
func (uc *CreateWorkspaceUseCase) Execute(ctx context.Context, req dto.CreateWorkspaceRequest) (*dto.WorkspaceInfo, error) {
	// Get organization from context
	orgID, ok := middleware.OrgIDFromContext(ctx)
	if !ok {
		return nil, domainErr.New(domainErr.ErrUnauthorized, "organization context required", nil)
	}

	// Verify organization exists
	org, err := uc.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrNotFound, "organization not found", err)
	}

	if !org.IsActive() {
		return nil, domainErr.New(domainErr.ErrForbidden, "organization is not active", nil)
	}

	// Normalize slug
	slug := strings.ToLower(strings.TrimSpace(req.Slug))

	// Check if slug is taken within the organization
	existing, _ := uc.workspaceRepo.GetBySlug(ctx, orgID, slug)
	if existing != nil {
		return nil, domainErr.New(domainErr.ErrAlreadyExists, "workspace slug already taken in this organization", nil)
	}

	workspace := &model.Workspace{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           strings.TrimSpace(req.Name),
		Slug:           slug,
		Description:    strings.TrimSpace(req.Description),
		Status:         model.WorkspaceStatusActive,
	}

	if err := uc.workspaceRepo.Create(ctx, workspace); err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "failed to create workspace", err)
	}

	// Publish domain event
	createdBy, _ := middleware.UserIDFromContext(ctx)
	_ = uc.eventBus.Publish(ctx, events.TopicTenant, tenantEvent.WorkspaceCreated{
		WorkspaceID:    workspace.ID,
		OrganizationID: orgID,
		Name:           workspace.Name,
		CreatedBy:      createdBy,
		Timestamp:      time.Now().UTC(),
	})

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
