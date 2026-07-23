package dto_test

import (
	"testing"

	"github.com/masterfabric-go/masterfabric/internal/application/influencer/dto"
)

func TestValidatePlatform(t *testing.T) {
	if err := dto.ValidatePlatform("instagram"); err != nil {
		t.Fatalf("expected valid platform, got %v", err)
	}
	if err := dto.ValidatePlatform("invalid"); err == nil {
		t.Fatal("expected invalid platform error")
	}
}

func TestValidateScoreValue(t *testing.T) {
	if err := dto.ValidateScoreValue(50); err != nil {
		t.Fatalf("expected valid score, got %v", err)
	}
	if err := dto.ValidateScoreValue(101); err == nil {
		t.Fatal("expected score > 100 to fail")
	}
}

func TestScoreValueOrDefault(t *testing.T) {
	if got := dto.ScoreValueOrDefault(nil); got != 0 {
		t.Fatalf("expected 0 default, got %v", got)
	}
	v := 42.0
	if got := dto.ScoreValueOrDefault(&v); got != 42 {
		t.Fatalf("expected 42, got %v", got)
	}
}

func TestValidateOptionalScoreValue(t *testing.T) {
	if err := dto.ValidateOptionalScoreValue(nil); err != nil {
		t.Fatalf("expected nil score to skip validation, got %v", err)
	}
	bad := 150.0
	if err := dto.ValidateOptionalScoreValue(&bad); err == nil {
		t.Fatal("expected invalid optional score to fail")
	}
}

func TestClampListLimit(t *testing.T) {
	if got := dto.ClampListLimit(0); got != dto.DefaultListLimit {
		t.Fatalf("expected default limit %d, got %d", dto.DefaultListLimit, got)
	}
	if got := dto.ClampListLimit(9999); got != dto.MaxListLimit {
		t.Fatalf("expected max limit %d, got %d", dto.MaxListLimit, got)
	}
}
