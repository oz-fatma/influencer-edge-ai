package interceptors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/xeipuuv/gojsonschema"
)

// SchemaValidator validates request bodies against JSON schemas.
type SchemaValidator struct {
	schemaLoader *gojsonschema.SchemaLoader
}

// NewSchemaValidator creates a new schema validator interceptor.
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{
		schemaLoader: gojsonschema.NewSchemaLoader(),
	}
}

// InterceptRequest validates the request body against the provided schema.
func (s *SchemaValidator) InterceptRequest(ctx context.Context, req *http.Request) (*http.Request, error) {
	// Get schema from context (set by gateway pipeline)
	schemaJSON, ok := ctx.Value("endpoint_schema").([]byte)
	if !ok || len(schemaJSON) == 0 {
		// No schema defined, skip validation
		return req, nil
	}

	// Only validate POST, PUT, PATCH requests with body
	if req.Method != http.MethodPost && req.Method != http.MethodPut && req.Method != http.MethodPatch {
		return req, nil
	}

	// Read request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	defer req.Body.Close()

	if len(body) == 0 {
		// Empty body, skip validation
		return req, nil
	}

	// Parse schema
	var schema map[string]interface{}
	if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		return nil, fmt.Errorf("invalid schema JSON: %w", err)
	}

	// Validate
	schemaLoader := gojsonschema.NewGoLoader(schema)
	documentLoader := gojsonschema.NewBytesLoader(body)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, fmt.Errorf("schema validation error: %w", err)
	}

	if !result.Valid() {
		errors := make([]string, 0, len(result.Errors()))
		for _, desc := range result.Errors() {
			errors = append(errors, desc.String())
		}
		return nil, fmt.Errorf("validation failed: %v", errors)
	}

	// Restore body for downstream handlers
	req.Body = io.NopCloser(bytes.NewReader(body))

	return req, nil
}

// InterceptResponse is a no-op for schema validation.
func (s *SchemaValidator) InterceptResponse(ctx context.Context, req *http.Request, resp *http.Response) (*http.Response, error) {
	return resp, nil
}
