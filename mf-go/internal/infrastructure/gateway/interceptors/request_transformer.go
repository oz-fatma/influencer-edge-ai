package interceptors

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// RequestTransformer transforms request headers, query params, and body.
type RequestTransformer struct {
	headerTransforms map[string]string // Header name -> value or template
	queryTransforms  map[string]string // Query param -> value or template
	bodyTransform    func([]byte) ([]byte, error)
}

// NewRequestTransformer creates a new request transformer.
func NewRequestTransformer(
	headerTransforms map[string]string,
	queryTransforms map[string]string,
	bodyTransform func([]byte) ([]byte, error),
) *RequestTransformer {
	return &RequestTransformer{
		headerTransforms: headerTransforms,
		queryTransforms:  queryTransforms,
		bodyTransform:    bodyTransform,
	}
}

// InterceptRequest applies transformations to the request.
func (rt *RequestTransformer) InterceptRequest(ctx context.Context, req *http.Request) (*http.Request, error) {
	// Transform headers
	for name, value := range rt.headerTransforms {
		transformedValue := rt.transformTemplate(value, req)
		req.Header.Set(name, transformedValue)
	}

	// Transform query parameters
	if len(rt.queryTransforms) > 0 {
		q := req.URL.Query()
		for key, value := range rt.queryTransforms {
			transformedValue := rt.transformTemplate(value, req)
			q.Set(key, transformedValue)
		}
		req.URL.RawQuery = q.Encode()
	}

	// Transform body if function provided
	if rt.bodyTransform != nil && req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		defer req.Body.Close()

		if len(body) > 0 {
			transformedBody, err := rt.bodyTransform(body)
			if err != nil {
				return nil, fmt.Errorf("body transformation failed: %w", err)
			}
			req.Body = io.NopCloser(bytes.NewReader(transformedBody))
			req.ContentLength = int64(len(transformedBody))
		}
	}

	return req, nil
}

// InterceptResponse is a no-op for request transformer.
func (rt *RequestTransformer) InterceptResponse(ctx context.Context, req *http.Request, resp *http.Response) (*http.Response, error) {
	return resp, nil
}

// transformTemplate replaces template variables like {org_id}, {user_id}, etc.
func (rt *RequestTransformer) transformTemplate(template string, req *http.Request) string {
	result := template
	// Simple template replacement - can be enhanced with proper templating library
	// For now, support basic variable substitution from context or headers
	ctx := req.Context()

	// Replace common variables
	if orgID, ok := ctx.Value("org_id").(string); ok {
		result = strings.ReplaceAll(result, "{org_id}", orgID)
	}
	if userID, ok := ctx.Value("user_id").(string); ok {
		result = strings.ReplaceAll(result, "{user_id}", userID)
	}
	if appID, ok := ctx.Value("app_id").(string); ok {
		result = strings.ReplaceAll(result, "{app_id}", appID)
	}

	return result
}
