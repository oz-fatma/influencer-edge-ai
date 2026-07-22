package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	tenantRepo "github.com/masterfabric-go/masterfabric/internal/domain/tenant/repository"
	"github.com/masterfabric-go/masterfabric/internal/shared/logger"
	"github.com/masterfabric-go/masterfabric/internal/shared/response"
)

const (
	ContextKeyTenantID    contextKey = "tenant_id"
	ContextKeyWorkspaceID contextKey = "workspace_id"
	ContextKeyAppID       contextKey = "tenant_app_id"
)

// TenantResolver resolves the tenant (organization) and optionally workspace from the request.
// Resolution order: X-Organization-ID header > JWT claims > subdomain.
// Workspace resolution: X-Workspace-ID header > X-Workspace-Slug header (requires org context).
func TenantResolver(orgRepo tenantRepo.OrgRepository) func(http.Handler) http.Handler {
	return TenantResolverWithWorkspace(orgRepo, nil)
}

// TenantResolverWithWorkspace resolves tenant and workspace from the request.
func TenantResolverWithWorkspace(orgRepo tenantRepo.OrgRepository, workspaceRepo tenantRepo.WorkspaceRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			var orgID uuid.UUID

			// 1. Check explicit header
			if header := r.Header.Get("X-Organization-ID"); header != "" {
				parsed, err := uuid.Parse(header)
				if err != nil {
					response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid X-Organization-ID"})
					return
				}
				orgID = parsed
			}

			// 2. Fall back to JWT claims (if auth middleware already ran)
			if orgID == uuid.Nil {
				if claimOrgID, ok := ctx.Value(ContextKeyOrganizationID).(uuid.UUID); ok && claimOrgID != uuid.Nil {
					orgID = claimOrgID
				}
			}

			// 3. Fall back to subdomain
			if orgID == uuid.Nil {
				host := r.Host
				parts := strings.Split(host, ".")
				if len(parts) > 2 {
					slug := parts[0]
					if orgRepo != nil {
						org, err := orgRepo.GetBySlug(ctx, slug)
						if err == nil && org != nil {
							orgID = org.ID
						}
					}
				}
			}

			if orgID != uuid.Nil {
				ctx = context.WithValue(ctx, ContextKeyTenantID, orgID)
				ctx = logger.ContextWithOrganizationID(ctx, orgID.String())

				// Resolve workspace if workspace repo is provided
				var workspaceID uuid.UUID
				if workspaceRepo != nil {
					// 1. Check explicit header
					if header := r.Header.Get("X-Workspace-ID"); header != "" {
						parsed, err := uuid.Parse(header)
						if err == nil {
							workspaceID = parsed
						}
					}

					// 2. Check workspace slug header (requires org context)
					if workspaceID == uuid.Nil {
						if slug := r.Header.Get("X-Workspace-Slug"); slug != "" {
							if ws, err := workspaceRepo.GetBySlug(ctx, orgID, slug); err == nil && ws != nil {
								workspaceID = ws.ID
							}
						}
					}

					if workspaceID != uuid.Nil {
						ctx = context.WithValue(ctx, ContextKeyWorkspaceID, workspaceID)
					}
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireTenant ensures a tenant ID is present in the context.
func RequireTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := TenantIDFromContext(r.Context()); !ok {
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": "tenant (organization) not resolved"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// TenantIDFromContext extracts the tenant ID from context.
func TenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ContextKeyTenantID).(uuid.UUID)
	return id, ok && id != uuid.Nil
}

// WorkspaceIDFromContext extracts the workspace ID from context.
func WorkspaceIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ContextKeyWorkspaceID).(uuid.UUID)
	return id, ok && id != uuid.Nil
}
