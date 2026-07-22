package gateway

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/repository"
	gatewayDomain "github.com/masterfabric-go/masterfabric/internal/domain/gateway"
	iamService "github.com/masterfabric-go/masterfabric/internal/domain/iam/service"
	"github.com/masterfabric-go/masterfabric/internal/shared/middleware"
	"github.com/masterfabric-go/masterfabric/internal/shared/response"
	"github.com/redis/go-redis/v9"
)

// Pipeline is the gateway policy pipeline that enforces policies on managed endpoints.
type Pipeline struct {
	endpointRepo    repository.EndpointRepository
	policyRepo      repository.PolicyRepository
	rbac            iamService.RBACService
	redis           *redis.Client
	logger          *slog.Logger
	interceptors    *gatewayDomain.Chain
	backendRegistry HandlerResolver // Can be BackendRegistry or DynamicHandlerResolver
}

// HandlerResolver is an interface for resolving backend handlers.
// Both BackendRegistry and DynamicHandlerResolver implement this.
type HandlerResolver interface {
	Handle(ctx context.Context, endpoint *model.Endpoint, req *http.Request) (*http.Response, error)
	IsRegistered(serviceName string) bool
}

// NewPipeline creates a new gateway Pipeline.
func NewPipeline(
	endpointRepo repository.EndpointRepository,
	policyRepo repository.PolicyRepository,
	rbac iamService.RBACService,
	redisClient *redis.Client,
	logger *slog.Logger,
	backendRegistry HandlerResolver,
	interceptors ...gatewayDomain.Interceptor,
) *Pipeline {
	var chain *gatewayDomain.Chain
	if len(interceptors) > 0 {
		chain = gatewayDomain.NewChain(interceptors...)
	}
	if backendRegistry == nil {
		backendRegistry = NewBackendRegistry()
	}
	return &Pipeline{
		endpointRepo:    endpointRepo,
		policyRepo:      policyRepo,
		rbac:            rbac,
		redis:           redisClient,
		logger:          logger,
		interceptors:    chain,
		backendRegistry: backendRegistry,
	}
}

// Enforce is middleware that runs the full policy pipeline for managed endpoints.
// It performs: endpoint lookup -> auth policy check -> permission enforcement -> rate limiting.
func (p *Pipeline) Enforce(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Skip for non-managed paths (health, auth, admin)
		if shouldSkipPipeline(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// 1. Resolve app from context
		appIDStr := r.Header.Get("X-App-ID")
		if appIDStr == "" {
			next.ServeHTTP(w, r) // No app context, skip pipeline
			return
		}

		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid X-App-ID"})
			return
		}

		// 2. Endpoint lookup
		// Normalize path: strip /api/v1 prefix if present to match stored endpoint paths
		lookupPath := r.URL.Path
		if strings.HasPrefix(lookupPath, "/api/v1") {
			lookupPath = strings.TrimPrefix(lookupPath, "/api/v1")
			if lookupPath == "" {
				lookupPath = "/"
			}
		}

		endpoint, err := p.endpointRepo.GetByMethodPath(ctx, appID, r.Method, lookupPath, "v1")
		if err != nil {
			// No managed endpoint found, pass through
			next.ServeHTTP(w, r)
			return
		}

		if !endpoint.IsActive() {
			response.JSON(w, http.StatusGone, map[string]string{"error": "endpoint is not active"})
			return
		}

		// 3. Fetch endpoint policy
		policy, _ := p.policyRepo.GetByEndpointID(ctx, endpoint.ID)

		// 4. Permission enforcement
		if policy != nil && policy.RequiredPermission != "" {
			userID, ok := middleware.UserIDFromContext(ctx)
			if !ok {
				response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
				return
			}

			orgID, _ := middleware.OrgIDFromContext(ctx)
			has, err := p.rbac.HasPermission(ctx, userID, orgID, policy.RequiredPermission)
			if err != nil {
				p.logger.Error("permission check failed", "error", err)
				response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "permission check failed"})
				return
			}
			if !has {
				response.JSON(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions for this endpoint"})
				return
			}
		}

		// 5. Rate limiting
		if policy != nil && policy.RateLimit > 0 && p.redis != nil {
			if err := p.checkRateLimit(r, appID, endpoint.ID, policy.RateLimit); err != nil {
				response.JSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
				return
			}
		}

		// 6. Prepare context for interceptors
		ctx = context.WithValue(ctx, "endpoint_schema", endpoint.Schema)
		ctx = context.WithValue(ctx, "pii_masking", endpoint.PIIMasking)
		ctx = context.WithValue(ctx, "endpoint_id", endpoint.ID.String())
		ctx = context.WithValue(ctx, "app_id", appID.String())
		if orgID, ok := middleware.OrgIDFromContext(ctx); ok {
			ctx = context.WithValue(ctx, "org_id", orgID.String())
		}
		if userID, ok := middleware.UserIDFromContext(ctx); ok {
			ctx = context.WithValue(ctx, "user_id", userID.String())
		}
		r = r.WithContext(ctx)

		// 7. Apply request interceptors
		if p.interceptors != nil {
			interceptedReq, err := p.interceptors.InterceptRequest(ctx, r)
			if err != nil {
				p.logger.Error("request interceptor failed", "error", err)
				response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			r = interceptedReq
		}

		// 8. Route to backend handler (dynamic resolution)
		if p.backendRegistry != nil {
			backendResp, err := p.backendRegistry.Handle(ctx, endpoint, r)
			if err != nil {
				p.logger.Error("backend handler failed", "error", err, "service", endpoint.BackendService)
				response.JSON(w, http.StatusBadGateway, map[string]string{
					"error":   "backend service error",
					"message": err.Error(),
				})
				return
			}

			if backendResp != nil {
				defer backendResp.Body.Close()

				// Copy response headers
				for k, v := range backendResp.Header {
					for _, val := range v {
						w.Header().Add(k, val)
					}
				}

				// Copy status code and body
				w.WriteHeader(backendResp.StatusCode)
				if backendResp.Body != nil {
					_, _ = io.Copy(w, backendResp.Body)
				}
				return
			}
		}

		// Fallback: No backend handler registered, return endpoint info
		response.JSON(w, http.StatusOK, map[string]interface{}{
			"message":         "Endpoint validated successfully",
			"endpoint_id":     endpoint.ID.String(),
			"method":          endpoint.Method,
			"path":            endpoint.Path,
			"backend_service": endpoint.BackendService,
			"backend_action":  endpoint.BackendAction,
			"note":            fmt.Sprintf("No handler found for '%s'. Register a handler or configure HTTP proxy.", endpoint.BackendService),
		})
		return

		// 9. Apply response interceptors (if needed)
		// Note: Full response interception requires capturing the response,
		// which is more complex. For now, we handle it in middleware or
		// use a response wrapper.
	})
}

// responseWriter wraps http.ResponseWriter to capture response for interceptors.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	header     http.Header
	wroteHeader bool
}

func (rw *responseWriter) Header() http.Header {
	return rw.header
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.statusCode = code
	rw.wroteHeader = true
	for k, v := range rw.header {
		rw.ResponseWriter.Header()[k] = v
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// checkRateLimit uses a Redis sliding window counter.
func (p *Pipeline) checkRateLimit(r *http.Request, appID, endpointID uuid.UUID, limit int) error {
	ctx := r.Context()
	key := fmt.Sprintf("rate:%s:%s", appID, endpointID)

	count, err := p.redis.Incr(ctx, key).Result()
	if err != nil {
		return err
	}

	if count == 1 {
		// Set TTL on first request in the window (1 minute window)
		p.redis.Expire(ctx, key, 60_000_000_000) // 60 seconds in nanoseconds
	}

	if count > int64(limit) {
		return fmt.Errorf("rate limit exceeded")
	}

	return nil
}

// shouldSkipPipeline returns true for paths that bypass the gateway policy pipeline.
func shouldSkipPipeline(path string) bool {
	skipPrefixes := []string{"/health", "/api/v1/auth", "/api/v1/admin", "/api/v1/ws"}
	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
