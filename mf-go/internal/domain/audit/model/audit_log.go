package model

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an immutable audit log entry.
type AuditLog struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	AppID          *uuid.UUID `json:"app_id,omitempty"`
	EndpointID     *uuid.UUID `json:"endpoint_id,omitempty"`
	UserID         *uuid.UUID `json:"user_id,omitempty"`
	RequestID      string     `json:"request_id"`
	Action         string     `json:"action"`
	ResourceType   string     `json:"resource_type"`
	ResourceID     string     `json:"resource_id"`
	Metadata       []byte     `json:"metadata,omitempty"`
	IPAddress      string     `json:"ip_address"`
	UserAgent      string     `json:"user_agent"`
	CreatedAt      time.Time  `json:"created_at"`
}
