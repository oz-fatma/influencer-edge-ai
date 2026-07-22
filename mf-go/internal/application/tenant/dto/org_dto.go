package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateOrgRequest is the input for creating an organization.
type CreateOrgRequest struct {
	Name string `json:"name" validate:"required,min=2"`
	Slug string `json:"slug" validate:"required,min=2,alphanum"`
}

// OrgInfo is a public organization representation.
type OrgInfo struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// UpdateOrgRequest is the input for updating an organization.
type UpdateOrgRequest struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
