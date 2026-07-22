package model

import (
	"time"

	"github.com/google/uuid"
)

// OrgStatus represents the status of an organization.
type OrgStatus string

const (
	OrgStatusActive    OrgStatus = "active"
	OrgStatusSuspended OrgStatus = "suspended"
	OrgStatusArchived  OrgStatus = "archived"
)

// Organization represents a tenant entity.
type Organization struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	Slug                string    `json:"slug"`
	Status              OrgStatus `json:"status"`
	SSOConfig           []byte    `json:"sso_config,omitempty"`
	DataRetentionPolicy []byte    `json:"data_retention_policy,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// IsActive checks if the organization is active.
func (o *Organization) IsActive() bool {
	return o.Status == OrgStatusActive
}
