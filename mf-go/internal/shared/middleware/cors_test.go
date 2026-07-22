package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORSOptions_DisablesCredentialsForWildcard(t *testing.T) {
	opts := CORSOptions([]string{"*"})
	assert.False(t, opts.AllowCredentials)
	assert.Equal(t, []string{"*"}, opts.AllowedOrigins)
}

func TestCORSOptions_DisablesCredentialsWhenEmpty(t *testing.T) {
	opts := CORSOptions(nil)
	assert.False(t, opts.AllowCredentials)
	assert.Empty(t, opts.AllowedOrigins)
}

func TestCORSOptions_AllowsCredentialsForExplicitOrigins(t *testing.T) {
	opts := CORSOptions([]string{"https://app.example.com"})
	assert.True(t, opts.AllowCredentials)
}
