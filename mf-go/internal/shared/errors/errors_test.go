package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDomainError_Error(t *testing.T) {
	err := New(ErrNotFound, "user not found", nil)
	assert.Contains(t, err.Error(), "user not found")
	assert.Contains(t, err.Error(), "resource not found")
}

func TestDomainError_Unwrap(t *testing.T) {
	err := New(ErrNotFound, "user not found", nil)
	assert.True(t, errors.Is(err, ErrNotFound))
	assert.False(t, errors.Is(err, ErrUnauthorized))
}

func TestHTTPStatusCode(t *testing.T) {
	tests := []struct {
		err      error
		expected int
	}{
		{New(ErrNotFound, "not found", nil), http.StatusNotFound},
		{New(ErrAlreadyExists, "exists", nil), http.StatusConflict},
		{New(ErrUnauthorized, "unauthorized", nil), http.StatusUnauthorized},
		{New(ErrForbidden, "forbidden", nil), http.StatusForbidden},
		{New(ErrValidation, "invalid", nil), http.StatusUnprocessableEntity},
		{New(ErrBadRequest, "bad", nil), http.StatusBadRequest},
		{New(ErrRateLimited, "rate", nil), http.StatusTooManyRequests},
		{New(ErrInternal, "internal", nil), http.StatusInternalServerError},
		{errors.New("unknown"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, HTTPStatusCode(tt.err))
	}
}
