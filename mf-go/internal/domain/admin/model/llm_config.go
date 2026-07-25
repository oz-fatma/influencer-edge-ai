package model

import (
	"time"

	"github.com/google/uuid"
)

const SingletonConfigID = "00000000-0000-0000-0000-000000000001"

type LLMConfig struct {
	ID           uuid.UUID
	SystemPrompt string
	Temperature  float64
	MaxTokens    int
	Model        string
	UpdatedAt    time.Time
	UpdatedBy    *uuid.UUID
}

type LLMRequestLog struct {
	ModelName  string
	DurationMs int64
	Success    bool
	CreatedAt  time.Time
}
