package gateway

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDynamicHandlerResolver_RefusesRedirects(t *testing.T) {
	resolver := NewDynamicHandlerResolver(nil, nil, nil)

	client := resolver.httpClient
	assert.Equal(t, 30*time.Second, client.Timeout)
	assert.NotNil(t, client.CheckRedirect)

	redirectErr := client.CheckRedirect(&http.Request{}, []*http.Request{{}})
	assert.ErrorIs(t, redirectErr, http.ErrUseLastResponse)
}
