package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/tenant/dto"
	tenantEvent "github.com/masterfabric-go/masterfabric/internal/domain/tenant/event"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/repository"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
)

// CreateAppUseCase handles app creation within an organization.
type CreateAppUseCase struct {
	appRepo  repository.AppRepository
	orgRepo  repository.OrgRepository
	eventBus events.EventBus
}

// NewCreateAppUseCase creates a new CreateAppUseCase.
func NewCreateAppUseCase(appRepo repository.AppRepository, orgRepo repository.OrgRepository, eventBus events.EventBus) *CreateAppUseCase {
	return &CreateAppUseCase{appRepo: appRepo, orgRepo: orgRepo, eventBus: eventBus}
}

// Execute creates a new app within an organization.
func (uc *CreateAppUseCase) Execute(ctx context.Context, orgID uuid.UUID, req dto.CreateAppRequest) (*dto.AppInfo, error) {
	// Verify organization exists
	org, err := uc.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if !org.IsActive() {
		return nil, domainErr.New(domainErr.ErrForbidden, "organization is not active", nil)
	}

	// Check slug uniqueness within org
	existing, _ := uc.appRepo.GetBySlug(ctx, orgID, req.Slug)
	if existing != nil {
		return nil, domainErr.New(domainErr.ErrAlreadyExists, "app slug already taken in this organization", nil)
	}

	slaTier := req.SLATier
	if slaTier == "" {
		slaTier = "standard"
	}

	app := &model.App{
		OrganizationID: orgID,
		Name:           req.Name,
		Slug:           req.Slug,
		Status:         model.AppStatusActive,
		SLATier:        slaTier,
	}

	if err := uc.appRepo.Create(ctx, app); err != nil {
		return nil, err
	}

	// Publish domain event to Kafka
	_ = uc.eventBus.Publish(ctx, events.TopicTenant, tenantEvent.AppCreated{
		AppID:          app.ID,
		OrganizationID: orgID,
		Name:           app.Name,
		Timestamp:      time.Now().UTC(),
	})

	return &dto.AppInfo{
		ID:             app.ID,
		OrganizationID: app.OrganizationID,
		Name:           app.Name,
		Slug:           app.Slug,
		Status:         string(app.Status),
		SLATier:        app.SLATier,
		CreatedAt:      app.CreatedAt,
	}, nil
}
