package model

import (
	"time"

	"github.com/google/uuid"
)

type BrandProfile struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Name           string
	Industry       string
	TargetAudience string
	BudgetRange    string
	BrandValues    string
	CampaignGoal   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
