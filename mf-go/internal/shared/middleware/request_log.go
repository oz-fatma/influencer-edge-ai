package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// RequestLogWriter persists HTTP request metadata.
type RequestLogWriter interface {
	Insert(ctx context.Context, method, path string, statusCode int, durationMs int64) error
}

type requestLogResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *requestLogResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// RequestLogging records each HTTP request to the configured writer asynchronously.
func RequestLogging(writer RequestLogWriter) func(http.Handler) http.Handler {
	if writer == nil {
		return func(next http.Handler) http.Handler { return next }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := &requestLogResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			method := r.Method
			path := routePattern(r)
			statusCode := wrapped.statusCode
			durationMs := time.Since(start).Milliseconds()

			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				if err := writer.Insert(ctx, method, path, statusCode, durationMs); err != nil {
					slog.Default().Warn("failed to record request log",
						"error", err,
						"method", method,
						"path", path,
						"status", statusCode,
					)
				}
			}()
		})
	}
}

func routePattern(r *http.Request) string {
	if rc := chi.RouteContext(r.Context()); rc != nil {
		if pattern := rc.RoutePattern(); pattern != "" {
			return pattern
		}
	}
	return r.URL.Path
}
