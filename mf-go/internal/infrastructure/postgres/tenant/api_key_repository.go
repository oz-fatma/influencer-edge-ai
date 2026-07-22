package tenant

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/model"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
)

// APIKeyRepo implements repository.APIKeyRepository with PostgreSQL.
type APIKeyRepo struct {
	db *pgxpool.Pool
}

// NewAPIKeyRepo creates a new APIKeyRepo.
func NewAPIKeyRepo(db *pgxpool.Pool) *APIKeyRepo {
	return &APIKeyRepo{db: db}
}

func (r *APIKeyRepo) Create(ctx context.Context, key *model.AppAPIKey) error {
	if key.ID == uuid.Nil {
		key.ID = uuid.New()
	}
	key.CreatedAt = time.Now().UTC()

	_, err := r.db.Exec(ctx,
		`INSERT INTO app_api_keys (id, app_id, key_hash, name, scopes, expires_at, is_active, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		key.ID, key.AppID, key.KeyHash, key.Name, key.Scopes, key.ExpiresAt, key.IsActive, key.CreatedAt,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to create api key", err)
	}
	return nil
}

func (r *APIKeyRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.AppAPIKey, error) {
	var k model.AppAPIKey
	err := r.db.QueryRow(ctx,
		`SELECT id, app_id, key_hash, name, scopes, expires_at, is_active, created_at
		 FROM app_api_keys WHERE id = $1`, id,
	).Scan(&k.ID, &k.AppID, &k.KeyHash, &k.Name, &k.Scopes, &k.ExpiresAt, &k.IsActive, &k.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "api key not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get api key", err)
	}
	return &k, nil
}

func (r *APIKeyRepo) GetByHash(ctx context.Context, hash string) (*model.AppAPIKey, error) {
	var k model.AppAPIKey
	err := r.db.QueryRow(ctx,
		`SELECT id, app_id, key_hash, name, scopes, expires_at, is_active, created_at
		 FROM app_api_keys WHERE key_hash = $1 AND is_active = true`, hash,
	).Scan(&k.ID, &k.AppID, &k.KeyHash, &k.Name, &k.Scopes, &k.ExpiresAt, &k.IsActive, &k.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "api key not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get api key by hash", err)
	}
	return &k, nil
}

func (r *APIKeyRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE app_api_keys SET is_active = false WHERE id=$1`, id,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to revoke api key", err)
	}
	return nil
}

func (r *APIKeyRepo) ListByApp(ctx context.Context, appID uuid.UUID) ([]*model.AppAPIKey, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, app_id, key_hash, name, scopes, expires_at, is_active, created_at
		 FROM app_api_keys WHERE app_id = $1 ORDER BY created_at DESC`, appID,
	)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "failed to list api keys", err)
	}
	defer rows.Close()

	var keys []*model.AppAPIKey
	for rows.Next() {
		var k model.AppAPIKey
		if err := rows.Scan(&k.ID, &k.AppID, &k.KeyHash, &k.Name, &k.Scopes, &k.ExpiresAt, &k.IsActive, &k.CreatedAt); err != nil {
			return nil, domainErr.New(domainErr.ErrInternal, "failed to scan api key", err)
		}
		keys = append(keys, &k)
	}
	return keys, nil
}
