package event

import (
	"time"

	"github.com/google/uuid"
)

// UserRegistered is emitted when a new user registers.
type UserRegistered struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Timestamp time.Time `json:"timestamp"`
}

// UserInvited is emitted when a user is invited to an organization.
type UserInvited struct {
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	InvitedBy      uuid.UUID `json:"invited_by"`
	Timestamp      time.Time `json:"timestamp"`
}

// RoleAssigned is emitted when a role is assigned to a user.
type RoleAssigned struct {
	UserID         uuid.UUID `json:"user_id"`
	RoleID         uuid.UUID `json:"role_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Timestamp      time.Time `json:"timestamp"`
}

// RoleRevoked is emitted when a role is removed from a user.
type RoleRevoked struct {
	UserID         uuid.UUID `json:"user_id"`
	RoleID         uuid.UUID `json:"role_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Timestamp      time.Time `json:"timestamp"`
}
