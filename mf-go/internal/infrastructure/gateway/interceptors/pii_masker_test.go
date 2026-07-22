package interceptors

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPIIMasker_InterceptRequest(t *testing.T) {
	masker := NewPIIMasker([]string{"password", "api_key", "secret"}, "***")

	tests := []struct {
		name           string
		maskPII        bool
		body           map[string]interface{}
		expectedMasked map[string]interface{}
	}{
		{
			name:    "mask PII fields",
			maskPII: true,
			body: map[string]interface{}{
				"username": "john",
				"password": "secret123",
				"api_key":  "key123",
				"email":    "john@example.com",
			},
			expectedMasked: map[string]interface{}{
				"username": "john",
				"password": "***",
				"api_key":  "***",
				"email":    "john@example.com",
			},
		},
		{
			name:    "PII masking disabled",
			maskPII: false,
			body: map[string]interface{}{
				"password": "secret123",
			},
			expectedMasked: map[string]interface{}{
				"password": "secret123", // Should not be masked
			},
		},
		{
			name:    "nested PII fields",
			maskPII: true,
			body: map[string]interface{}{
				"user": map[string]interface{}{
					"name":     "John",
					"password": "secret123",
				},
			},
			expectedMasked: map[string]interface{}{
				"user": map[string]interface{}{
					"name":     "John",
					"password": "***",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.maskPII {
				ctx = context.WithValue(ctx, "pii_masking", true)
			}

			bodyJSON, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/test", bytes.NewReader(bodyJSON))
			require.NoError(t, err)

			result, err := masker.InterceptRequest(ctx, req)
			require.NoError(t, err)

			if tt.maskPII {
				var maskedBody map[string]interface{}
				err = json.NewDecoder(result.Body).Decode(&maskedBody)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedMasked, maskedBody)
			} else {
				// Body should remain unchanged
				var originalBody map[string]interface{}
				err = json.NewDecoder(result.Body).Decode(&originalBody)
				require.NoError(t, err)
				assert.Equal(t, tt.body, originalBody)
			}
		})
	}
}

func TestPIIMasker_InterceptResponse(t *testing.T) {
	masker := NewPIIMasker([]string{"password", "token"}, "***")

	ctx := context.WithValue(context.Background(), "pii_masking", true)
	body := map[string]interface{}{
		"user": map[string]interface{}{
			"id":       "123",
			"password": "secret",
		},
	}

	bodyJSON, _ := json.Marshal(body)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(bodyJSON)),
	}

	result, err := masker.InterceptResponse(ctx, nil, resp)
	require.NoError(t, err)

	var maskedBody map[string]interface{}
	err = json.NewDecoder(result.Body).Decode(&maskedBody)
	require.NoError(t, err)

	expected := map[string]interface{}{
		"user": map[string]interface{}{
			"id":       "123",
			"password": "***",
		},
	}
	assert.Equal(t, expected, maskedBody)
}

func TestPIIMasker_NonJSONBody(t *testing.T) {
	masker := NewPIIMasker([]string{"password"}, "***")
	ctx := context.WithValue(context.Background(), "pii_masking", true)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/test", bytes.NewReader([]byte("not json")))
	result, err := masker.InterceptRequest(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
}
