package models

import (
	"time"

	"gorm.io/gorm"
)

type InfluencerScore struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	UserID          uint           `gorm:"index;not null" json:"user_id"`
	InfluencerName  string         `gorm:"not null;size:255" json:"influencer_name"`
	Platform        string         `gorm:"not null;size:64" json:"platform"`
	OverallScore    float64        `gorm:"not null" json:"overall_score"`
	EngagementScore float64        `json:"engagement_score"`
	AudienceScore   float64        `json:"audience_score"`
	BrandFitScore   float64        `json:"brand_fit_score"`
	RawPayload      string         `gorm:"type:text" json:"raw_payload,omitempty"`
	Notes           string         `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type InfluencerAnalysis struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UserID         uint           `gorm:"index;not null" json:"user_id"`
	InfluencerName string         `gorm:"not null;size:255" json:"influencer_name"`
	Platform       string         `gorm:"not null;size:64" json:"platform"`
	AnalysisType   string         `gorm:"not null;size:64" json:"analysis_type"`
	Summary        string         `gorm:"type:text" json:"summary"`
	Insights       string         `gorm:"type:text" json:"insights"`
	RawLLMOutput   string         `gorm:"type:text" json:"raw_llm_output,omitempty"`
	ScoreID        *uint          `gorm:"index" json:"score_id,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
