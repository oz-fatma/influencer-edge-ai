package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/application/influencer/dto"
	"github.com/masterfabric-go/masterfabric/internal/infrastructure/llm"
)

const (
	RequestTypeAnalyzeInfluencer = "analyze_influencer"
	SourceOllama                 = "mcp-ollama"
)

var supportedRequestTypes = map[string]struct{}{
	RequestTypeAnalyzeInfluencer: {},
}

// Service routes MCP requests to the configured LLM adapter.
type Service struct {
	analyzer *llm.Analyzer
	model    string
}

func NewService(analyzer *llm.Analyzer, model string) *Service {
	return &Service{analyzer: analyzer, model: strings.TrimSpace(model)}
}

// HandleMCPRequest validates the payload, selects an adapter, proxies to Ollama, and enriches the result.
func (s *Service) HandleMCPRequest(ctx context.Context, req MCPPayload) (RichResult, error) {
	start := time.Now()

	if strings.TrimSpace(req.Query) == "" {
		return RichResult{}, fmt.Errorf("query is required")
	}
	if _, ok := supportedRequestTypes[req.RequestType]; !ok {
		return RichResult{}, fmt.Errorf("unsupported request_type: %q", req.RequestType)
	}
	if s.analyzer == nil {
		return RichResult{}, fmt.Errorf("LLM service not configured (set LLM_BASE_URL on the server)")
	}

	model := s.resolveModel(req)

	switch req.RequestType {
	case RequestTypeAnalyzeInfluencer:
		return s.handleAnalyzeInfluencer(ctx, req, model, start)
	default:
		return RichResult{}, fmt.Errorf("unsupported request_type: %q", req.RequestType)
	}
}

func (s *Service) resolveModel(_ MCPPayload) string {
	if s.model != "" {
		return s.model
	}
	if s.analyzer != nil {
		return s.analyzer.Model()
	}
	return "gemma-influencer-ft"
}

func (s *Service) handleAnalyzeInfluencer(
	ctx context.Context,
	req MCPPayload,
	model string,
	start time.Time,
) (RichResult, error) {
	name := contextString(req.Context, "influencer_name")
	platform := contextString(req.Context, "platform")
	profile := dto.ProfileFromMap(req.Context)
	legacyNotes := strings.TrimSpace(profile.Notes)
	if legacyNotes == "" {
		legacyNotes = contextString(req.Context, "notes")
	}
	profile.Notes = ""
	notes := buildNotes(dto.BuildAnalyzePrompt(profile, legacyNotes), req.Query)

	if err := dto.ValidateInfluencerName(name); err != nil {
		return RichResult{}, fmt.Errorf("context.influencer_name: %s", err.Error())
	}
	if err := dto.ValidatePlatform(platform); err != nil {
		return RichResult{}, fmt.Errorf("context.platform: %s", err.Error())
	}
	if len(notes) > dto.MaxNotesLen {
		return RichResult{}, fmt.Errorf("notes must be at most %d characters", dto.MaxNotesLen)
	}

	result, rawOutput, err := s.analyzer.Analyze(ctx, name, platform, notes)
	if err != nil {
		return RichResult{}, err
	}

	return RichResult{
		Data: map[string]any{
			"overall_score":    result.OverallScore,
			"engagement_score": result.EngagementScore,
			"audience_score":   result.AudienceScore,
			"brand_fit_score":  result.BrandFitScore,
			"summary":          result.Summary,
			"insights":         result.Insights,
			"raw_output":       rawOutput,
		},
		Metadata: map[string]any{
			"model":        model,
			"latency_ms":   time.Since(start).Milliseconds(),
			"timestamp":    time.Now().UTC().Format(time.RFC3339Nano),
			"request_type": req.RequestType,
		},
		Source: SourceOllama,
	}, nil
}

func contextString(ctx map[string]any, key string) string {
	if ctx == nil {
		return ""
	}
	value, ok := ctx[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func buildNotes(contextNotes, query string) string {
	contextNotes = strings.TrimSpace(contextNotes)
	query = strings.TrimSpace(query)
	if contextNotes == "" {
		return query
	}
	if query == "" || query == contextNotes {
		return contextNotes
	}
	return contextNotes + "\n\n" + query
}
