package interceptors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

// PIIMasker masks sensitive fields in request/response bodies.
type PIIMasker struct {
	maskPatterns []*regexp.Regexp // Patterns to match PII fields
	maskValue    string           // Value to replace with (default: "***")
}

// NewPIIMasker creates a new PII masker.
func NewPIIMasker(maskFields []string, maskValue string) *PIIMasker {
	if maskValue == "" {
		maskValue = "***"
	}

	patterns := make([]*regexp.Regexp, 0, len(maskFields))
	for _, field := range maskFields {
		// Case-insensitive pattern matching common PII field names
		pattern := regexp.MustCompile(fmt.Sprintf(`(?i)(%s)`, regexp.QuoteMeta(field)))
		patterns = append(patterns, pattern)
	}

	return &PIIMasker{
		maskPatterns: patterns,
		maskValue:    maskValue,
	}
}

// InterceptRequest masks PII in request body if PII masking is enabled.
func (p *PIIMasker) InterceptRequest(ctx context.Context, req *http.Request) (*http.Request, error) {
	// Check if PII masking is enabled for this endpoint
	maskPII, ok := ctx.Value("pii_masking").(bool)
	if !ok || !maskPII {
		return req, nil
	}

	// Only mask POST, PUT, PATCH requests with body
	if req.Method != http.MethodPost && req.Method != http.MethodPut && req.Method != http.MethodPatch {
		return req, nil
	}

	if req.Body == nil {
		return req, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	defer req.Body.Close()

	if len(body) == 0 {
		return req, nil
	}

	maskedBody, err := p.maskBody(body)
	if err != nil {
		return nil, fmt.Errorf("PII masking failed: %w", err)
	}

	req.Body = io.NopCloser(bytes.NewReader(maskedBody))
	req.ContentLength = int64(len(maskedBody))

	return req, nil
}

// InterceptResponse masks PII in response body if PII masking is enabled.
func (p *PIIMasker) InterceptResponse(ctx context.Context, req *http.Request, resp *http.Response) (*http.Response, error) {
	// Check if PII masking is enabled for this endpoint
	maskPII, ok := ctx.Value("pii_masking").(bool)
	if !ok || !maskPII {
		return resp, nil
	}

	if resp.Body == nil {
		return resp, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	defer resp.Body.Close()

	if len(body) == 0 {
		return resp, nil
	}

	maskedBody, err := p.maskBody(body)
	if err != nil {
		return nil, fmt.Errorf("PII masking failed: %w", err)
	}

	resp.Body = io.NopCloser(bytes.NewReader(maskedBody))
	resp.ContentLength = int64(len(maskedBody))

	return resp, nil
}

// maskBody masks PII fields in JSON body.
func (p *PIIMasker) maskBody(body []byte) ([]byte, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		// Not JSON, return as-is
		return body, nil
	}

	p.maskMap(data)

	return json.Marshal(data)
}

// maskMap recursively masks PII fields in a map.
func (p *PIIMasker) maskMap(m map[string]interface{}) {
	for key, value := range m {
		// Check if key matches any PII pattern
		for _, pattern := range p.maskPatterns {
			if pattern.MatchString(key) {
				m[key] = p.maskValue
				continue
			}
		}

		// Recursively process nested maps
		if nestedMap, ok := value.(map[string]interface{}); ok {
			p.maskMap(nestedMap)
		} else if nestedSlice, ok := value.([]interface{}); ok {
			for _, item := range nestedSlice {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					p.maskMap(nestedMap)
				}
			}
		}
	}
}
