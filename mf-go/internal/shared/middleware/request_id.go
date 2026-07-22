package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/shared/logger"
)

const RequestIDHeader = "X-Request-ID"

// RequestID adds a unique request ID to every request context and response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := logger.ContextWithRequestID(r.Context(), requestID)
		w.Header().Set(RequestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
