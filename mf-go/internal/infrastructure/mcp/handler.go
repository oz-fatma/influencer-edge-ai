package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/influencer/dto"
	domainRepo "github.com/masterfabric-go/masterfabric/internal/domain/influencer/repository"
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
	analyzer      *llm.Analyzer
	model         string
	brandProfiles domainRepo.BrandProfileRepository
}

func NewService(analyzer *llm.Analyzer, model string, brandProfiles domainRepo.BrandProfileRepository) *Service {
	return &Service{
		analyzer:      analyzer,
		model:         strings.TrimSpace(model),
		brandProfiles: brandProfiles,
	}
}

// HandleMCPRequest validates the payload, selects an adapter, proxies to Ollama, and enriches the result.
func (s *Service) HandleMCPRequest(ctx context.Context, userID uuid.UUID, req MCPPayload) (RichResult, error) {
	start := time.Now()

	if _, ok := supportedRequestTypes[req.RequestType]; !ok {
		return RichResult{}, fmt.Errorf("unsupported request_type: %q", req.RequestType)
	}
	if s.analyzer == nil {
		return RichResult{}, fmt.Errorf("LLM service not configured (set LLM_BASE_URL on the server)")
	}

	model := s.resolveModel(req)

	switch req.RequestType {
	case RequestTypeAnalyzeInfluencer:
		return s.handleAnalyzeInfluencer(ctx, userID, req, model, start)
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
	userID uuid.UUID,
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

	query := strings.TrimSpace(req.Query)
	if query == "" {
		query = dto.DefaultAnalyzeQuery
	}
	if err := dto.ValidateAnalyzeQuery(query); err != nil {
		return RichResult{}, fmt.Errorf("query: %s", err.Error())
	}

	notes := dto.BuildAnalyzePromptWithQuery(profile, legacyNotes, query)
	notes = dto.AppendBrandContext(notes, s.resolveBrandContext(ctx, userID, req.BrandProfileID))

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

func (s *Service) resolveBrandContext(ctx context.Context, userID uuid.UUID, brandProfileID *string) string {
	if s == nil || s.brandProfiles == nil || brandProfileID == nil {
		return ""
	}
	raw := strings.TrimSpace(*brandProfileID)
	if raw == "" {
		return ""
	}
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return ""
	}

	profile, err := s.brandProfiles.GetByID(ctx, userID, id)
	if err != nil {
		return ""
	}

	return dto.BuildBrandContext(
		profile.Industry,
		profile.TargetAudience,
		profile.BudgetRange,
		profile.BrandValues,
		profile.CampaignGoal,
	)
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
