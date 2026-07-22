package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/iam/service"
	"github.com/masterfabric-go/masterfabric/internal/shared/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestJWTService() *JWTService {
	return NewJWTService(config.JWTConfig{
		Secret:          "test-secret-key-for-testing-only",
		ExpirationHours: 1,
		Issuer:          "test",
	})
}

func TestJWTService_HashAndVerifyPassword(t *testing.T) {
	svc := newTestJWTService()

	hash, err := svc.HashPassword("mypassword123")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "mypassword123", hash)

	// Verify correct password
	err = svc.VerifyPassword(hash, "mypassword123")
	assert.NoError(t, err)

	// Verify wrong password
	err = svc.VerifyPassword(hash, "wrongpassword")
	assert.Error(t, err)
}

func TestJWTService_GenerateAndValidateToken(t *testing.T) {
	svc := newTestJWTService()
	ctx := context.Background()

	userID := uuid.New()
	orgID := uuid.New()

	claims := service.TokenClaims{
		UserID:         userID,
		Email:          "test@example.com",
		OrganizationID: orgID,
		Roles:          []string{"org.admin"},
		Permissions:    []string{"org.manage_users"},
	}

	token, err := svc.GenerateToken(ctx, claims)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	parsed, err := svc.ValidateToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, userID, parsed.UserID)
	assert.Equal(t, "test@example.com", parsed.Email)
	assert.Equal(t, orgID, parsed.OrganizationID)
	assert.Equal(t, []string{"org.admin"}, parsed.Roles)
	assert.Equal(t, []string{"org.manage_users"}, parsed.Permissions)
}

func TestJWTService_InvalidToken(t *testing.T) {
	svc := newTestJWTService()
	ctx := context.Background()

	_, err := svc.ValidateToken(ctx, "invalid-token-string")
	assert.Error(t, err)
}

func TestJWTService_WrongSecret(t *testing.T) {
	svc1 := newTestJWTService()
	svc2 := NewJWTService(config.JWTConfig{
		Secret:          "different-secret",
		ExpirationHours: 1,
		Issuer:          "test",
	})
	ctx := context.Background()

	token, err := svc1.GenerateToken(ctx, service.TokenClaims{
		UserID: uuid.New(),
		Email:  "test@example.com",
	})
	require.NoError(t, err)

	_, err = svc2.ValidateToken(ctx, token)
	assert.Error(t, err)
}
