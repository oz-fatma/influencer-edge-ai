package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/audit/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/audit/repository"
)

// AuditLog is middleware that records audit log entries for each request.
func AuditLog(auditRepo repository.AuditRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Serve the request first
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			// Record audit log asynchronously (best-effort)
			go func() {
				ctx := r.Context()

				orgID, _ := TenantIDFromContext(ctx)
				userID, _ := UserIDFromContext(ctx)

				// Get request ID from response header (set by RequestID middleware)
				requestID := w.Header().Get(RequestIDHeader)

				var userIDPtr *uuid.UUID
				if userID != uuid.Nil {
					userIDPtr = &userID
				}

				entry := &model.AuditLog{
					OrganizationID: orgID,
					UserID:         userIDPtr,
					RequestID:      requestID,
					Action:         r.Method + " " + r.URL.Path,
					ResourceType:   "http_request",
					ResourceID:     r.URL.Path,
					IPAddress:      r.RemoteAddr,
					UserAgent:      r.UserAgent(),
				}

				// Best-effort: ignore errors
				_ = auditRepo.Create(ctx, entry)
			}()
		})
	}
}
