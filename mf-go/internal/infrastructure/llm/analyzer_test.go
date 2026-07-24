package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
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
	if NewAnalyzer(config.LLMConfig{Timeout: time.Minute}, nil) != nil {
		t.Fatal("expected nil analyzer")
	}
}

type recordingLLMWriter struct {
	mu     sync.Mutex
	calls  []llmRequestRecord
}

type llmRequestRecord struct {
	model        string
	promptLength int
	durationMs   int64
	success      bool
}

func (w *recordingLLMWriter) Insert(_ context.Context, modelName string, promptLength int, durationMs int64, success bool) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.calls = append(w.calls, llmRequestRecord{
		model:        modelName,
		promptLength: promptLength,
		durationMs:   durationMs,
		success:      success,
	})
	return nil
}

func TestAnalyze_recordsSuccessfulLLMRequest(t *testing.T) {
	validJSON := `{"overall_score":80,"engagement_score":80,"audience_score":80,"brand_fit_score":80,"summary":"Good fit","insights":["Strong niche alignment"]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":` + strconvQuote(validJSON) + `}}]}`))
	}))
	defer srv.Close()

	writer := &recordingLLMWriter{}
	analyzer := NewAnalyzer(config.LLMConfig{
		BaseURL: srv.URL,
		Model:   "gemma2:2b",
		Timeout: 5 * time.Second,
	}, writer)

	_, _, err := analyzer.Analyze(context.Background(), "Ada", "instagram", "tech")
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		writer.mu.Lock()
		n := len(writer.calls)
		writer.mu.Unlock()
		if n > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("expected async llm request log")
		}
		time.Sleep(10 * time.Millisecond)
	}

	writer.mu.Lock()
	defer writer.mu.Unlock()
	if len(writer.calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(writer.calls))
	}
	call := writer.calls[0]
	if call.model != "gemma2:2b" {
		t.Fatalf("model = %q", call.model)
	}
	if call.promptLength <= 0 {
		t.Fatalf("promptLength = %d", call.promptLength)
	}
	if call.durationMs < 0 {
		t.Fatalf("durationMs = %d", call.durationMs)
	}
	if !call.success {
		t.Fatal("expected success=true")
	}
}

func TestAnalyze_recordsFailedLLMRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	writer := &recordingLLMWriter{}
	analyzer := NewAnalyzer(config.LLMConfig{
		BaseURL: srv.URL,
		Model:   "gemma2:2b",
		Timeout: 5 * time.Second,
	}, writer)

	_, _, err := analyzer.Analyze(context.Background(), "Ada", "instagram", "tech")
	if err == nil {
		t.Fatal("expected error")
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		writer.mu.Lock()
		n := len(writer.calls)
		writer.mu.Unlock()
		if n > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("expected async llm request log")
		}
		time.Sleep(10 * time.Millisecond)
	}

	writer.mu.Lock()
	defer writer.mu.Unlock()
	if !writer.calls[0].success {
		return
	}
	t.Fatal("expected success=false")
}

func strconvQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
