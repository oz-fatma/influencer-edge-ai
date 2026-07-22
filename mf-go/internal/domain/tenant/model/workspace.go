package model

import (
	"time"

	"github.com/google/uuid"
)

// WorkspaceStatus represents the status of a workspace.
type WorkspaceStatus string

const (
	WorkspaceStatusActive    WorkspaceStatus = "active"
	WorkspaceStatusSuspended WorkspaceStatus = "suspended"
	WorkspaceStatusArchived  WorkspaceStatus = "archived"
)

// Workspace represents a workspace within an organization.
// Workspaces provide an additional level of tenant isolation and organization.
// Hierarchy: Organization -> Workspace -> App
type Workspace struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	Name           string          `json:"name"`
	Slug           string          `json:"slug"`
	Description    string          `json:"description,omitempty"`
	Status         WorkspaceStatus `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// IsActive checks if the workspace is active.
func (w *Workspace) IsActive() bool {
	return w.Status == WorkspaceStatusActive
}
