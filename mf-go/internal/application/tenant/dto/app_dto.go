package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateAppRequest is the input for creating an app.
type CreateAppRequest struct {
	Name    string `json:"name" validate:"required,min=2"`
	Slug    string `json:"slug" validate:"required,min=2,alphanum"`
	SLATier string `json:"sla_tier"`
}

// AppInfo is a public app representation.
type AppInfo struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Status         string    `json:"status"`
	SLATier        string    `json:"sla_tier"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateAPIKeyRequest is the input for creating an API key.
type CreateAPIKeyRequest struct {
	Name string `json:"name" validate:"required"`
}

// APIKeyResponse is the output for a created API key (includes the raw key only once).
type APIKeyResponse struct {
	ID        uuid.UUID  `json:"id"`
	AppID     uuid.UUID  `json:"app_id"`
	Name      string     `json:"name"`
	Key       string     `json:"key,omitempty"` // Only returned on creation
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
}
