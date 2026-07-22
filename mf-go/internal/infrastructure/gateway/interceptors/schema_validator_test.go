package interceptors

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaValidator_InterceptRequest(t *testing.T) {
	validator := NewSchemaValidator()

	tests := []struct {
		name          string
		schema        []byte
		body          []byte
		shouldPass    bool
		expectedError string
	}{
		{
			name: "valid request with schema",
			schema: []byte(`{
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"age": {"type": "number"}
				},
				"required": ["name"]
			}`),
			body:       []byte(`{"name": "John", "age": 30}`),
			shouldPass: true,
		},
		{
			name: "invalid request - missing required field",
			schema: []byte(`{
				"type": "object",
				"properties": {
					"name": {"type": "string"}
				},
				"required": ["name"]
			}`),
			body:          []byte(`{"age": 30}`),
			shouldPass:    false,
			expectedError: "validation failed",
		},
		{
			name:          "no schema defined - should pass",
			schema:        nil,
			body:          []byte(`{"name": "John"}`),
			shouldPass:    true,
		},
		{
			name:          "empty body - should pass",
			schema:        []byte(`{"type": "object"}`),
			body:          []byte{},
			shouldPass:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.schema != nil {
				ctx = context.WithValue(ctx, "endpoint_schema", tt.schema)
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/test", bytes.NewReader(tt.body))
			require.NoError(t, err)

			result, err := validator.InterceptRequest(ctx, req)

			if tt.shouldPass {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			} else {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			}
		})
	}
}

func TestSchemaValidator_InterceptRequest_GET_ShouldSkip(t *testing.T) {
	validator := NewSchemaValidator()
	schema := []byte(`{"type": "object"}`)

	ctx := context.WithValue(context.Background(), "endpoint_schema", schema)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/test", nil)
	require.NoError(t, err)

	result, err := validator.InterceptRequest(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSchemaValidator_InterceptResponse(t *testing.T) {
	validator := NewSchemaValidator()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
	}

	result, err := validator.InterceptResponse(context.Background(), req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp, result)
}
