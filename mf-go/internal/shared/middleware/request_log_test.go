package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

type mockRequestLogWriter struct {
	mu      sync.Mutex
	entries []requestLogEntry
}

type requestLogEntry struct {
	method     string
	path       string
	statusCode int
	durationMs int64
}

func (m *mockRequestLogWriter) Insert(
	_ context.Context,
	method, path string,
	statusCode int,
	durationMs int64,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, requestLogEntry{
		method:     method,
		path:       path,
		statusCode: statusCode,
		durationMs: durationMs,
	})
	return nil
}

func (m *mockRequestLogWriter) lastEntry() (requestLogEntry, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.entries) == 0 {
		return requestLogEntry{}, false
	}
	return m.entries[len(m.entries)-1], true
}

func TestRequestLogging_recordsRequestAsync(t *testing.T) {
	writer := &mockRequestLogWriter{}

	r := chi.NewRouter()
	r.Use(RequestLogging(writer))
	r.Get("/api/v1/scores", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/scores", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		entry, ok := writer.lastEntry()
		if ok {
			if entry.method != http.MethodGet {
				t.Fatalf("method = %q, want GET", entry.method)
			}
			if entry.path != "/api/v1/scores" {
				t.Fatalf("path = %q, want /api/v1/scores", entry.path)
			}
			if entry.statusCode != http.StatusOK {
				t.Fatalf("status = %d, want 200", entry.statusCode)
			}
			if entry.durationMs < 0 {
				t.Fatalf("duration_ms = %d, want >= 0", entry.durationMs)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for async request log insert")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestRequestLogging_noopWhenWriterNil(t *testing.T) {
	r := chi.NewRouter()
	r.Use(RequestLogging(nil))
	r.Get("/health/live", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
