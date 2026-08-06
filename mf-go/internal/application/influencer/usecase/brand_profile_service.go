package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/influencer/dto"
	"github.com/masterfabric-go/masterfabric/internal/domain/influencer/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/influencer/repository"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
)

type BrandProfileService struct {
	profiles repository.BrandProfileRepository
}

func NewBrandProfileService(profiles repository.BrandProfileRepository) *BrandProfileService {
	return &BrandProfileService{profiles: profiles}
}

func (s *BrandProfileService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateBrandProfileRequest) (*dto.BrandProfileResponse, error) {
	if err := dto.ValidateCreateBrandProfile(req); err != nil {
		return nil, err
	}

	profile := &model.BrandProfile{
		UserID:         userID,
		Name:           strings.TrimSpace(req.Name),
		Industry:       strings.TrimSpace(req.Industry),
		TargetAudience: strings.TrimSpace(req.TargetAudience),
		BudgetRange:    strings.TrimSpace(req.BudgetRange),
		BrandValues:    strings.TrimSpace(req.BrandValues),
		CampaignGoal:   strings.TrimSpace(req.CampaignGoal),
	}
	if err := s.profiles.Create(ctx, profile); err != nil {
		return nil, err
	}
	resp := toBrandProfileResponse(profile)
	return &resp, nil
}

func (s *BrandProfileService) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.BrandProfileResponse, error) {
	items, err := s.profiles.ListByUser(ctx, userID, dto.ClampListLimit(limit), dto.ClampListOffset(offset))
	if err != nil {
		return nil, err
	}
	out := make([]dto.BrandProfileResponse, 0, len(items))
	for i := range items {
		out = append(out, toBrandProfileResponse(&items[i]))
	}
	return out, nil
}

func (s *BrandProfileService) Get(ctx context.Context, userID, id uuid.UUID) (*dto.BrandProfileResponse, error) {
	profile, err := s.profiles.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	resp := toBrandProfileResponse(profile)
	return &resp, nil
}

func (s *BrandProfileService) Update(ctx context.Context, userID, id uuid.UUID, req dto.UpdateBrandProfileRequest) (*dto.BrandProfileResponse, error) {
	profile, err := s.profiles.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if err := applyBrandProfileUpdate(profile, req); err != nil {
		return nil, err
	}
	if err := s.profiles.Update(ctx, profile); err != nil {
		return nil, err
	}
	resp := toBrandProfileResponse(profile)
	return &resp, nil
}

func (s *BrandProfileService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.profiles.Delete(ctx, userID, id)
}

func applyBrandProfileUpdate(profile *model.BrandProfile, req dto.UpdateBrandProfileRequest) error {
	if req.Name != nil {
		if err := dto.ValidateBrandProfileName(*req.Name); err != nil {
			return err
		}
		profile.Name = strings.TrimSpace(*req.Name)
	}
	if req.Industry != nil {
		if strings.TrimSpace(*req.Industry) == "" {
			return domainErr.New(domainErr.ErrValidation, "industry is required", nil)
		}
		if len(strings.TrimSpace(*req.Industry)) > dto.MaxBrandProfileIndustryLen {
			return domainErr.New(domainErr.ErrValidation, "industry must be at most 128 characters", nil)
		}
		profile.Industry = strings.TrimSpace(*req.Industry)
	}
	if req.TargetAudience != nil {
		if strings.TrimSpace(*req.TargetAudience) == "" {
			return domainErr.New(domainErr.ErrValidation, "target_audience is required", nil)
		}
		if len(*req.TargetAudience) > dto.MaxBrandProfileTargetAudienceLen {
			return domainErr.New(domainErr.ErrValidation, "target_audience must be at most 4096 characters", nil)
		}
		profile.TargetAudience = strings.TrimSpace(*req.TargetAudience)
	}
	if req.BudgetRange != nil {
		if len(*req.BudgetRange) > dto.MaxBrandProfileBudgetRangeLen {
			return domainErr.New(domainErr.ErrValidation, "budget_range must be at most 64 characters", nil)
		}
		profile.BudgetRange = strings.TrimSpace(*req.BudgetRange)
	}
	if req.BrandValues != nil {
		if strings.TrimSpace(*req.BrandValues) == "" {
			return domainErr.New(domainErr.ErrValidation, "brand_values is required", nil)
		}
		if len(*req.BrandValues) > dto.MaxBrandProfileBrandValuesLen {
			return domainErr.New(domainErr.ErrValidation, "brand_values must be at most 4096 characters", nil)
		}
		profile.BrandValues = strings.TrimSpace(*req.BrandValues)
	}
	if req.CampaignGoal != nil {
		if strings.TrimSpace(*req.CampaignGoal) == "" {
			return domainErr.New(domainErr.ErrValidation, "campaign_goal is required", nil)
		}
		if len(strings.TrimSpace(*req.CampaignGoal)) > dto.MaxBrandProfileCampaignGoalLen {
			return domainErr.New(domainErr.ErrValidation, "campaign_goal must be at most 128 characters", nil)
		}
		profile.CampaignGoal = strings.TrimSpace(*req.CampaignGoal)
	}
	return nil
}

func toBrandProfileResponse(profile *model.BrandProfile) dto.BrandProfileResponse {
	return dto.BrandProfileResponse{
		ID:             profile.ID,
		UserID:         profile.UserID,
		Name:           profile.Name,
		Industry:       profile.Industry,
		TargetAudience: profile.TargetAudience,
		BudgetRange:    profile.BudgetRange,
		BrandValues:    profile.BrandValues,
		CampaignGoal:   profile.CampaignGoal,
		CreatedAt:      profile.CreatedAt,
		UpdatedAt:      profile.UpdatedAt,
	}
}
