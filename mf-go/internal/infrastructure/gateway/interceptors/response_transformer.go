package interceptors

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ResponseTransformer transforms response headers, status code, and body.
type ResponseTransformer struct {
	headerTransforms map[string]string // Header name -> value or template
	statusTransform  func(int) int     // Status code transformer
	bodyTransform    func([]byte) ([]byte, error)
}

// NewResponseTransformer creates a new response transformer.
func NewResponseTransformer(
	headerTransforms map[string]string,
	statusTransform func(int) int,
	bodyTransform func([]byte) ([]byte, error),
) *ResponseTransformer {
	return &ResponseTransformer{
		headerTransforms: headerTransforms,
		statusTransform:  statusTransform,
		bodyTransform:    bodyTransform,
	}
}

// InterceptRequest is a no-op for response transformer.
func (rt *ResponseTransformer) InterceptRequest(ctx context.Context, req *http.Request) (*http.Request, error) {
	return req, nil
}

// InterceptResponse applies transformations to the response.
func (rt *ResponseTransformer) InterceptResponse(ctx context.Context, req *http.Request, resp *http.Response) (*http.Response, error) {
	// Transform headers
	for name, value := range rt.headerTransforms {
		transformedValue := rt.transformTemplate(value, req, resp)
		resp.Header.Set(name, transformedValue)
	}

	// Transform status code
	if rt.statusTransform != nil {
		resp.StatusCode = rt.statusTransform(resp.StatusCode)
		resp.Status = fmt.Sprintf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	// Transform body if function provided
	if rt.bodyTransform != nil && resp.Body != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		defer resp.Body.Close()

		if len(body) > 0 {
			transformedBody, err := rt.bodyTransform(body)
			if err != nil {
				return nil, fmt.Errorf("body transformation failed: %w", err)
			}
			resp.Body = io.NopCloser(bytes.NewReader(transformedBody))
			resp.ContentLength = int64(len(transformedBody))
		}
	}

	return resp, nil
}

// transformTemplate replaces template variables.
func (rt *ResponseTransformer) transformTemplate(template string, req *http.Request, resp *http.Response) string {
	result := template
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

	// Add response-specific variables
	result = strings.ReplaceAll(result, "{status_code}", fmt.Sprintf("%d", resp.StatusCode))

	return result
}
