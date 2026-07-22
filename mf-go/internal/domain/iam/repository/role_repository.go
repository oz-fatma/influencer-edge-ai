package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/iam/model"
)

// RoleRepository defines the interface for role persistence.
type RoleRepository interface {
	Create(ctx context.Context, role *model.Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Role, error)
	ListByScope(ctx context.Context, scopeType model.ScopeType, scopeID uuid.UUID) ([]*model.Role, error)
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Permissions
	AddPermission(ctx context.Context, roleID uuid.UUID, permission string) error
	RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error
	GetPermissions(ctx context.Context, roleID uuid.UUID) ([]string, error)

	// User-role assignments
	AssignRoleToUser(ctx context.Context, userRole *model.UserRole) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID, orgID uuid.UUID) ([]*model.UserRole, error)
	GetUserPermissions(ctx context.Context, userID, orgID uuid.UUID) ([]string, error)
}
