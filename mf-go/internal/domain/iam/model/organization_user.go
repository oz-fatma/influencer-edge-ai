package model

import (
	"time"

	"github.com/google/uuid"
)

// OrgUserStatus represents the membership status.
type OrgUserStatus string

const (
	OrgUserStatusActive  OrgUserStatus = "active"
	OrgUserStatusInvited OrgUserStatus = "invited"
	OrgUserStatusRemoved OrgUserStatus = "removed"
)

// OrganizationUser represents a user's membership in an organization.
type OrganizationUser struct {
	OrganizationID uuid.UUID     `json:"organization_id"`
	UserID         uuid.UUID     `json:"user_id"`
	Status         OrgUserStatus `json:"status"`
	InvitedBy      *uuid.UUID    `json:"invited_by,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}
