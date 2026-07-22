package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
	"github.com/stretchr/testify/assert"
)

func TestError_ClientMessageSanitizedOnInternalError(t *testing.T) {
	rec := httptest.NewRecorder()
	err := domainErr.New(domainErr.ErrInternal, "database connection failed", errors.New("host=db.internal:5432"))

	Error(rec, err)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var body domainErr.ErrorResponse
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, "an internal error occurred", body.Message)
	assert.NotContains(t, body.Message, "db.internal")
}

func TestError_PreservesClientSafeMessage(t *testing.T) {
	rec := httptest.NewRecorder()
	err := domainErr.New(domainErr.ErrNotFound, "endpoint not found", nil)

	Error(rec, err)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var body domainErr.ErrorResponse
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Contains(t, body.Message, "endpoint not found")
}
