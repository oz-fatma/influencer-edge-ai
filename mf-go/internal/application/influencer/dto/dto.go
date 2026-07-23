package dto

import (
	"strings"
	"time"

	"github.com/google/uuid"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
)

const (
	MaxNotesLen        = 4096
	MaxRawPayloadLen   = 65536
	MaxRawLLMOutputLen = 65536
	DefaultListLimit   = 50
	MaxListLimit       = 200
)

var allowedPlatforms = map[string]struct{}{
	"instagram": {}, "tiktok": {}, "youtube": {}, "twitter": {},
	"linkedin": {}, "other": {},
}

type CreateScoreRequest struct {
	InfluencerName  string   `json:"influencer_name" validate:"required"`
	Platform        string   `json:"platform" validate:"required"`
	OverallScore    *float64 `json:"overall_score,omitempty"`
	EngagementScore *float64 `json:"engagement_score,omitempty"`
	AudienceScore   *float64 `json:"audience_score,omitempty"`
	BrandFitScore   *float64 `json:"brand_fit_score,omitempty"`
	RawPayload      string   `json:"raw_payload"`
	Notes           string   `json:"notes"`
}

type UpdateScoreRequest struct {
	InfluencerName  *string  `json:"influencer_name"`
	Platform        *string  `json:"platform"`
	OverallScore    *float64 `json:"overall_score"`
	EngagementScore *float64 `json:"engagement_score"`
	AudienceScore   *float64 `json:"audience_score"`
	BrandFitScore   *float64 `json:"brand_fit_score"`
	RawPayload      *string  `json:"raw_payload"`
	Notes           *string  `json:"notes"`
}

type CreateAnalysisRequest struct {
	InfluencerName string     `json:"influencer_name" validate:"required"`
	Platform       string     `json:"platform" validate:"required"`
	AnalysisType   string     `json:"analysis_type" validate:"required"`
	Summary        string     `json:"summary" validate:"required"`
	Insights       string     `json:"insights"`
	RawLLMOutput   string     `json:"raw_llm_output"`
	ScoreID        *uuid.UUID `json:"score_id"`
}

type AnalyzeInfluencerRequest struct {
	InfluencerName string `json:"influencer_name" validate:"required"`
	Platform       string `json:"platform" validate:"required"`
	Notes          string `json:"notes"`
}

type AnalyzeInfluencerResponse struct {
	Result    AnalyzeInfluencerResult `json:"result"`
	RawOutput string                  `json:"raw_output"`
}

type AnalyzeInfluencerResult struct {
	OverallScore    float64  `json:"overall_score"`
	EngagementScore float64  `json:"engagement_score"`
	AudienceScore   float64  `json:"audience_score"`
	BrandFitScore   float64  `json:"brand_fit_score"`
	Summary         string   `json:"summary"`
	Insights        []string `json:"insights"`
}

type RecordLLMMetricRequest struct {
	InfluencerName string `json:"influencer_name" validate:"required"`
	LatencyMs      int64  `json:"latency_ms" validate:"required"`
	Status         string `json:"status" validate:"required"`
	Model          string `json:"model" validate:"required"`
}

type ScoreResponse struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	InfluencerName  string    `json:"influencer_name"`
	Platform        string    `json:"platform"`
	OverallScore    float64   `json:"overall_score"`
	EngagementScore float64   `json:"engagement_score"`
	AudienceScore   float64   `json:"audience_score"`
	BrandFitScore   float64   `json:"brand_fit_score"`
	RawPayload      string    `json:"raw_payload,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AnalysisResponse struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	InfluencerName string     `json:"influencer_name"`
	Platform       string     `json:"platform"`
	AnalysisType   string     `json:"analysis_type"`
	Summary        string     `json:"summary"`
	Insights       string     `json:"insights,omitempty"`
	RawLLMOutput   string     `json:"raw_llm_output,omitempty"`
	ScoreID        *uuid.UUID `json:"score_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type LLMCallMetric struct {
	ID             string `json:"id"`
	InfluencerName string `json:"influencer_name"`
	LatencyMs      int64  `json:"latency_ms"`
	Status         string `json:"status"`
	Model          string `json:"model"`
	Timestamp      int64  `json:"timestamp"`
}

type MonitoringStatsResponse struct {
	TotalCalls   int64           `json:"total_calls"`
	AvgLatencyMs float64         `json:"avg_latency_ms"`
	ErrorRate    float64         `json:"error_rate"`
	RecentCalls  []LLMCallMetric `json:"recent_calls"`
}

func NormalizePlatform(platform string) string {
	return strings.ToLower(strings.TrimSpace(platform))
}

func ValidateInfluencerName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return domainErr.New(domainErr.ErrValidation, "influencer_name is required", nil)
	}
	if len(name) > 255 {
		return domainErr.New(domainErr.ErrValidation, "influencer_name must be at most 255 characters", nil)
	}
	return nil
}

func ValidatePlatform(platform string) error {
	p := NormalizePlatform(platform)
	if p == "" {
		return domainErr.New(domainErr.ErrValidation, "platform is required", nil)
	}
	if _, ok := allowedPlatforms[p]; !ok {
		return domainErr.New(domainErr.ErrValidation, "platform must be one of: instagram, tiktok, youtube, twitter, linkedin, other", nil)
	}
	return nil
}

func ValidateScoreValue(score float64) error {
	if score < 0 || score > 100 {
		return domainErr.New(domainErr.ErrValidation, "score must be between 0 and 100", nil)
	}
	return nil
}

func ScoreValueOrDefault(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func ValidateOptionalScoreValue(v *float64) error {
	if v == nil {
		return nil
	}
	return ValidateScoreValue(*v)
}

func ClampListLimit(raw int) int {
	if raw <= 0 {
		return DefaultListLimit
	}
	if raw > MaxListLimit {
		return MaxListLimit
	}
	return raw
}

func ClampListOffset(raw int) int {
	if raw < 0 {
		return 0
	}
	return raw
}
