package dto_test

import (
	"strings"
	"testing"

	"github.com/masterfabric-go/masterfabric/internal/application/admin/dto"
)

func TestValidateUpdateLLMConfig(t *testing.T) {
	valid := dto.UpdateLLMConfigRequest{
		SystemPrompt: "You are an analyst.",
		Temperature:  0.1,
		MaxTokens:    100,
		Model:        "gemma-influencer-ft",
	}
	if err := dto.ValidateUpdateLLMConfig(valid); err != nil {
		t.Fatalf("valid config: %v", err)
	}

	invalidModel := valid
	invalidModel.Model = "llama3"
	if err := dto.ValidateUpdateLLMConfig(invalidModel); err == nil {
		t.Fatal("expected error for invalid model")
	}

	invalidTemp := valid
	invalidTemp.Temperature = 1.5
	if err := dto.ValidateUpdateLLMConfig(invalidTemp); err == nil {
		t.Fatal("expected error for temperature > 1")
	}

	emptyPrompt := valid
	emptyPrompt.SystemPrompt = ""
	if err := dto.ValidateUpdateLLMConfig(emptyPrompt); err == nil {
		t.Fatal("expected error for empty prompt")
	}

	longPrompt := valid
	longPrompt.SystemPrompt = strings.Repeat("a", 8193)
	if err := dto.ValidateUpdateLLMConfig(longPrompt); err == nil {
		t.Fatal("expected error for prompt too long")
	}
}

func TestAllowedModels(t *testing.T) {
	models := dto.AllowedModels()
	if len(models) != 3 {
		t.Fatalf("models len = %d", len(models))
	}
}
