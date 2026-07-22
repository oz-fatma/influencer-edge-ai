package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateWorkspaceRequest is the input for creating a workspace.
type CreateWorkspaceRequest struct {
	Name        string `json:"name" validate:"required,min=2"`
	Slug        string `json:"slug" validate:"required,min=2,alphanum"`
	Description string `json:"description,omitempty"`
}

// WorkspaceInfo is a public workspace representation.
type WorkspaceInfo struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// UpdateWorkspaceRequest is the input for updating a workspace.
type UpdateWorkspaceRequest struct {
	Name        string `json:"name,omitempty"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}
