package usecase

import (
	"context"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/application/tenant/dto"
	tenantEvent "github.com/masterfabric-go/masterfabric/internal/domain/tenant/event"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/repository"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
	"github.com/masterfabric-go/masterfabric/internal/shared/middleware"
)

// CreateOrgUseCase handles organization creation.
type CreateOrgUseCase struct {
	orgRepo  repository.OrgRepository
	eventBus events.EventBus
}

// NewCreateOrgUseCase creates a new CreateOrgUseCase.
func NewCreateOrgUseCase(orgRepo repository.OrgRepository, eventBus events.EventBus) *CreateOrgUseCase {
	return &CreateOrgUseCase{orgRepo: orgRepo, eventBus: eventBus}
}

// Execute creates a new organization.
func (uc *CreateOrgUseCase) Execute(ctx context.Context, req dto.CreateOrgRequest) (*dto.OrgInfo, error) {
	// Check if slug is taken
	existing, _ := uc.orgRepo.GetBySlug(ctx, req.Slug)
	if existing != nil {
		return nil, domainErr.New(domainErr.ErrAlreadyExists, "organization slug already taken", nil)
	}

	org := &model.Organization{
		Name:   req.Name,
		Slug:   req.Slug,
		Status: model.OrgStatusActive,
	}

	if err := uc.orgRepo.Create(ctx, org); err != nil {
		return nil, err
	}

	// Publish domain event to Kafka
	createdBy, _ := middleware.UserIDFromContext(ctx)
	_ = uc.eventBus.Publish(ctx, events.TopicTenant, tenantEvent.OrganizationCreated{
		OrganizationID: org.ID,
		Name:           org.Name,
		CreatedBy:      createdBy,
		Timestamp:      time.Now().UTC(),
	})

	return &dto.OrgInfo{
		ID:        org.ID,
		Name:      org.Name,
		Slug:      org.Slug,
		Status:    string(org.Status),
		CreatedAt: org.CreatedAt,
	}, nil
}
