package dto

import (
	"time"

	"github.com/google/uuid"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
)

const (
	DefaultSystemPrompt = "You are an expert influencer marketing analyst. ONLY return valid JSON. No markdown, no explanation, no code fences."
	DefaultTemperature  = 0.1
	DefaultMaxTokens    = 100
	DefaultModel        = "gemma-influencer-ft"
)

var allowedAdminModels = map[string]struct{}{
	"gemma-influencer-ft": {},
	"gemma2:2b":           {},
}

type LLMConfigResponse struct {
	ID           uuid.UUID  `json:"id"`
	SystemPrompt string     `json:"system_prompt"`
	Temperature  float64    `json:"temperature"`
	MaxTokens    int        `json:"max_tokens"`
	Model        string     `json:"model"`
	UpdatedAt    time.Time  `json:"updated_at"`
	UpdatedBy    *uuid.UUID `json:"updated_by,omitempty"`
}

type UpdateLLMConfigRequest struct {
	SystemPrompt string  `json:"system_prompt" validate:"required"`
	Temperature  float64 `json:"temperature" validate:"required"`
	MaxTokens    int     `json:"max_tokens" validate:"required"`
	Model        string  `json:"model" validate:"required"`
}

type LLMLogEntry struct {
	ModelName  string    `json:"model_name"`
	DurationMs int64     `json:"duration_ms"`
	Success    bool      `json:"success"`
	CreatedAt  time.Time `json:"created_at"`
}

type LLMLogsResponse struct {
	Logs []LLMLogEntry `json:"logs"`
}

type MeResponse struct {
	User    UserInfo `json:"user"`
	IsAdmin bool     `json:"is_admin"`
}

type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func ValidateUpdateLLMConfig(req UpdateLLMConfigRequest) error {
	if len(req.SystemPrompt) == 0 || len(req.SystemPrompt) > 8192 {
		return domainErr.New(domainErr.ErrValidation, "system_prompt must be between 1 and 8192 characters", nil)
	}
	if req.Temperature < 0 || req.Temperature > 1 {
		return domainErr.New(domainErr.ErrValidation, "temperature must be between 0 and 1", nil)
	}
	if req.MaxTokens < 1 || req.MaxTokens > 2000 {
		return domainErr.New(domainErr.ErrValidation, "max_tokens must be between 1 and 2000", nil)
	}
	if _, ok := allowedAdminModels[req.Model]; !ok {
		return domainErr.New(domainErr.ErrValidation, "model must be gemma-influencer-ft or gemma2:2b", nil)
	}
	return nil
}

func AllowedModels() []string {
	return []string{"gemma-influencer-ft", "gemma2:2b"}
}
