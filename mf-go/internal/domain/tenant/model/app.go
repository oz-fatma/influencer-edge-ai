package model

import (
	"time"

	"github.com/google/uuid"
)

// AppStatus represents the status of an app.
type AppStatus string

const (
	AppStatusActive   AppStatus = "active"
	AppStatusInactive AppStatus = "inactive"
	AppStatusArchived AppStatus = "archived"
)

// App represents a logical application within an organization.
type App struct {
	ID              uuid.UUID `json:"id"`
	OrganizationID  uuid.UUID `json:"organization_id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Status          AppStatus `json:"status"`
	SLATier         string    `json:"sla_tier"`
	RateLimitPolicy []byte    `json:"rate_limit_policy,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// IsActive checks if the app is active.
func (a *App) IsActive() bool {
	return a.Status == AppStatusActive
}
