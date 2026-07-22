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

type ScoreService struct {
	scores repository.ScoreRepository
}

func NewScoreService(scores repository.ScoreRepository) *ScoreService {
	return &ScoreService{scores: scores}
}

func (s *ScoreService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateScoreRequest) (*dto.ScoreResponse, error) {
	if err := validateCreateScore(req); err != nil {
		return nil, err
	}

	score := &model.InfluencerScore{
		UserID:          userID,
		InfluencerName:  strings.TrimSpace(req.InfluencerName),
		Platform:        dto.NormalizePlatform(req.Platform),
		OverallScore:    req.OverallScore,
		EngagementScore: req.EngagementScore,
		AudienceScore:   req.AudienceScore,
		BrandFitScore:   req.BrandFitScore,
		RawPayload:      req.RawPayload,
		Notes:           req.Notes,
	}
	if err := s.scores.Create(ctx, score); err != nil {
		return nil, err
	}
	resp := toScoreResponse(score)
	return &resp, nil
}

func (s *ScoreService) List(ctx context.Context, userID uuid.UUID, platform string, limit, offset int) ([]dto.ScoreResponse, error) {
	items, err := s.scores.ListByUser(ctx, userID, dto.NormalizePlatform(platform), dto.ClampListLimit(limit), dto.ClampListOffset(offset))
	if err != nil {
		return nil, err
	}
	out := make([]dto.ScoreResponse, 0, len(items))
	for i := range items {
		out = append(out, toScoreResponse(&items[i]))
	}
	return out, nil
}

func (s *ScoreService) Get(ctx context.Context, userID, id uuid.UUID) (*dto.ScoreResponse, error) {
	score, err := s.scores.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	resp := toScoreResponse(score)
	return &resp, nil
}

func (s *ScoreService) Update(ctx context.Context, userID, id uuid.UUID, req dto.UpdateScoreRequest) (*dto.ScoreResponse, error) {
	score, err := s.scores.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if err := applyScoreUpdate(score, req); err != nil {
		return nil, err
	}
	if err := s.scores.Update(ctx, score); err != nil {
		return nil, err
	}
	resp := toScoreResponse(score)
	return &resp, nil
}

func (s *ScoreService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.scores.Delete(ctx, userID, id)
}

func validateCreateScore(req dto.CreateScoreRequest) error {
	if err := dto.ValidateInfluencerName(req.InfluencerName); err != nil {
		return err
	}
	if err := dto.ValidatePlatform(req.Platform); err != nil {
		return err
	}
	for _, v := range []float64{req.OverallScore, req.EngagementScore, req.AudienceScore, req.BrandFitScore} {
		if err := dto.ValidateScoreValue(v); err != nil {
			return err
		}
	}
	if len(req.Notes) > dto.MaxNotesLen {
		return domainErr.New(domainErr.ErrValidation, "notes must be at most 4096 characters", nil)
	}
	if len(req.RawPayload) > dto.MaxRawPayloadLen {
		return domainErr.New(domainErr.ErrValidation, "raw_payload must be at most 65536 characters", nil)
	}
	return nil
}

func applyScoreUpdate(score *model.InfluencerScore, req dto.UpdateScoreRequest) error {
	if req.InfluencerName != nil {
		if err := dto.ValidateInfluencerName(*req.InfluencerName); err != nil {
			return err
		}
		score.InfluencerName = strings.TrimSpace(*req.InfluencerName)
	}
	if req.Platform != nil {
		if err := dto.ValidatePlatform(*req.Platform); err != nil {
			return err
		}
		score.Platform = dto.NormalizePlatform(*req.Platform)
	}
	if req.OverallScore != nil {
		if err := dto.ValidateScoreValue(*req.OverallScore); err != nil {
			return err
		}
		score.OverallScore = *req.OverallScore
	}
	if req.EngagementScore != nil {
		if err := dto.ValidateScoreValue(*req.EngagementScore); err != nil {
			return err
		}
		score.EngagementScore = *req.EngagementScore
	}
	if req.AudienceScore != nil {
		if err := dto.ValidateScoreValue(*req.AudienceScore); err != nil {
			return err
		}
		score.AudienceScore = *req.AudienceScore
	}
	if req.BrandFitScore != nil {
		if err := dto.ValidateScoreValue(*req.BrandFitScore); err != nil {
			return err
		}
		score.BrandFitScore = *req.BrandFitScore
	}
	if req.RawPayload != nil {
		if len(*req.RawPayload) > dto.MaxRawPayloadLen {
			return domainErr.New(domainErr.ErrValidation, "raw_payload must be at most 65536 characters", nil)
		}
		score.RawPayload = *req.RawPayload
	}
	if req.Notes != nil {
		if len(*req.Notes) > dto.MaxNotesLen {
			return domainErr.New(domainErr.ErrValidation, "notes must be at most 4096 characters", nil)
		}
		score.Notes = *req.Notes
	}
	return nil
}

func toScoreResponse(score *model.InfluencerScore) dto.ScoreResponse {
	return dto.ScoreResponse{
		ID:              score.ID,
		UserID:          score.UserID,
		InfluencerName:  score.InfluencerName,
		Platform:        score.Platform,
		OverallScore:    score.OverallScore,
		EngagementScore: score.EngagementScore,
		AudienceScore:   score.AudienceScore,
		BrandFitScore:   score.BrandFitScore,
		RawPayload:      score.RawPayload,
		Notes:             score.Notes,
		CreatedAt:         score.CreatedAt,
		UpdatedAt:         score.UpdatedAt,
	}
}
