package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/model"
	"github.com/masterfabric-go/masterfabric/internal/shared/middleware"
)

// DynamicHandlerResolver resolves handlers dynamically based on endpoint configuration.
// It supports multiple strategies:
// 1. Registered handlers (if service is registered)
// 2. HTTP proxy to external services (if backend_service is a URL)
// 3. Generic dynamic database handler (automatically performs CRUD operations)
type DynamicHandlerResolver struct {
	registry      *BackendRegistry
	httpClient    *http.Client
	logger        *slog.Logger
	db            *pgxpool.Pool
	serviceConfig map[string]ServiceConfig // Maps service name to configuration
}

// ServiceConfig holds configuration for a backend service.
type ServiceConfig struct {
	BaseURL string            `json:"base_url,omitempty"` // For HTTP proxy
	Headers map[string]string `json:"headers,omitempty"`  // Default headers to add
	Timeout int               `json:"timeout,omitempty"`  // Request timeout in seconds
}

const maxProxyResponseBytes = 1 << 20 // 1 MiB

// NewDynamicHandlerResolver creates a new dynamic handler resolver.
func NewDynamicHandlerResolver(registry *BackendRegistry, logger *slog.Logger, db *pgxpool.Pool) *DynamicHandlerResolver {
	if registry == nil {
		registry = NewBackendRegistry()
	}
	return &DynamicHandlerResolver{
		registry: registry,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		logger:        logger,
		db:            db,
		serviceConfig: make(map[string]ServiceConfig),
	}
}

// RegisterServiceConfig registers configuration for a backend service.
// This allows services to be proxied to external URLs dynamically.
func (r *DynamicHandlerResolver) RegisterServiceConfig(serviceName string, config ServiceConfig) {
	if r.serviceConfig == nil {
		r.serviceConfig = make(map[string]ServiceConfig)
	}
	r.serviceConfig[serviceName] = config
}

// Handle implements HandlerResolver interface for compatibility.
func (r *DynamicHandlerResolver) Handle(ctx context.Context, endpoint *model.Endpoint, req *http.Request) (*http.Response, error) {
	return r.ResolveHandler(ctx, endpoint, req)
}

// IsRegistered checks if a service has a registered handler or configuration.
func (r *DynamicHandlerResolver) IsRegistered(serviceName string) bool {
	// Check if handler is registered
	if r.registry != nil && r.registry.IsRegistered(serviceName) {
		return true
	}
	// Check if service has configuration
	if _, hasConfig := r.serviceConfig[serviceName]; hasConfig {
		return true
	}
	// Check if service name is a URL
	return isURL(serviceName)
}

// ResolveHandler resolves a handler for the given endpoint dynamically.
// It tries multiple strategies in order:
// 1. Check if handler is registered
// 2. Check if backend_service is a URL or has service config
// 3. Use generic dynamic handler
func (r *DynamicHandlerResolver) ResolveHandler(ctx context.Context, endpoint *model.Endpoint, req *http.Request) (*http.Response, error) {
	serviceName := endpoint.BackendService

	// Strategy 1: Check if handler is registered
	if r.registry != nil && r.registry.IsRegistered(serviceName) {
		return r.registry.Handle(ctx, endpoint, req)
	}

	// Strategy 2: Check if backend_service is a URL or has service config
	if config, hasConfig := r.serviceConfig[serviceName]; hasConfig && config.BaseURL != "" {
		return r.handleHTTPProxy(ctx, endpoint, req, config)
	}

	// Check if backend_service looks like a URL
	if isURL(serviceName) {
		config := ServiceConfig{BaseURL: serviceName}
		return r.handleHTTPProxy(ctx, endpoint, req, config)
	}

	// Strategy 3: Use generic dynamic handler
	return r.handleGeneric(ctx, endpoint, req)
}

// handleHTTPProxy proxies the request to an external HTTP service.
func (r *DynamicHandlerResolver) handleHTTPProxy(ctx context.Context, endpoint *model.Endpoint, req *http.Request, config ServiceConfig) (*http.Response, error) {
	// Build target URL
	targetURL, err := r.buildTargetURL(config.BaseURL, endpoint, req)
	if err != nil {
		return nil, fmt.Errorf("failed to build target URL: %w", err)
	}

	// Create proxy request
	proxyReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL, req.Body) // #nosec G704 -- admin-configured outbound proxy
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy request: %w", err)
	}

	// Copy headers from original request
	for k, v := range req.Header {
		proxyReq.Header[k] = v
	}

	// Add service-specific headers
	for k, v := range config.Headers {
		proxyReq.Header.Set(k, v)
	}

	// Remove gateway-specific headers that shouldn't be forwarded
	proxyReq.Header.Del("X-App-ID")
	proxyReq.Header.Del("X-Organization-ID")

	// Execute proxy request
	if r.logger != nil {
		r.logger.Info("proxying request",
			"service", endpoint.BackendService,
			"method", req.Method,
			"target", targetURL,
		)
	}

	resp, err := r.httpClient.Do(proxyReq) // #nosec G704 -- dispatch to validated operator backend
	if err != nil {
		return nil, fmt.Errorf("proxy request failed: %w", err)
	}

	if resp.Body != nil {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxProxyResponseBytes))
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("failed to read proxy response: %w", readErr)
		}
		resp.Body = io.NopCloser(bytes.NewReader(body))
		resp.ContentLength = int64(len(body))
	}

	return resp, nil
}

// buildTargetURL constructs the target URL for proxying.
func (r *DynamicHandlerResolver) buildTargetURL(baseURL string, endpoint *model.Endpoint, req *http.Request) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	// Use endpoint path or original request path
	path := endpoint.Path
	if path == "" {
		path = req.URL.Path
		// Remove /api/v1 prefix if present
		if strings.HasPrefix(path, "/api/v1") {
			path = strings.TrimPrefix(path, "/api/v1")
		}
	}

	// Build full URL
	targetURL := base.ResolveReference(&url.URL{Path: path})
	
	// Preserve query parameters
	targetURL.RawQuery = req.URL.RawQuery

	return targetURL.String(), nil
}

// deriveTableName dynamically derives table name from backend_service.
// Examples: "order-service" -> "orders", "product-service" -> "products"
// Also checks Schema JSON for optional "table_name" metadata.
func (r *DynamicHandlerResolver) deriveTableName(endpoint *model.Endpoint) string {
	// First, check if table_name is specified in Schema metadata
	if len(endpoint.Schema) > 0 {
		var schemaData map[string]interface{}
		if err := json.Unmarshal(endpoint.Schema, &schemaData); err == nil {
			if tableName, ok := schemaData["table_name"].(string); ok && tableName != "" {
				return tableName
			}
		}
	}

	// Derive from backend_service name dynamically
	serviceName := endpoint.BackendService
	
	// Remove common suffixes: "-service", "-api", "-handler"
	serviceName = strings.TrimSuffix(serviceName, "-service")
	serviceName = strings.TrimSuffix(serviceName, "-api")
	serviceName = strings.TrimSuffix(serviceName, "-handler")
	
	// Simple pluralization: add 's' if not ending in 's'
	if !strings.HasSuffix(serviceName, "s") {
		serviceName = serviceName + "s"
	}
	
	return serviceName
}

// handleGeneric handles requests using automatic database operations.
func (r *DynamicHandlerResolver) handleGeneric(ctx context.Context, endpoint *model.Endpoint, req *http.Request) (*http.Response, error) {
	// Get organization and app IDs from context
	orgID, ok := middleware.OrgIDFromContext(ctx)
	if !ok {
		return r.errorResponse(http.StatusBadRequest, "organization context required"), nil
	}

	appIDStr := req.Header.Get("X-App-ID")
	if appIDStr == "" {
		return r.errorResponse(http.StatusBadRequest, "app context required"), nil
	}
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		return r.errorResponse(http.StatusBadRequest, "invalid app ID"), nil
	}

	// Derive table name dynamically from backend_service
	tableName := r.deriveTableName(endpoint)
	if tableName == "" {
		return r.errorResponse(http.StatusBadRequest, fmt.Sprintf("cannot determine table for service: %s", endpoint.BackendService)), nil
	}

	// Check if database is available
	if r.db == nil {
		return r.errorResponse(http.StatusInternalServerError, "database not available"), nil
	}

	// Parse request body if present
	var requestData map[string]interface{}
	if req.Body != nil && req.Method != "GET" && req.Method != "DELETE" {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil && len(bodyBytes) > 0 {
			_ = json.Unmarshal(bodyBytes, &requestData)
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	// Process based on backend_action
	var responseData map[string]interface{}
	var statusCode int
	var dbErr error

	switch strings.ToLower(endpoint.BackendAction) {
	case "list":
		responseData, statusCode, dbErr = r.handleList(ctx, tableName, orgID, appID, req)
	case "get":
		responseData, statusCode, dbErr = r.handleGet(ctx, tableName, orgID, appID, req)
	case "create":
		responseData, statusCode, dbErr = r.handleCreate(ctx, tableName, orgID, appID, requestData)
	case "update", "patch":
		responseData, statusCode, dbErr = r.handleUpdate(ctx, tableName, orgID, appID, req, requestData)
	case "delete":
		responseData, statusCode, dbErr = r.handleDelete(ctx, tableName, orgID, appID, req)
	default:
		return r.errorResponse(http.StatusBadRequest, fmt.Sprintf("unknown action: %s", endpoint.BackendAction)), nil
	}

	if dbErr != nil {
		if r.logger != nil {
			r.logger.Error("database operation failed",
				"error", dbErr,
				"table", tableName,
				"action", endpoint.BackendAction,
				"service", endpoint.BackendService,
			)
		}
		return r.errorResponse(http.StatusInternalServerError, "an internal error occurred"), nil
	}

	body, _ := json.Marshal(responseData)
	resp := &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
	resp.Header.Set("Content-Type", "application/json")

	if r.logger != nil {
		r.logger.Info("handled request with dynamic database handler",
			"service", endpoint.BackendService,
			"table", tableName,
			"action", endpoint.BackendAction,
			"method", req.Method,
		)
	}

	return resp, nil
}

// handleList performs SELECT query with pagination.
func (r *DynamicHandlerResolver) handleList(ctx context.Context, tableName string, orgID, appID uuid.UUID, req *http.Request) (map[string]interface{}, int, error) {
	offset, _ := strconv.Atoi(req.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Count total (with multi-tenant isolation)
	var total int
	err := r.db.QueryRow(ctx,
		fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE organization_id = $1 AND app_id = $2`, tableName),
		orgID, appID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count records: %w", err)
	}

	// Query records (with multi-tenant isolation) using JSON aggregation
	rows, err := r.db.Query(ctx,
		fmt.Sprintf(`SELECT row_to_json(t.*) FROM %s t WHERE organization_id = $1 AND app_id = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`, tableName),
		orgID, appID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query records: %w", err)
	}
	defer rows.Close()

	// Convert rows to JSON
	var records []map[string]interface{}
	for rows.Next() {
		var jsonData []byte
		if err := rows.Scan(&jsonData); err != nil {
			return nil, 0, fmt.Errorf("failed to scan row: %w", err)
		}
		var record map[string]interface{}
		if err := json.Unmarshal(jsonData, &record); err != nil {
			return nil, 0, fmt.Errorf("failed to parse record: %w", err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return map[string]interface{}{
		"data":  records,
		"count": len(records),
		"total": total,
		"offset": offset,
		"limit": limit,
	}, http.StatusOK, nil
}

// handleGet performs SELECT by ID.
func (r *DynamicHandlerResolver) handleGet(ctx context.Context, tableName string, orgID, appID uuid.UUID, req *http.Request) (map[string]interface{}, int, error) {
	// Extract ID from path (e.g., /orders/{id})
	recordID, err := r.extractIDFromPath(req.URL.Path, tableName)
	if err != nil {
		return nil, 0, err
	}

	// Use JSON aggregation to get field names automatically
	var jsonData []byte
	err = r.db.QueryRow(ctx,
		fmt.Sprintf(`SELECT row_to_json(t.*) FROM %s t WHERE id = $1 AND organization_id = $2 AND app_id = $3`, tableName),
		recordID, orgID, appID,
	).Scan(&jsonData)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, fmt.Errorf("record not found")
		}
		return nil, 0, fmt.Errorf("failed to get record: %w", err)
	}

	var record map[string]interface{}
	if err := json.Unmarshal(jsonData, &record); err != nil {
		return nil, 0, fmt.Errorf("failed to parse record: %w", err)
	}

	return map[string]interface{}{"data": record}, http.StatusOK, nil
}

// handleCreate performs INSERT operation.
func (r *DynamicHandlerResolver) handleCreate(ctx context.Context, tableName string, orgID, appID uuid.UUID, input map[string]interface{}) (map[string]interface{}, int, error) {
	recordID := uuid.New()
	now := time.Now().UTC()

	// Build dynamic INSERT query
	columns := []string{"id", "organization_id", "app_id", "created_at", "updated_at"}
	values := []interface{}{recordID, orgID, appID, now, now}
	placeholders := []string{"$1", "$2", "$3", "$4", "$5"}
	paramIndex := 6

	for key, value := range input {
		if key != "id" && key != "organization_id" && key != "app_id" && key != "created_at" && key != "updated_at" {
			columns = append(columns, key)
			values = append(values, value)
			placeholders = append(placeholders, fmt.Sprintf("$%d", paramIndex))
			paramIndex++
		}
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s) RETURNING row_to_json(%s.*)`,
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
		tableName,
	)

	var jsonData []byte
	err := r.db.QueryRow(ctx, query, values...).Scan(&jsonData)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create record: %w", err)
	}

	var record map[string]interface{}
	if err := json.Unmarshal(jsonData, &record); err != nil {
		return nil, 0, fmt.Errorf("failed to parse record: %w", err)
	}

	return map[string]interface{}{"data": record}, http.StatusCreated, nil
}

// handleUpdate performs UPDATE operation.
func (r *DynamicHandlerResolver) handleUpdate(ctx context.Context, tableName string, orgID, appID uuid.UUID, req *http.Request, input map[string]interface{}) (map[string]interface{}, int, error) {
	recordID, err := r.extractIDFromPath(req.URL.Path, tableName)
	if err != nil {
		return nil, 0, err
	}

	// Build dynamic UPDATE query
	setParts := []string{"updated_at = $1"}
	values := []interface{}{time.Now().UTC()}
	paramIndex := 2

	for key, value := range input {
		if key != "id" && key != "organization_id" && key != "app_id" && key != "created_at" && key != "updated_at" {
			setParts = append(setParts, fmt.Sprintf("%s = $%d", key, paramIndex))
			values = append(values, value)
			paramIndex++
		}
	}

	values = append(values, recordID, orgID, appID)
	query := fmt.Sprintf(
		`UPDATE %s SET %s WHERE id = $%d AND organization_id = $%d AND app_id = $%d RETURNING row_to_json(%s.*)`,
		tableName,
		strings.Join(setParts, ", "),
		paramIndex, paramIndex+1, paramIndex+2,
		tableName,
	)

	var jsonData []byte
	err = r.db.QueryRow(ctx, query, values...).Scan(&jsonData)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, fmt.Errorf("record not found")
		}
		return nil, 0, fmt.Errorf("failed to update record: %w", err)
	}

	var record map[string]interface{}
	if err := json.Unmarshal(jsonData, &record); err != nil {
		return nil, 0, fmt.Errorf("failed to parse record: %w", err)
	}

	return map[string]interface{}{"data": record}, http.StatusOK, nil
}

// handleDelete performs DELETE operation.
func (r *DynamicHandlerResolver) handleDelete(ctx context.Context, tableName string, orgID, appID uuid.UUID, req *http.Request) (map[string]interface{}, int, error) {
	recordID, err := r.extractIDFromPath(req.URL.Path, tableName)
	if err != nil {
		return nil, 0, err
	}

	result, err := r.db.Exec(ctx,
		fmt.Sprintf(`DELETE FROM %s WHERE id = $1 AND organization_id = $2 AND app_id = $3`, tableName),
		recordID, orgID, appID,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to delete record: %w", err)
	}

	if result.RowsAffected() == 0 {
		return nil, 0, fmt.Errorf("record not found")
	}

	return map[string]interface{}{"message": "Record deleted successfully"}, http.StatusOK, nil
}

// Helper functions
func (r *DynamicHandlerResolver) extractIDFromPath(path, tableName string) (uuid.UUID, error) {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	singularName := strings.TrimSuffix(tableName, "s") // "orders" -> "order"
	
	for i, part := range pathParts {
		if i > 0 && strings.EqualFold(pathParts[i-1], singularName) {
			recordID, err := uuid.Parse(part)
			if err != nil {
				return uuid.Nil, fmt.Errorf("invalid ID format")
			}
			return recordID, nil
		}
	}
	return uuid.Nil, fmt.Errorf("record ID required in path")
}

func (r *DynamicHandlerResolver) errorResponse(statusCode int, message string) *http.Response {
	body, _ := json.Marshal(map[string]interface{}{"error": message})
	resp := &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
	resp.Header.Set("Content-Type", "application/json")
	return resp
}

// isURL checks if a string is a valid URL.
func isURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}
