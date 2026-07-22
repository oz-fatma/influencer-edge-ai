package model

import (
	"time"

	"github.com/google/uuid"
)

// EndpointPolicy represents the policies applied to an endpoint.
type EndpointPolicy struct {
	ID                 uuid.UUID `json:"id"`
	EndpointID         uuid.UUID `json:"endpoint_id"`
	RequiredPermission string    `json:"required_permission"`
	RateLimit          int       `json:"rate_limit"`
	AuthPolicy         string    `json:"auth_policy"`
	ValidationPolicy   []byte    `json:"validation_policy,omitempty"`
	ExtraPolicies      []byte    `json:"extra_policies,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
