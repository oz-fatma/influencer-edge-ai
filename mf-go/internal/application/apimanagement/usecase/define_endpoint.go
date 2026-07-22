package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/apimanagement/dto"
	apimgmtEvent "github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/event"
	"github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/repository"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
)

// DefineEndpointUseCase handles endpoint creation.
type DefineEndpointUseCase struct {
	endpointRepo repository.EndpointRepository
	eventBus     events.EventBus
}

// NewDefineEndpointUseCase creates a new DefineEndpointUseCase.
func NewDefineEndpointUseCase(endpointRepo repository.EndpointRepository, eventBus events.EventBus) *DefineEndpointUseCase {
	return &DefineEndpointUseCase{endpointRepo: endpointRepo, eventBus: eventBus}
}

// Execute defines a new endpoint for an app.
func (uc *DefineEndpointUseCase) Execute(ctx context.Context, appID uuid.UUID, req dto.DefineEndpointRequest) (*dto.EndpointInfo, error) {
	version := req.Version
	if version == "" {
		version = "v1"
	}

	// Check for duplicates
	existing, _ := uc.endpointRepo.GetByMethodPath(ctx, appID, req.Method, req.Path, version)
	if existing != nil {
		return nil, domainErr.New(domainErr.ErrAlreadyExists, "endpoint already exists for this method/path/version", nil)
	}

	auditLevel := req.AuditLevel
	if auditLevel == "" {
		auditLevel = "standard"
	}

	endpoint := &model.Endpoint{
		AppID:          appID,
		Method:         req.Method,
		Path:           req.Path,
		Version:        version,
		BackendService: req.BackendService,
		BackendAction:  req.BackendAction,
		Schema:         req.Schema,
		AuditLevel:     auditLevel,
		PIIMasking:     req.PIIMasking,
		EventAfter:     req.EventAfter,
		Status:         model.EndpointStatusActive,
	}

	if err := uc.endpointRepo.Create(ctx, endpoint); err != nil {
		return nil, err
	}

	// Publish domain event to Kafka
	_ = uc.eventBus.Publish(ctx, events.TopicAPIManagement, apimgmtEvent.EndpointCreated{
		EndpointID: endpoint.ID,
		AppID:      appID,
		Method:     endpoint.Method,
		Path:       endpoint.Path,
		Timestamp:  time.Now().UTC(),
	})

	return &dto.EndpointInfo{
		ID:             endpoint.ID,
		AppID:          endpoint.AppID,
		Method:         endpoint.Method,
		Path:           endpoint.Path,
		Version:        endpoint.Version,
		BackendService: endpoint.BackendService,
		BackendAction:  endpoint.BackendAction,
		Schema:         endpoint.Schema,
		AuditLevel:     endpoint.AuditLevel,
		PIIMasking:     endpoint.PIIMasking,
		EventAfter:     endpoint.EventAfter,
		Status:         string(endpoint.Status),
		CreatedAt:      endpoint.CreatedAt,
	}, nil
}
