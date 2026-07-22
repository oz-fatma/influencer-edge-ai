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

func TestClampListLimit(t *testing.T) {
	if got := dto.ClampListLimit(0); got != dto.DefaultListLimit {
		t.Fatalf("expected default limit %d, got %d", dto.DefaultListLimit, got)
	}
	if got := dto.ClampListLimit(9999); got != dto.MaxListLimit {
		t.Fatalf("expected max limit %d, got %d", dto.MaxListLimit, got)
	}
}
