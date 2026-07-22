package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractBearerToken_FromQuery(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/ws?token=query-token", nil)
	assert.Equal(t, "query-token", ExtractBearerToken(r))
}

func TestExtractBearerToken_FromHeader(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/ws", nil)
	r.Header.Set("Authorization", "Bearer header-token")
	assert.Equal(t, "header-token", ExtractBearerToken(r))
}

func TestExtractBearerToken_QueryTakesPrecedence(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/ws?token=query-token", nil)
	r.Header.Set("Authorization", "Bearer header-token")
	assert.Equal(t, "query-token", ExtractBearerToken(r))
}

func TestExtractBearerToken_Missing(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/ws", nil)
	assert.Empty(t, ExtractBearerToken(r))
}
