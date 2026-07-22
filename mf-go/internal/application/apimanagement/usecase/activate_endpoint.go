package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	apimgmtEvent "github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/event"
	"github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/repository"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
)

// ActivateEndpointUseCase handles activating a retired or inactive endpoint.
type ActivateEndpointUseCase struct {
	endpointRepo repository.EndpointRepository
	eventBus     events.EventBus
}

// NewActivateEndpointUseCase creates a new ActivateEndpointUseCase.
func NewActivateEndpointUseCase(endpointRepo repository.EndpointRepository, eventBus events.EventBus) *ActivateEndpointUseCase {
	return &ActivateEndpointUseCase{endpointRepo: endpointRepo, eventBus: eventBus}
}

// Execute activates an endpoint.
func (uc *ActivateEndpointUseCase) Execute(ctx context.Context, endpointID uuid.UUID) error {
	endpoint, err := uc.endpointRepo.GetByID(ctx, endpointID)
	if err != nil {
		return err
	}

	if endpoint.Status == model.EndpointStatusActive {
		return domainErr.New(domainErr.ErrBadRequest, "endpoint is already active", nil)
	}

	endpoint.Status = model.EndpointStatusActive
	if err := uc.endpointRepo.Update(ctx, endpoint); err != nil {
		return err
	}

	// Publish domain event to Kafka
	_ = uc.eventBus.Publish(ctx, events.TopicAPIManagement, apimgmtEvent.EndpointActivated{
		EndpointID: endpointID,
		AppID:      endpoint.AppID,
		Timestamp:  time.Now().UTC(),
	})

	return nil
}
