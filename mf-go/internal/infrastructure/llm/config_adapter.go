package llm

import (
	"context"

	adminRepo "github.com/masterfabric-go/masterfabric/internal/domain/admin/repository"
)

// ConfigAdapter reads admin LLM config for the analyzer runtime.
type ConfigAdapter struct {
	repo adminRepo.LLMConfigRepository
}

func NewConfigAdapter(repo adminRepo.LLMConfigRepository) *ConfigAdapter {
	return &ConfigAdapter{repo: repo}
}

func (a *ConfigAdapter) GetRuntimeSettings(ctx context.Context) (RuntimeLLMSettings, error) {
	cfg, err := a.repo.Get(ctx)
	if err != nil {
		return RuntimeLLMSettings{}, err
	}
	return RuntimeLLMSettings{
		SystemPrompt: cfg.SystemPrompt,
		Temperature:  cfg.Temperature,
		MaxTokens:    cfg.MaxTokens,
		Model:        cfg.Model,
	}, nil
}

var _ LLMConfigReader = (*ConfigAdapter)(nil)
