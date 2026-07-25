package middleware

import (
	"net/http"

	adminRepo "github.com/masterfabric-go/masterfabric/internal/domain/admin/repository"
	"github.com/masterfabric-go/masterfabric/internal/shared/response"
)

// RequireAdmin ensures the authenticated user has the platform admin role.
func RequireAdmin(adminRepo adminRepo.AdminRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
				return
			}

			isAdmin, err := adminRepo.IsAdmin(r.Context(), userID)
			if err != nil {
				response.Error(w, err)
				return
			}
			if !isAdmin {
				response.JSON(w, http.StatusForbidden, map[string]string{"error": "admin access required"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
