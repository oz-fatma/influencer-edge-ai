package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/infrastructure/llm"
	"github.com/masterfabric-go/masterfabric/internal/shared/config"
)

func TestHandleMCPRequest_usesDefaultQueryWhenEmpty(t *testing.T) {
	mockLLM := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": `{"overall_score":82.5,"engagement_score":88,"audience_score":79,"brand_fit_score":80,"summary":"Strong profile","insights":["Good engagement"]}`,
					},
				},
			},
		})
	}))
	defer mockLLM.Close()

	analyzer := llm.NewAnalyzer(config.LLMConfig{
		BaseURL: mockLLM.URL,
		Model:   "gemma-influencer-ft",
		Timeout: time.Minute,
	}, nil, nil)
	svc := NewService(analyzer, "gemma-influencer-ft")

	result, err := svc.HandleMCPRequest(context.Background(), MCPPayload{
		RequestType: RequestTypeAnalyzeInfluencer,
		Query:       "   ",
		Context: map[string]any{
			"influencer_name": "Ada Lovelace",
			"platform":        "instagram",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Data["summary"] != "Strong profile" {
		t.Fatalf("data.summary = %v", result.Data["summary"])
	}
}

func TestHandleMCPRequest_rejectsQueryTooLong(t *testing.T) {
	svc := NewService(&llm.Analyzer{}, "gemma-influencer-ft")

	_, err := svc.HandleMCPRequest(context.Background(), MCPPayload{
		RequestType: RequestTypeAnalyzeInfluencer,
		Query:       strings.Repeat("a", 501),
		Context: map[string]any{
			"influencer_name": "Ada Lovelace",
			"platform":        "instagram",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "query") {
		t.Fatalf("expected query validation error, got %v", err)
	}
}

func TestHandleMCPRequest_rejectsUnknownRequestType(t *testing.T) {
	svc := NewService(&llm.Analyzer{}, "gemma-influencer-ft")

	_, err := svc.HandleMCPRequest(context.Background(), MCPPayload{
		RequestType: "unknown_type",
		Query:       "analyze this influencer",
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported request_type") {
		t.Fatalf("expected unsupported request_type error, got %v", err)
	}
}

func TestHandleMCPRequest_requiresConfiguredAnalyzer(t *testing.T) {
	svc := NewService(nil, "gemma-influencer-ft")

	_, err := svc.HandleMCPRequest(context.Background(), MCPPayload{
		RequestType: RequestTypeAnalyzeInfluencer,
		Query:       "Tech niche creator",
		Context: map[string]any{
			"influencer_name": "Ada Lovelace",
			"platform":        "instagram",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("expected not configured error, got %v", err)
	}
}

func TestHandleMCPRequest_requiresInfluencerContext(t *testing.T) {
	svc := NewService(&llm.Analyzer{}, "gemma-influencer-ft")

	_, err := svc.HandleMCPRequest(context.Background(), MCPPayload{
		RequestType: RequestTypeAnalyzeInfluencer,
		Query:       "Tech niche creator",
		Context: map[string]any{
			"platform": "instagram",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "context.influencer_name") {
		t.Fatalf("expected influencer_name validation error, got %v", err)
	}
}

func TestHandleMCPRequest_analyzeInfluencerSuccess(t *testing.T) {
	mockLLM := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": `{"overall_score":82.5,"engagement_score":88,"audience_score":79,"brand_fit_score":80,"summary":"Strong profile","insights":["Good engagement","Audience fit"]}`,
					},
				},
			},
		})
	}))
	defer mockLLM.Close()

	analyzer := llm.NewAnalyzer(config.LLMConfig{
		BaseURL: mockLLM.URL,
		Model:   "gemma-influencer-ft",
		Timeout: time.Minute,
	}, nil, nil)
	svc := NewService(analyzer, "gemma-influencer-ft")

	result, err := svc.HandleMCPRequest(context.Background(), MCPPayload{
		RequestType: RequestTypeAnalyzeInfluencer,
		Query:       "Focus on brand-fit for cosmetics campaigns",
		Context: map[string]any{
			"influencer_name": "Ada Lovelace",
			"platform":        "instagram",
			"notes":           "Tech & education niche",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Source != SourceOllama {
		t.Fatalf("source = %q, want %q", result.Source, SourceOllama)
	}
	if result.Metadata["model"] != "gemma-influencer-ft" {
		t.Fatalf("metadata.model = %v", result.Metadata["model"])
	}
	if result.Metadata["request_type"] != RequestTypeAnalyzeInfluencer {
		t.Fatalf("metadata.request_type = %v", result.Metadata["request_type"])
	}
	if _, ok := result.Metadata["latency_ms"]; !ok {
		t.Fatal("expected metadata.latency_ms")
	}
	if _, ok := result.Metadata["timestamp"]; !ok {
		t.Fatal("expected metadata.timestamp")
	}
	if result.Data["summary"] != "Strong profile" {
		t.Fatalf("data.summary = %v", result.Data["summary"])
	}
	if result.Data["raw_output"] == "" {
		t.Fatal("expected raw_output in data")
	}
}

func TestResolveModel_prefersConfiguredModel(t *testing.T) {
	svc := NewService(nil, "gemma-influencer-ft")
	if got := svc.resolveModel(MCPPayload{}); got != "gemma-influencer-ft" {
		t.Fatalf("model = %q, want gemma-influencer-ft", got)
	}
}
