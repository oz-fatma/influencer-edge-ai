package dto

import (
	"strings"
	"time"

	"github.com/google/uuid"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
)

const (
	MaxBrandProfileNameLen           = 255
	MaxBrandProfileIndustryLen       = 128
	MaxBrandProfileTargetAudienceLen = 4096
	MaxBrandProfileBudgetRangeLen    = 64
	MaxBrandProfileBrandValuesLen    = 4096
	MaxBrandProfileCampaignGoalLen   = 128
)

type CreateBrandProfileRequest struct {
	Name           string `json:"name" validate:"required"`
	Industry       string `json:"industry" validate:"required"`
	TargetAudience string `json:"target_audience" validate:"required"`
	BudgetRange    string `json:"budget_range"`
	BrandValues    string `json:"brand_values" validate:"required"`
	CampaignGoal   string `json:"campaign_goal" validate:"required"`
}

type UpdateBrandProfileRequest struct {
	Name           *string `json:"name"`
	Industry       *string `json:"industry"`
	TargetAudience *string `json:"target_audience"`
	BudgetRange    *string `json:"budget_range"`
	BrandValues    *string `json:"brand_values"`
	CampaignGoal   *string `json:"campaign_goal"`
}

type BrandProfileResponse struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Name           string    `json:"name"`
	Industry       string    `json:"industry"`
	TargetAudience string    `json:"target_audience"`
	BudgetRange    string    `json:"budget_range,omitempty"`
	BrandValues    string    `json:"brand_values"`
	CampaignGoal   string    `json:"campaign_goal"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func ValidateBrandProfileName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return domainErr.New(domainErr.ErrValidation, "name is required", nil)
	}
	if len(name) > MaxBrandProfileNameLen {
		return domainErr.New(domainErr.ErrValidation, "name must be at most 255 characters", nil)
	}
	return nil
}

func ValidateCreateBrandProfile(req CreateBrandProfileRequest) error {
	if err := ValidateBrandProfileName(req.Name); err != nil {
		return err
	}
	if strings.TrimSpace(req.Industry) == "" {
		return domainErr.New(domainErr.ErrValidation, "industry is required", nil)
	}
	if len(strings.TrimSpace(req.Industry)) > MaxBrandProfileIndustryLen {
		return domainErr.New(domainErr.ErrValidation, "industry must be at most 128 characters", nil)
	}
	if strings.TrimSpace(req.TargetAudience) == "" {
		return domainErr.New(domainErr.ErrValidation, "target_audience is required", nil)
	}
	if len(req.TargetAudience) > MaxBrandProfileTargetAudienceLen {
		return domainErr.New(domainErr.ErrValidation, "target_audience must be at most 4096 characters", nil)
	}
	if len(req.BudgetRange) > MaxBrandProfileBudgetRangeLen {
		return domainErr.New(domainErr.ErrValidation, "budget_range must be at most 64 characters", nil)
	}
	if strings.TrimSpace(req.BrandValues) == "" {
		return domainErr.New(domainErr.ErrValidation, "brand_values is required", nil)
	}
	if len(req.BrandValues) > MaxBrandProfileBrandValuesLen {
		return domainErr.New(domainErr.ErrValidation, "brand_values must be at most 4096 characters", nil)
	}
	if strings.TrimSpace(req.CampaignGoal) == "" {
		return domainErr.New(domainErr.ErrValidation, "campaign_goal is required", nil)
	}
	if len(strings.TrimSpace(req.CampaignGoal)) > MaxBrandProfileCampaignGoalLen {
		return domainErr.New(domainErr.ErrValidation, "campaign_goal must be at most 128 characters", nil)
	}
	return nil
}

// BuildBrandContext formats saved brand profile fields for LLM analyze prompts.
func BuildBrandContext(industry, targetAudience, budgetRange, brandValues, campaignGoal string) string {
	var lines []string
	if v := strings.TrimSpace(industry); v != "" {
		lines = append(lines, "- Industry: "+v)
	}
	if v := strings.TrimSpace(targetAudience); v != "" {
		lines = append(lines, "- Target Audience: "+v)
	}
	if v := strings.TrimSpace(budgetRange); v != "" {
		lines = append(lines, "- Budget: "+v)
	}
	if v := strings.TrimSpace(brandValues); v != "" {
		lines = append(lines, "- Brand Values: "+v)
	}
	if v := strings.TrimSpace(campaignGoal); v != "" {
		lines = append(lines, "- Campaign Goal: "+v)
	}
	if len(lines) == 0 {
		return ""
	}
	return "Brand Context:\n" + strings.Join(lines, "\n")
}

// AppendBrandContext appends brand context after influencer profile and question notes.
func AppendBrandContext(notes, brandContext string) string {
	brandContext = strings.TrimSpace(brandContext)
	if brandContext == "" {
		return notes
	}
	notes = strings.TrimSpace(notes)
	if notes == "" {
		return brandContext
	}
	return notes + "\n\n" + brandContext
}
