package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
	tenantRepo "github.com/masterfabric-go/masterfabric/internal/domain/tenant/repository"
)

// Resolver resolves tenant, app, and endpoint context from request metadata.
type Resolver struct {
	orgRepo    tenantRepo.OrgRepository
	appRepo    tenantRepo.AppRepository
	apiKeyRepo tenantRepo.APIKeyRepository
}

// NewResolver creates a new Resolver.
func NewResolver(
	orgRepo tenantRepo.OrgRepository,
	appRepo tenantRepo.AppRepository,
	apiKeyRepo tenantRepo.APIKeyRepository,
) *Resolver {
	return &Resolver{
		orgRepo:    orgRepo,
		appRepo:    appRepo,
		apiKeyRepo: apiKeyRepo,
	}
}

// ResolveAppByAPIKey resolves an app from a raw API key string.
func (r *Resolver) ResolveAppByAPIKey(ctx context.Context, rawKey string) (uuid.UUID, uuid.UUID, error) {
	h := sha256.Sum256([]byte(rawKey))
	hash := hex.EncodeToString(h[:])

	key, err := r.apiKeyRepo.GetByHash(ctx, hash)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	app, err := r.appRepo.GetByID(ctx, key.AppID)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return app.OrganizationID, app.ID, nil
}

// ResolveOrgBySlug resolves an organization by subdomain slug.
func (r *Resolver) ResolveOrgBySlug(ctx context.Context, slug string) (uuid.UUID, error) {
	org, err := r.orgRepo.GetBySlug(ctx, slug)
	if err != nil {
		return uuid.Nil, err
	}
	return org.ID, nil
}
