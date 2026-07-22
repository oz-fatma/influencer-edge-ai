package event

import (
	"time"

	"github.com/google/uuid"
)

// OrganizationCreated is emitted when a new organization is created.
type OrganizationCreated struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	CreatedBy      uuid.UUID `json:"created_by"`
	Timestamp      time.Time `json:"timestamp"`
}

// AppCreated is emitted when a new app is created.
type AppCreated struct {
	AppID          uuid.UUID `json:"app_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	Timestamp      time.Time `json:"timestamp"`
}

// AppUpdated is emitted when an app is updated.
type AppUpdated struct {
	AppID          uuid.UUID `json:"app_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Timestamp      time.Time `json:"timestamp"`
}

// WorkspaceCreated is emitted when a new workspace is created.
type WorkspaceCreated struct {
	WorkspaceID    uuid.UUID `json:"workspace_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	CreatedBy      uuid.UUID `json:"created_by"`
	Timestamp      time.Time `json:"timestamp"`
}
