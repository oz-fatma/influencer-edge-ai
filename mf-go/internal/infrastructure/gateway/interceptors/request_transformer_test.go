package interceptors

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestTransformer_InterceptRequest(t *testing.T) {
	headerTransforms := map[string]string{
		"X-Custom-Header": "value-{org_id}",
		"X-User-ID":       "{user_id}",
	}
	queryTransforms := map[string]string{
		"org": "{org_id}",
	}

	transformer := NewRequestTransformer(headerTransforms, queryTransforms, nil)

	ctx := context.WithValue(context.Background(), "org_id", "org-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/test?existing=value", nil)
	require.NoError(t, err)

	result, err := transformer.InterceptRequest(ctx, req)
	require.NoError(t, err)

	assert.Equal(t, "value-org-123", result.Header.Get("X-Custom-Header"))
	assert.Equal(t, "user-456", result.Header.Get("X-User-ID"))
	assert.Contains(t, result.URL.RawQuery, "org=org-123")
	assert.Contains(t, result.URL.RawQuery, "existing=value")
}

func TestRequestTransformer_BodyTransform(t *testing.T) {
	bodyTransform := func(body []byte) ([]byte, error) {
		// Simple transformation: add a field
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return body, err
		}
		data["transformed"] = true
		return json.Marshal(data)
	}

	transformer := NewRequestTransformer(nil, nil, bodyTransform)

	body := []byte(`{"name": "test"}`)
	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	require.NoError(t, err)

	result, err := transformer.InterceptRequest(context.Background(), req)
	require.NoError(t, err)

	var transformed map[string]interface{}
	err = json.NewDecoder(result.Body).Decode(&transformed)
	require.NoError(t, err)

	assert.Equal(t, "test", transformed["name"])
	assert.Equal(t, true, transformed["transformed"])
}
