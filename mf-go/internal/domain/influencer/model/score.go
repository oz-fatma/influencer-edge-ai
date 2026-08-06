package model

import (
	"time"

	"github.com/google/uuid"
)

type InfluencerScore struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	InfluencerName  string
	Platform        string
	OverallScore    float64
	EngagementScore float64
	AudienceScore   float64
	BrandFitScore   float64
	RawPayload      string
	Notes           string
	Niche           string
	AudienceGeo     string
	AudienceDemo    string
	FollowerRange   string
	EngagementRate  *float64
	ContentFormats  []string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type InfluencerAnalysis struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	InfluencerName string
	Platform       string
	AnalysisType   string
	Summary        string
	Insights       string
	RawLLMOutput   string
	ScoreID        *uuid.UUID
	BrandProfileID *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
