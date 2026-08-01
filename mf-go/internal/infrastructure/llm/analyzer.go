package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/application/influencer/dto"
	"github.com/masterfabric-go/masterfabric/internal/shared/config"
)

const (
	endpointTypeChat        = "chat"
	endpointTypeHFInference = "hf-inference"
	hfModelTurnPrefix       = "<start_of_turn>model\n"
	hfInferenceMaxNewTokens = 280
)

const defaultSystemPrompt = "You are an expert influencer marketing analyst. ONLY return valid JSON. No markdown, no explanation, no code fences."

// ollamaHTTPTimeout allows for fine-tuned model cold start (first load into memory).
const ollamaHTTPTimeout = 300 * time.Second

// LLMRequestWriter persists outbound LLM call metadata (model, prompt size, latency, success).
type LLMRequestWriter interface {
	Insert(ctx context.Context, modelName string, promptLength int, durationMs int64, success bool) error
}

// RuntimeLLMSettings holds per-request LLM parameters (from admin config or defaults).
type RuntimeLLMSettings struct {
	SystemPrompt string
	Temperature  float64
	MaxTokens    int
	Model        string
}

// LLMConfigReader loads runtime LLM settings from persistent admin config.
type LLMConfigReader interface {
	GetRuntimeSettings(ctx context.Context) (RuntimeLLMSettings, error)
}

const fewShotExample = `{
  "overall_score": 82.5,
  "engagement_score": 88,
  "audience_score": 79,
  "brand_fit_score": 80,
  "summary": "Strong engagement rate and audience alignment. Suitable profile for cosmetics campaigns.",
  "insights": [
    "Average 4.1% engagement over the last 30 days",
    "65% of followers match the target demographic",
    "Sponsored content reaches 90% of organic performance"
  ]
}`

// AnalysisResult mirrors the WebLLM JSON shape used by the matching panel.
type AnalysisResult struct {
	OverallScore    float64  `json:"overall_score"`
	EngagementScore float64  `json:"engagement_score"`
	AudienceScore   float64  `json:"audience_score"`
	BrandFitScore   float64  `json:"brand_fit_score"`
	Summary         string   `json:"summary"`
	Insights        []string `json:"insights"`
}

// Analyzer calls an OpenAI-compatible LLM server or HF Inference Endpoint.
type Analyzer struct {
	baseURL      string
	model        string
	apiKey       string
	endpointType string
	client       *http.Client
	requestLog   LLMRequestWriter
	config       LLMConfigReader
}

func NewAnalyzer(cfg config.LLMConfig, requestLog LLMRequestWriter, configReader LLMConfigReader) *Analyzer {
	if cfg.BaseURL == "" {
		return nil
	}
	endpointType := strings.ToLower(strings.TrimSpace(cfg.EndpointType))
	if endpointType == "" {
		endpointType = endpointTypeChat
	}
	return &Analyzer{
		baseURL:      strings.TrimRight(cfg.BaseURL, "/"),
		model:        cfg.Model,
		apiKey:       cfg.APIKey,
		endpointType: endpointType,
		client: &http.Client{
			Timeout: resolveHTTPTimeout(cfg),
		},
		requestLog: requestLog,
		config:     configReader,
	}
}

func resolveHTTPTimeout(cfg config.LLMConfig) time.Duration {
	if cfg.Timeout <= 0 {
		return ollamaHTTPTimeout
	}
	if cfg.Timeout < ollamaHTTPTimeout {
		return ollamaHTTPTimeout
	}
	return cfg.Timeout
}

func (a *Analyzer) Model() string {
	return a.model
}

func (a *Analyzer) resolveSettings(ctx context.Context) RuntimeLLMSettings {
	defaults := RuntimeLLMSettings{
		SystemPrompt: defaultSystemPrompt,
		Temperature:  0.1,
		MaxTokens:    100,
		Model:        a.model,
	}
	if a.config == nil {
		return defaults
	}
	settings, err := a.config.GetRuntimeSettings(ctx)
	if err != nil {
		return defaults
	}
	if settings.SystemPrompt == "" {
		settings.SystemPrompt = defaults.SystemPrompt
	}
	if settings.Model == "" {
		settings.Model = defaults.Model
	}
	if settings.MaxTokens <= 0 {
		settings.MaxTokens = defaults.MaxTokens
	}
	return settings
}

func (a *Analyzer) Analyze(ctx context.Context, name, platform, notes string) (*AnalysisResult, string, error) {
	return a.AnalyzeWithProfile(ctx, name, platform, dto.InfluencerProfile{Notes: notes}, notes)
}

func (a *Analyzer) AnalyzeWithProfile(ctx context.Context, name, platform string, profile dto.InfluencerProfile, legacyNotes string) (*AnalysisResult, string, error) {
	settings := a.resolveSettings(ctx)
	userPrompt := buildPrompt(name, platform, dto.BuildAnalyzePrompt(profile, legacyNotes))

	if a.endpointType == endpointTypeHFInference {
		return a.analyzeViaHFInference(ctx, settings, userPrompt)
	}
	return a.analyzeViaChat(ctx, settings, userPrompt)
}

func (a *Analyzer) analyzeViaChat(ctx context.Context, settings RuntimeLLMSettings, userPrompt string) (*AnalysisResult, string, error) {
	promptLength := len(settings.SystemPrompt) + len(userPrompt)

	start := time.Now()
	success := false
	defer func() {
		a.recordRequest(settings.Model, promptLength, time.Since(start).Milliseconds(), success)
	}()

	payload := chatCompletionRequest{
		Model: settings.Model,
		Messages: []chatMessage{
			{
				Role:    "system",
				Content: settings.SystemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Temperature: settings.Temperature,
		MaxTokens:   settings.MaxTokens,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("create chat request: %w", err)
	}
	a.setRequestHeaders(req)

	respBody, err := a.doRequest(req)
	if err != nil {
		return nil, "", err
	}

	var completion chatCompletionResponse
	if err := json.Unmarshal(respBody, &completion); err != nil {
		return nil, "", fmt.Errorf("decode LLM response: %w", err)
	}

	rawOutput := strings.TrimSpace(completion.firstContent())
	if rawOutput == "" {
		return nil, "", fmt.Errorf("LLM returned an empty response")
	}

	result, err := parseAnalysisJSON(rawOutput)
	if err != nil {
		return nil, rawOutput, err
	}
	success = true
	return result, rawOutput, nil
}

func (a *Analyzer) analyzeViaHFInference(ctx context.Context, settings RuntimeLLMSettings, userPrompt string) (*AnalysisResult, string, error) {
	inputs := buildHFInferenceInputs(settings.SystemPrompt, userPrompt)
	promptLength := len(inputs)

	start := time.Now()
	success := false
	defer func() {
		a.recordRequest(settings.Model, promptLength, time.Since(start).Milliseconds(), success)
	}()

	maxNewTokens := hfInferenceMaxNewTokens
	if settings.MaxTokens > 0 && settings.MaxTokens < maxNewTokens {
		maxNewTokens = settings.MaxTokens
	}

	payload := hfInferenceRequest{
		Inputs: inputs,
		Parameters: hfInferenceParameters{
			MaxNewTokens: maxNewTokens,
			Temperature:  settings.Temperature,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("marshal HF inference request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("create HF inference request: %w", err)
	}
	a.setRequestHeaders(req)

	slog.Info("HF inference request",
		"url", a.baseURL,
		"body_preview", truncateForLog(string(body), 200),
	)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("LLM request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, "", fmt.Errorf("read LLM response: %w", err)
	}

	slog.Info("HF inference response",
		"status", resp.StatusCode,
		"body_preview", lastNForLog(string(respBody), 1500),
	)

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("LLM returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	rawOutput, err := parseHFInferenceResponse(respBody)
	if err != nil {
		return nil, "", err
	}

	result, err := parseAnalysisJSON(rawOutput)
	if err != nil {
		slog.Info("HF parse failed, full extracted output",
			"error", err.Error(),
			"output", rawOutput,
		)
		return nil, rawOutput, err
	}
	success = true
	return result, rawOutput, nil
}

func (a *Analyzer) setRequestHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if a.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.apiKey)
	}
}

func (a *Analyzer) doRequest(req *http.Request) ([]byte, error) {
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read LLM response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return respBody, nil
}

func buildHFInferenceInputs(systemPrompt, userPrompt string) string {
	return fmt.Sprintf(
		"<start_of_turn>user\n%s\n\n%s<end_of_turn>\n%s",
		systemPrompt,
		userPrompt,
		hfModelTurnPrefix,
	)
}

func parseHFInferenceResponse(respBody []byte) (string, error) {
	generatedText, err := decodeHFGeneratedText(respBody)
	if err != nil {
		return "", err
	}

	rawOutput := extractHFModelOutput(generatedText)
	if rawOutput == "" {
		return "", fmt.Errorf("HF inference returned an empty generated_text")
	}
	return rawOutput, nil
}

func decodeHFGeneratedText(respBody []byte) (string, error) {
	var items []hfInferenceResponseItem
	if err := json.Unmarshal(respBody, &items); err == nil {
		if len(items) == 0 {
			return "", fmt.Errorf("HF inference returned an empty response array")
		}
		return items[0].GeneratedText, nil
	}

	var item hfInferenceResponseItem
	if err := json.Unmarshal(respBody, &item); err != nil {
		return "", fmt.Errorf("decode HF inference response: %w", err)
	}
	if strings.TrimSpace(item.GeneratedText) == "" {
		return "", fmt.Errorf("HF inference returned an empty generated_text")
	}
	return item.GeneratedText, nil
}

func truncateForLog(value string, maxLen int) string {
	if maxLen <= 0 || len(value) <= maxLen {
		return value
	}
	return value[:maxLen] + "..."
}

func lastNForLog(value string, maxLen int) string {
	if maxLen <= 0 || len(value) <= maxLen {
		return value
	}
	return "..." + value[len(value)-maxLen:]
}

func extractHFModelOutput(generatedText string) string {
	text := strings.TrimSpace(generatedText)
	if text == "" {
		return ""
	}
	if idx := strings.Index(text, hfModelTurnPrefix); idx >= 0 {
		return strings.TrimSpace(text[idx+len(hfModelTurnPrefix):])
	}
	return text
}

func (a *Analyzer) recordRequest(model string, promptLength int, durationMs int64, success bool) {
	if a == nil || a.requestLog == nil {
		return
	}
	if model == "" {
		model = a.model
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := a.requestLog.Insert(ctx, model, promptLength, durationMs, success); err != nil {
			slog.Default().Warn("failed to record llm request log",
				"error", err,
				"model", model,
				"success", success,
			)
		}
	}()
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type hfInferenceRequest struct {
	Inputs     string                 `json:"inputs"`
	Parameters hfInferenceParameters  `json:"parameters"`
}

type hfInferenceParameters struct {
	MaxNewTokens int     `json:"max_new_tokens"`
	Temperature  float64 `json:"temperature"`
}

type hfInferenceResponseItem struct {
	GeneratedText string `json:"generated_text"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (r chatCompletionResponse) firstContent() string {
	if len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[0].Message.Content
}

func buildPrompt(name, platform, notes string) string {
	trimmedNotes := strings.TrimSpace(notes)
	if trimmedNotes == "" {
		trimmedNotes = "No notes provided"
	}
	return fmt.Sprintf(`You are an influencer marketing analyst.

CRITICAL OUTPUT RULES:
- ONLY return valid JSON
- NO markdown
- NO explanation
- NO code fences
- NO text before or after the JSON object
- "insights" MUST be a string[] (plain text array), NOT an array of objects
  WRONG: [{"insights": "text"}] or [{"text": "..."}]
  CORRECT: ["Average 4.1%% engagement over the last 30 days", "Audience matches the target demographic"]

Example input:
Influencer: Jane Smith
Platform: instagram
Notes: Beauty & lifestyle niche

Example output:
%s

Now analyze this influencer and return JSON in the exact same format:

Influencer: %s
Platform: %s
Notes: %s`, fewShotExample, name, platform, trimmedNotes)
}

var jsonObjectPattern = regexp.MustCompile(`\{[\s\S]*\}`)

// extractFirstJSONObject returns the first balanced {...} block in text.
// Trailing model noise (repeated JSON, explanations) after the closing brace is ignored.
func extractFirstJSONObject(text string) string {
	start := strings.Index(text, "{")
	if start < 0 {
		return ""
	}

	depth := 0
	inString := false
	escape := false
	for i := start; i < len(text); i++ {
		c := text[i]
		if inString {
			if escape {
				escape = false
				continue
			}
			if c == '\\' {
				escape = true
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}

		switch c {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}
	return ""
}

func parseAnalysisJSON(raw string) (*AnalysisResult, error) {
	for _, candidate := range collectJSONCandidates(raw) {
		var parsed map[string]any
		if err := json.Unmarshal([]byte(candidate), &parsed); err != nil {
			continue
		}
		return normalizeParsed(parsed)
	}
	return nil, fmt.Errorf("model did not return valid JSON")
}

func collectJSONCandidates(raw string) []string {
	stripped := stripCodeFences(raw)
	firstObject := extractFirstJSONObject(stripped)
	if firstObject == "" {
		firstObject = extractFirstJSONObject(raw)
	}
	// Legacy greedy match fallback (e.g. deeply nested edge cases).
	greedyObject := extractJSONObjectGreedy(stripped)
	if greedyObject == "" {
		greedyObject = extractJSONObjectGreedy(raw)
	}

	seen := make(map[string]struct{})
	var out []string
	for _, c := range []string{firstObject, stripped, greedyObject, strings.TrimSpace(raw)} {
		if c == "" {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	return out
}

func stripCodeFences(text string) string {
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`(?i)^\s*`+"```"+`(?:json)?\s*`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`(?i)\s*`+"```"+`\s*$`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`(?i)`+"```"+`(?:json)?`).ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}

func extractJSONObjectGreedy(text string) string {
	return jsonObjectPattern.FindString(text)
}

func normalizeParsed(parsed map[string]any) (*AnalysisResult, error) {
	insights := normalizeInsights(parsed["insights"])
	summary := strings.TrimSpace(fmt.Sprint(parsed["summary"]))
	if summary == "" || summary == "<nil>" {
		return nil, fmt.Errorf("model returned an empty summary")
	}
	if len(insights) == 0 {
		return nil, fmt.Errorf("model returned no insights")
	}

	return &AnalysisResult{
		OverallScore:    clampScore(parsed["overall_score"]),
		EngagementScore: clampScore(parsed["engagement_score"]),
		AudienceScore:   clampScore(parsed["audience_score"]),
		BrandFitScore:   clampScore(parsed["brand_fit_score"]),
		Summary:         summary,
		Insights:        insights,
	}, nil
}

func clampScore(value any) float64 {
	switch v := value.(type) {
	case float64:
		return roundScore(v)
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return 0
		}
		return roundScore(f)
	default:
		var f float64
		if _, err := fmt.Sscan(fmt.Sprint(value), &f); err != nil {
			return 0
		}
		return roundScore(f)
	}
}

func roundScore(n float64) float64 {
	if math.IsNaN(n) {
		return 0
	}
	n = math.Min(100, math.Max(0, n))
	return math.Round(n*10) / 10
}

func normalizeInsights(raw any) []string {
	switch v := raw.(type) {
	case nil:
		return nil
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return nil
		}
		if strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{") {
			var parsed any
			if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
				return normalizeInsights(parsed)
			}
		}
		var out []string
		for _, line := range strings.Split(trimmed, "\n") {
			if text := normalizeInsightItem(line); text != "" {
				out = append(out, text)
			}
		}
		return out
	case []any:
		var out []string
		for _, item := range v {
			if text := normalizeInsightItem(item); text != "" {
				out = append(out, text)
			}
		}
		return out
	case map[string]any:
		if text := extractTextFromInsightObject(v); text != "" {
			return []string{text}
		}
		return nil
	default:
		if text := normalizeInsightItem(v); text != "" {
			return []string{text}
		}
		return nil
	}
}

func normalizeInsightItem(item any) string {
	switch v := item.(type) {
	case nil:
		return ""
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return ""
		}
		if strings.HasPrefix(trimmed, "{") {
			var obj map[string]any
			if err := json.Unmarshal([]byte(trimmed), &obj); err == nil {
				if text := extractTextFromInsightObject(obj); text != "" {
					return text
				}
			}
		}
		trimmed = strings.Trim(trimmed, `"'`)
		trimmed = regexp.MustCompile(`^[-•*]\s*`).ReplaceAllString(trimmed, "")
		return strings.TrimSpace(trimmed)
	case map[string]any:
		return extractTextFromInsightObject(v)
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func extractTextFromInsightObject(obj map[string]any) string {
	for _, key := range []string{"insights", "text", "value", "insight", "content", "message"} {
		if val, ok := obj[key]; ok {
			if text, ok := val.(string); ok && strings.TrimSpace(text) != "" {
				return strings.Trim(strings.TrimSpace(text), `"'`)
			}
		}
	}
	if len(obj) == 1 {
		for _, val := range obj {
			if text, ok := val.(string); ok && strings.TrimSpace(text) != "" {
				return strings.Trim(strings.TrimSpace(text), `"'`)
			}
		}
	}
	return ""
}
