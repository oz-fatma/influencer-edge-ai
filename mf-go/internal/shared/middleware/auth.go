package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/iam/service"
	"github.com/masterfabric-go/masterfabric/internal/shared/logger"
	"github.com/masterfabric-go/masterfabric/internal/shared/response"
)

type contextKey string

const (
	ContextKeyUserID         contextKey = "auth_user_id"
	ContextKeyEmail          contextKey = "auth_email"
	ContextKeyOrganizationID contextKey = "auth_organization_id"
	ContextKeyPermissions    contextKey = "auth_permissions"
	ContextKeyClaims         contextKey = "auth_claims"
)

// JWTAuth is middleware that validates JWT tokens and injects claims into context.
func JWTAuth(authService service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid authorization format"})
				return
			}

			claims, err := authService.ValidateToken(r.Context(), parts[1])
			if err != nil {
				response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
				return
			}

			// Inject claims into context
			ctx := r.Context()
			ctx = context.WithValue(ctx, ContextKeyClaims, claims)
			ctx = context.WithValue(ctx, ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeyEmail, claims.Email)
			ctx = context.WithValue(ctx, ContextKeyOrganizationID, claims.OrganizationID)
			ctx = context.WithValue(ctx, ContextKeyPermissions, claims.Permissions)

			// Also populate logger context
			ctx = logger.ContextWithUserID(ctx, claims.UserID.String())
			if claims.OrganizationID != uuid.Nil {
				ctx = logger.ContextWithOrganizationID(ctx, claims.OrganizationID.String())
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission is middleware that checks if the authenticated user has a specific permission.
func RequirePermission(rbac service.RBACService, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(ContextKeyUserID).(uuid.UUID)
			if !ok {
				response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "user not authenticated"})
				return
			}

			orgID, _ := r.Context().Value(ContextKeyOrganizationID).(uuid.UUID)

			has, err := rbac.HasPermission(r.Context(), userID, orgID, permission)
			if err != nil {
				response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "permission check failed"})
				return
			}
			if !has {
				response.JSON(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// UserIDFromContext extracts the authenticated user ID from the context.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ContextKeyUserID).(uuid.UUID)
	return id, ok
}

// OrgIDFromContext extracts the organization ID from the context.
func OrgIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ContextKeyOrganizationID).(uuid.UUID)
	return id, ok
}
