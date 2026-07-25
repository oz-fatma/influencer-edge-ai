package dto

import (
	"time"

	"github.com/google/uuid"
)

// RegisterRequest is the input for user registration.
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

// LoginRequest is the input for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse is the output for successful login.
type LoginResponse struct {
	Token   string   `json:"token"`
	User    UserInfo `json:"user"`
	IsAdmin bool     `json:"is_admin"`
}

// UserInfo is a public user representation.
type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// MeResponse is the output for GET /auth/me.
type MeResponse struct {
	User    UserInfo `json:"user"`
	IsAdmin bool     `json:"is_admin"`
}
type AssignRoleRequest struct {
	UserID         uuid.UUID  `json:"user_id" validate:"required"`
	RoleID         uuid.UUID  `json:"role_id" validate:"required"`
	OrganizationID uuid.UUID  `json:"organization_id" validate:"required"`
	AppID          *uuid.UUID `json:"app_id,omitempty"`
}

// UpdateUserRequest is the input for updating a user.
type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Status    string `json:"status"`
}
