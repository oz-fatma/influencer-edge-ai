package llm

import (
	"testing"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/shared/config"
)

func TestParseAnalysisJSON_validObject(t *testing.T) {
	raw := `{"overall_score":82.5,"engagement_score":88,"audience_score":79,"brand_fit_score":80,"summary":"Strong profile","insights":["Good engagement","Audience fit"]}`
	result, err := parseAnalysisJSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary != "Strong profile" {
		t.Fatalf("summary = %q", result.Summary)
	}
	if len(result.Insights) != 2 {
		t.Fatalf("insights len = %d", len(result.Insights))
	}
	if result.OverallScore != 82.5 {
		t.Fatalf("overall_score = %v", result.OverallScore)
	}
}

func TestParseAnalysisJSON_stripsCodeFence(t *testing.T) {
	raw := "```json\n{\"overall_score\":70,\"engagement_score\":70,\"audience_score\":70,\"brand_fit_score\":70,\"summary\":\"OK\",\"insights\":[\"One insight\"]}\n```"
	result, err := parseAnalysisJSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary != "OK" {
		t.Fatalf("summary = %q", result.Summary)
	}
}

func TestParseAnalysisJSON_invalid(t *testing.T) {
	_, err := parseAnalysisJSON("not json at all")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewAnalyzer_nilWhenNoBaseURL(t *testing.T) {
	if NewAnalyzer(config.LLMConfig{Timeout: time.Minute}) != nil {
		t.Fatal("expected nil analyzer")
	}
}
