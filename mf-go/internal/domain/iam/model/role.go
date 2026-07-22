package model

import (
	"time"

	"github.com/google/uuid"
)

// ScopeType represents the scope a role applies to.
type ScopeType string

const (
	ScopeTypeOrganization ScopeType = "organization"
	ScopeTypeApp          ScopeType = "app"
)

// Role represents an RBAC role.
type Role struct {
	ID          uuid.UUID `json:"id"`
	ScopeType   ScopeType `json:"scope_type"`
	ScopeID     uuid.UUID `json:"scope_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a named permission string.
type Permission struct {
	RoleID     uuid.UUID `json:"role_id"`
	Permission string    `json:"permission"`
	CreatedAt  time.Time `json:"created_at"`
}

// UserRole represents the assignment of a role to a user within a scope.
type UserRole struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	RoleID         uuid.UUID  `json:"role_id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	AppID          *uuid.UUID `json:"app_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}
