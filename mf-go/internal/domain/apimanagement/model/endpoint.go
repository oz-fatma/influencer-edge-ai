package model

import (
	"time"

	"github.com/google/uuid"
)

// EndpointStatus represents the status of an endpoint.
type EndpointStatus string

const (
	EndpointStatusActive     EndpointStatus = "active"
	EndpointStatusInactive   EndpointStatus = "inactive"
	EndpointStatusDeprecated EndpointStatus = "deprecated"
	EndpointStatusRetired    EndpointStatus = "retired"
)

// Endpoint represents a managed API endpoint within an app.
type Endpoint struct {
	ID             uuid.UUID      `json:"id"`
	AppID          uuid.UUID      `json:"app_id"`
	Method         string         `json:"method"`
	Path           string         `json:"path"`
	Version        string         `json:"version"`
	BackendService string         `json:"backend_service"`
	BackendAction  string         `json:"backend_action"`
	Schema         []byte         `json:"schema,omitempty"`
	AuditLevel     string         `json:"audit_level"`
	PIIMasking     bool           `json:"pii_masking"`
	EventAfter     string         `json:"event_after,omitempty"`
	Status         EndpointStatus `json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// IsActive checks if the endpoint is active.
func (e *Endpoint) IsActive() bool {
	return e.Status == EndpointStatusActive
}
