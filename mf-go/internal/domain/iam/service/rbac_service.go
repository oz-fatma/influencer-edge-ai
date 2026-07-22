package service

import (
	"context"

	"github.com/google/uuid"
)

// RBACService defines authorization checking operations.
type RBACService interface {
	// HasPermission checks if a user has a specific permission within an organization.
	HasPermission(ctx context.Context, userID, orgID uuid.UUID, permission string) (bool, error)
	// HasAnyPermission checks if a user has at least one of the given permissions.
	HasAnyPermission(ctx context.Context, userID, orgID uuid.UUID, permissions []string) (bool, error)
	// GetUserPermissions returns all permissions for a user within an organization.
	GetUserPermissions(ctx context.Context, userID, orgID uuid.UUID) ([]string, error)
	// InvalidateCache invalidates the cached permissions for a user.
	InvalidateCache(ctx context.Context, userID, orgID uuid.UUID) error
}
