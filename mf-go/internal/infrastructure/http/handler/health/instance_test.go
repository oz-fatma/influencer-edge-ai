package health

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveInstanceID_prefersInstanceEnv(t *testing.T) {
	t.Setenv("INSTANCE_ID", "pod-a")
	t.Setenv("HOSTNAME", "container-xyz")
	assert.Equal(t, "pod-a", ResolveInstanceID())
}

func TestResolveInstanceID_fallsBackToHostname(t *testing.T) {
	t.Setenv("INSTANCE_ID", "")
	t.Setenv("HOSTNAME", "replica-2")
	assert.Equal(t, "replica-2", ResolveInstanceID())
}

func TestLiveness_setsInstanceIDHeaderAndBody(t *testing.T) {
	t.Setenv("INSTANCE_ID", "demo-instance-1")
	handler := NewHandler(nil, nil, "")

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	handler.Liveness(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "demo-instance-1", rec.Header().Get("X-Instance-ID"))
	assert.Contains(t, rec.Body.String(), `"instance_id":"demo-instance-1"`)
}

func TestNewHandler_usesResolveWhenEmpty(t *testing.T) {
	t.Setenv("INSTANCE_ID", "")
	host, err := os.Hostname()
	if err != nil {
		t.Skip("hostname unavailable")
	}
	t.Setenv("HOSTNAME", host)

	handler := NewHandler(nil, nil, "")
	assert.Equal(t, host, handler.instanceID)
}
