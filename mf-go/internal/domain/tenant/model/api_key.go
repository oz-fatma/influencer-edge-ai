package model

import (
	"time"

	"github.com/google/uuid"
)

// AppAPIKey represents an API key for an app.
type AppAPIKey struct {
	ID        uuid.UUID  `json:"id"`
	AppID     uuid.UUID  `json:"app_id"`
	KeyHash   string     `json:"-"`
	Name      string     `json:"name"`
	Scopes    []byte     `json:"scopes,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
}
