package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DefineEndpointRequest is the input for defining a new endpoint.
type DefineEndpointRequest struct {
	Method         string `json:"method" validate:"required,oneof=GET POST PUT PATCH DELETE"`
	Path           string `json:"path" validate:"required"`
	Version        string `json:"version"`
	BackendService string `json:"backend_service" validate:"required"`
	BackendAction  string `json:"backend_action" validate:"required"`
	Schema         json.RawMessage `json:"schema,omitempty"`
	AuditLevel     string `json:"audit_level"`
	PIIMasking     bool   `json:"pii_masking"`
	EventAfter     string `json:"event_after,omitempty"`
}

// EndpointInfo is a public endpoint representation.
type EndpointInfo struct {
	ID             uuid.UUID       `json:"id"`
	AppID          uuid.UUID       `json:"app_id"`
	Method         string          `json:"method"`
	Path           string          `json:"path"`
	Version        string          `json:"version"`
	BackendService string          `json:"backend_service"`
	BackendAction  string          `json:"backend_action"`
	Schema         json.RawMessage `json:"schema,omitempty"`
	AuditLevel     string          `json:"audit_level"`
	PIIMasking     bool            `json:"pii_masking"`
	EventAfter     string          `json:"event_after,omitempty"`
	Status         string          `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
}

// UpdatePolicyRequest is the input for updating an endpoint policy.
type UpdatePolicyRequest struct {
	RequiredPermission string          `json:"required_permission"`
	RateLimit          int             `json:"rate_limit"`
	AuthPolicy         string          `json:"auth_policy"`
	ValidationPolicy   json.RawMessage `json:"validation_policy,omitempty"`
	ExtraPolicies      json.RawMessage `json:"extra_policies,omitempty"`
}

// PolicyInfo is a public endpoint policy representation.
type PolicyInfo struct {
	ID                 uuid.UUID       `json:"id"`
	EndpointID         uuid.UUID       `json:"endpoint_id"`
	RequiredPermission string          `json:"required_permission"`
	RateLimit          int             `json:"rate_limit"`
	AuthPolicy         string          `json:"auth_policy"`
	ValidationPolicy   json.RawMessage `json:"validation_policy,omitempty"`
	ExtraPolicies      json.RawMessage `json:"extra_policies,omitempty"`
}
