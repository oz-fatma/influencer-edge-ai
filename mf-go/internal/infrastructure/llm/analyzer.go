package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strings"

	"github.com/masterfabric-go/masterfabric/internal/shared/config"
)

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

// Analyzer calls an OpenAI-compatible MLC-LLM server.
type Analyzer struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewAnalyzer(cfg config.LLMConfig) *Analyzer {
	if cfg.BaseURL == "" {
		return nil
	}
	return &Analyzer{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		model:   cfg.Model,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (a *Analyzer) Model() string {
	return a.model
}

func (a *Analyzer) Analyze(ctx context.Context, name, platform, notes string) (*AnalysisResult, string, error) {
	payload := chatCompletionRequest{
		Model: a.model,
		Messages: []chatMessage{
			{
				Role:    "system",
				Content: "You are an expert influencer marketing analyst. ONLY return valid JSON. No markdown, no explanation, no code fences.",
			},
			{
				Role:    "user",
				Content: buildPrompt(name, platform, notes),
			},
		},
		Temperature: 0.1,
		MaxTokens:   700,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("create chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("LLM request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, "", fmt.Errorf("read LLM response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("LLM returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
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
	return result, rawOutput, nil
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
	fromBraces := extractJSONObject(stripped)
	if fromBraces == "" {
		fromBraces = extractJSONObject(raw)
	}

	seen := make(map[string]struct{})
	var out []string
	for _, c := range []string{stripped, fromBraces, strings.TrimSpace(raw)} {
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

func extractJSONObject(text string) string {
	match := jsonObjectPattern.FindString(text)
	return match
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
