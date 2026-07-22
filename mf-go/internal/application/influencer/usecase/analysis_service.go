package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/influencer/dto"
	"github.com/masterfabric-go/masterfabric/internal/domain/influencer/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/influencer/repository"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
	infraRedis "github.com/masterfabric-go/masterfabric/internal/infrastructure/redis"
)

type AnalysisService struct {
	analyses repository.AnalysisRepository
	scores   repository.ScoreRepository
}

func NewAnalysisService(analyses repository.AnalysisRepository, scores repository.ScoreRepository) *AnalysisService {
	return &AnalysisService{analyses: analyses, scores: scores}
}

func (s *AnalysisService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateAnalysisRequest) (*dto.AnalysisResponse, error) {
	if err := dto.ValidateInfluencerName(req.InfluencerName); err != nil {
		return nil, err
	}
	if err := dto.ValidatePlatform(req.Platform); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.AnalysisType) == "" {
		return nil, domainErr.New(domainErr.ErrValidation, "analysis_type is required", nil)
	}
	if strings.TrimSpace(req.Summary) == "" {
		return nil, domainErr.New(domainErr.ErrValidation, "summary is required", nil)
	}
	if len(req.RawLLMOutput) > dto.MaxRawLLMOutputLen {
		return nil, domainErr.New(domainErr.ErrValidation, "raw_llm_output must be at most 65536 characters", nil)
	}
	if req.ScoreID != nil {
		if _, err := s.scores.GetByID(ctx, userID, *req.ScoreID); err != nil {
			return nil, domainErr.New(domainErr.ErrBadRequest, "linked score_id not found", err)
		}
	}

	analysis := &model.InfluencerAnalysis{
		UserID:         userID,
		InfluencerName: strings.TrimSpace(req.InfluencerName),
		Platform:       dto.NormalizePlatform(req.Platform),
		AnalysisType:   strings.TrimSpace(req.AnalysisType),
		Summary:        req.Summary,
		Insights:       req.Insights,
		RawLLMOutput:   req.RawLLMOutput,
		ScoreID:        req.ScoreID,
	}
	if err := s.analyses.Create(ctx, analysis); err != nil {
		return nil, err
	}
	resp := toAnalysisResponse(analysis)
	return &resp, nil
}

func (s *AnalysisService) List(ctx context.Context, userID uuid.UUID, platform string, limit, offset int) ([]dto.AnalysisResponse, error) {
	items, err := s.analyses.ListByUser(ctx, userID, dto.NormalizePlatform(platform), dto.ClampListLimit(limit), dto.ClampListOffset(offset))
	if err != nil {
		return nil, err
	}
	out := make([]dto.AnalysisResponse, 0, len(items))
	for i := range items {
		out = append(out, toAnalysisResponse(&items[i]))
	}
	return out, nil
}

func (s *AnalysisService) Get(ctx context.Context, userID, id uuid.UUID) (*dto.AnalysisResponse, error) {
	analysis, err := s.analyses.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	resp := toAnalysisResponse(analysis)
	return &resp, nil
}

func toAnalysisResponse(a *model.InfluencerAnalysis) dto.AnalysisResponse {
	return dto.AnalysisResponse{
		ID:             a.ID,
		UserID:         a.UserID,
		InfluencerName: a.InfluencerName,
		Platform:       a.Platform,
		AnalysisType:   a.AnalysisType,
		Summary:        a.Summary,
		Insights:       a.Insights,
		RawLLMOutput:   a.RawLLMOutput,
		ScoreID:        a.ScoreID,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}

type MonitoringService struct {
	metrics *infraRedis.LLMMetricsStore
}

func NewMonitoringService(metrics *infraRedis.LLMMetricsStore) *MonitoringService {
	return &MonitoringService{metrics: metrics}
}

func (s *MonitoringService) Record(ctx context.Context, userID uuid.UUID, req dto.RecordLLMMetricRequest) (*dto.LLMCallMetric, error) {
	if err := dto.ValidateInfluencerName(req.InfluencerName); err != nil {
		return nil, err
	}
	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status != "success" && status != "error" {
		return nil, domainErr.New(domainErr.ErrValidation, "status must be success or error", nil)
	}
	if req.LatencyMs < 0 {
		return nil, domainErr.New(domainErr.ErrValidation, "latency_ms must be non-negative", nil)
	}
	modelName := strings.TrimSpace(req.Model)
	if modelName == "" {
		return nil, domainErr.New(domainErr.ErrValidation, "model is required", nil)
	}
	if len(modelName) > 128 {
		return nil, domainErr.New(domainErr.ErrValidation, "model must be at most 128 characters", nil)
	}

	metric := &dto.LLMCallMetric{
		InfluencerName: strings.TrimSpace(req.InfluencerName),
		LatencyMs:      req.LatencyMs,
		Status:         status,
		Model:          modelName,
	}
	if err := s.metrics.Record(ctx, userID, metric); err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "failed to record metric", err)
	}
	return metric, nil
}

func (s *MonitoringService) Stats(ctx context.Context, userID uuid.UUID) (*dto.MonitoringStatsResponse, error) {
	stats, err := s.metrics.GetStats(ctx, userID)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "failed to fetch monitoring stats", err)
	}
	return stats, nil
}
