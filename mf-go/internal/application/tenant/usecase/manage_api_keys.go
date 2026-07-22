package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/tenant/dto"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/repository"
)

// ManageAPIKeysUseCase handles API key management.
type ManageAPIKeysUseCase struct {
	keyRepo repository.APIKeyRepository
}

// NewManageAPIKeysUseCase creates a new ManageAPIKeysUseCase.
func NewManageAPIKeysUseCase(keyRepo repository.APIKeyRepository) *ManageAPIKeysUseCase {
	return &ManageAPIKeysUseCase{keyRepo: keyRepo}
}

// CreateKey creates a new API key for an app.
func (uc *ManageAPIKeysUseCase) CreateKey(ctx context.Context, appID uuid.UUID, req dto.CreateAPIKeyRequest) (*dto.APIKeyResponse, error) {
	// Generate random key
	rawKey, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("generate api key: %w", err)
	}

	// Hash it for storage
	hash := hashAPIKey(rawKey)

	key := &model.AppAPIKey{
		AppID:    appID,
		KeyHash:  hash,
		Name:     req.Name,
		IsActive: true,
	}

	if err := uc.keyRepo.Create(ctx, key); err != nil {
		return nil, err
	}

	return &dto.APIKeyResponse{
		ID:        key.ID,
		AppID:     key.AppID,
		Name:      key.Name,
		Key:       rawKey, // Only returned once
		ExpiresAt: key.ExpiresAt,
		IsActive:  key.IsActive,
		CreatedAt: key.CreatedAt,
	}, nil
}

// RevokeKey revokes an API key.
func (uc *ManageAPIKeysUseCase) RevokeKey(ctx context.Context, keyID uuid.UUID) error {
	return uc.keyRepo.Revoke(ctx, keyID)
}

// ListKeys lists all API keys for an app.
func (uc *ManageAPIKeysUseCase) ListKeys(ctx context.Context, appID uuid.UUID) ([]*dto.APIKeyResponse, error) {
	keys, err := uc.keyRepo.ListByApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	var result []*dto.APIKeyResponse
	for _, k := range keys {
		result = append(result, &dto.APIKeyResponse{
			ID:        k.ID,
			AppID:     k.AppID,
			Name:      k.Name,
			ExpiresAt: k.ExpiresAt,
			IsActive:  k.IsActive,
			CreatedAt: k.CreatedAt,
		})
	}
	return result, nil
}

func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "mf_" + hex.EncodeToString(bytes), nil
}

func hashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
