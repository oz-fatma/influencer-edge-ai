package usecase

import (
	"context"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/application/iam/dto"
	iamEvent "github.com/masterfabric-go/masterfabric/internal/domain/iam/event"
	"github.com/masterfabric-go/masterfabric/internal/domain/iam/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/iam/repository"
	"github.com/masterfabric-go/masterfabric/internal/domain/iam/service"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
)

// AssignRoleUseCase handles role assignment.
type AssignRoleUseCase struct {
	roleRepo repository.RoleRepository
	rbac     service.RBACService
	eventBus events.EventBus
}

// NewAssignRoleUseCase creates a new AssignRoleUseCase.
func NewAssignRoleUseCase(roleRepo repository.RoleRepository, rbac service.RBACService, eventBus events.EventBus) *AssignRoleUseCase {
	return &AssignRoleUseCase{roleRepo: roleRepo, rbac: rbac, eventBus: eventBus}
}

// Execute assigns a role to a user.
func (uc *AssignRoleUseCase) Execute(ctx context.Context, req dto.AssignRoleRequest) error {
	userRole := &model.UserRole{
		UserID:         req.UserID,
		RoleID:         req.RoleID,
		OrganizationID: req.OrganizationID,
		AppID:          req.AppID,
	}

	if err := uc.roleRepo.AssignRoleToUser(ctx, userRole); err != nil {
		return err
	}

	// Invalidate permission cache
	_ = uc.rbac.InvalidateCache(ctx, req.UserID, req.OrganizationID)

	// Publish domain event to Kafka
	_ = uc.eventBus.Publish(ctx, events.TopicIAM, iamEvent.RoleAssigned{
		UserID:         req.UserID,
		RoleID:         req.RoleID,
		OrganizationID: req.OrganizationID,
		Timestamp:      time.Now().UTC(),
	})

	return nil
}
