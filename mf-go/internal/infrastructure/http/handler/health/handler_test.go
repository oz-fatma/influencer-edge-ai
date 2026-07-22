package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type failingPinger struct{}

func (failingPinger) Ping(context.Context) error {
	return errors.New("connection refused host=secret-db:5432")
}

func TestReadiness_DoesNotExposeInternalErrors(t *testing.T) {
	handler := &Handler{db: failingPinger{}}

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()
	handler.Readiness(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.NotContains(t, rec.Body.String(), "secret-db")
	assert.Contains(t, rec.Body.String(), "unhealthy")
}
