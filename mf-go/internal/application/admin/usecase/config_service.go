package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/admin/dto"
	"github.com/masterfabric-go/masterfabric/internal/domain/admin/model"
	adminRepo "github.com/masterfabric-go/masterfabric/internal/domain/admin/repository"
)

type ConfigService struct {
	config adminRepo.LLMConfigRepository
}

func NewConfigService(config adminRepo.LLMConfigRepository) *ConfigService {
	return &ConfigService{config: config}
}

func (s *ConfigService) Get(ctx context.Context) (*dto.LLMConfigResponse, error) {
	cfg, err := s.config.Get(ctx)
	if err != nil {
		return nil, err
	}
	return toConfigResponse(cfg), nil
}

func (s *ConfigService) Update(ctx context.Context, userID uuid.UUID, req dto.UpdateLLMConfigRequest) (*dto.LLMConfigResponse, error) {
	if err := dto.ValidateUpdateLLMConfig(req); err != nil {
		return nil, err
	}
	cfg, err := s.config.Update(ctx, &model.LLMConfig{
		SystemPrompt: req.SystemPrompt,
		Temperature:  req.Temperature,
		MaxTokens:    req.MaxTokens,
		Model:        req.Model,
		UpdatedBy:    &userID,
	})
	if err != nil {
		return nil, err
	}
	return toConfigResponse(cfg), nil
}

func toConfigResponse(cfg *model.LLMConfig) *dto.LLMConfigResponse {
	return &dto.LLMConfigResponse{
		ID:           cfg.ID,
		SystemPrompt: cfg.SystemPrompt,
		Temperature:  cfg.Temperature,
		MaxTokens:    cfg.MaxTokens,
		Model:        cfg.Model,
		UpdatedAt:    cfg.UpdatedAt,
		UpdatedBy:    cfg.UpdatedBy,
	}
}
