package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	iamService "github.com/masterfabric-go/masterfabric/internal/domain/iam/service"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/repository"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
)

// ConnectInput holds validated connection parameters.
type ConnectInput struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	AppID          uuid.UUID
}

// ValidateConnectUseCase validates WebSocket connection prerequisites.
type ValidateConnectUseCase struct {
	appRepo     repository.AppRepository
	rbacService iamService.RBACService
}

// NewValidateConnectUseCase creates a new ValidateConnectUseCase.
func NewValidateConnectUseCase(appRepo repository.AppRepository, rbac iamService.RBACService) *ValidateConnectUseCase {
	return &ValidateConnectUseCase{appRepo: appRepo, rbacService: rbac}
}

// Execute verifies org/app ownership and RBAC before upgrading the connection.
func (uc *ValidateConnectUseCase) Execute(ctx context.Context, userID, orgID, appID uuid.UUID) (*ConnectInput, error) {
	if userID == uuid.Nil {
		return nil, domainErr.New(domainErr.ErrUnauthorized, "user not authenticated", nil)
	}
	if orgID == uuid.Nil {
		return nil, domainErr.New(domainErr.ErrBadRequest, "organization context required", nil)
	}
	if appID == uuid.Nil {
		return nil, domainErr.New(domainErr.ErrBadRequest, "app context required", nil)
	}

	app, err := uc.appRepo.GetByID(ctx, appID)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrNotFound, "app not found", err)
	}
	if app.OrganizationID != orgID {
		return nil, domainErr.New(domainErr.ErrForbidden, "app does not belong to organization", nil)
	}
	if !app.IsActive() {
		return nil, domainErr.New(domainErr.ErrForbidden, "app is not active", nil)
	}

	has, err := uc.rbacService.HasPermission(ctx, userID, orgID, "app:read")
	if err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "permission check failed", err)
	}
	if !has {
		return nil, domainErr.New(domainErr.ErrForbidden, "insufficient permissions to connect", nil)
	}

	return &ConnectInput{
		UserID:         userID,
		OrganizationID: orgID,
		AppID:          appID,
	}, nil
}

// ParseAppHeader parses the X-App-ID header value.
func ParseAppHeader(appIDStr string) (uuid.UUID, error) {
	if appIDStr == "" {
		return uuid.Nil, domainErr.New(domainErr.ErrBadRequest, "X-App-ID header required", nil)
	}
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		return uuid.Nil, domainErr.New(domainErr.ErrBadRequest, "invalid X-App-ID", fmt.Errorf("parse app id: %w", err))
	}
	return appID, nil
}
