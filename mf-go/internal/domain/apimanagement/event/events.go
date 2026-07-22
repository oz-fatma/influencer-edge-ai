package event

import (
	"time"

	"github.com/google/uuid"
)

// EndpointCreated is emitted when a new endpoint is defined.
type EndpointCreated struct {
	EndpointID uuid.UUID `json:"endpoint_id"`
	AppID      uuid.UUID `json:"app_id"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	Timestamp  time.Time `json:"timestamp"`
}

// EndpointUpdated is emitted when an endpoint is modified.
type EndpointUpdated struct {
	EndpointID uuid.UUID `json:"endpoint_id"`
	AppID      uuid.UUID `json:"app_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// EndpointRetired is emitted when an endpoint is retired.
type EndpointRetired struct {
	EndpointID uuid.UUID `json:"endpoint_id"`
	AppID      uuid.UUID `json:"app_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// EndpointActivated is emitted when an endpoint is activated (reactivated from retired/inactive status).
type EndpointActivated struct {
	EndpointID uuid.UUID `json:"endpoint_id"`
	AppID      uuid.UUID `json:"app_id"`
	Timestamp  time.Time `json:"timestamp"`
}
