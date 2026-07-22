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

// OrgRepo implements repository.OrgRepository with PostgreSQL.
type OrgRepo struct {
	db *pgxpool.Pool
}

// NewOrgRepo creates a new OrgRepo.
func NewOrgRepo(db *pgxpool.Pool) *OrgRepo {
	return &OrgRepo{db: db}
}

func (r *OrgRepo) Create(ctx context.Context, org *model.Organization) error {
	if org.ID == uuid.Nil {
		org.ID = uuid.New()
	}
	now := time.Now().UTC()
	org.CreatedAt = now
	org.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO organizations (id, name, slug, status, sso_config, data_retention_policy, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		org.ID, org.Name, org.Slug, org.Status, org.SSOConfig, org.DataRetentionPolicy, org.CreatedAt, org.UpdatedAt,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to create organization", err)
	}
	return nil
}

func (r *OrgRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	var o model.Organization
	err := r.db.QueryRow(ctx,
		`SELECT id, name, slug, status, sso_config, data_retention_policy, created_at, updated_at
		 FROM organizations WHERE id = $1`, id,
	).Scan(&o.ID, &o.Name, &o.Slug, &o.Status, &o.SSOConfig, &o.DataRetentionPolicy, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "organization not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get organization", err)
	}
	return &o, nil
}

func (r *OrgRepo) GetBySlug(ctx context.Context, slug string) (*model.Organization, error) {
	var o model.Organization
	err := r.db.QueryRow(ctx,
		`SELECT id, name, slug, status, sso_config, data_retention_policy, created_at, updated_at
		 FROM organizations WHERE slug = $1`, slug,
	).Scan(&o.ID, &o.Name, &o.Slug, &o.Status, &o.SSOConfig, &o.DataRetentionPolicy, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "organization not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get organization by slug", err)
	}
	return &o, nil
}

func (r *OrgRepo) Update(ctx context.Context, org *model.Organization) error {
	org.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx,
		`UPDATE organizations SET name=$1, status=$2, sso_config=$3, data_retention_policy=$4, updated_at=$5 WHERE id=$6`,
		org.Name, org.Status, org.SSOConfig, org.DataRetentionPolicy, org.UpdatedAt, org.ID,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to update organization", err)
	}
	return nil
}

func (r *OrgRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM organizations WHERE id=$1`, id)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to delete organization", err)
	}
	return nil
}

func (r *OrgRepo) List(ctx context.Context, offset, limit int) ([]*model.Organization, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM organizations`).Scan(&total)
	if err != nil {
		return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to count organizations", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, name, slug, status, sso_config, data_retention_policy, created_at, updated_at
		 FROM organizations ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset,
	)
	if err != nil {
		return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to list organizations", err)
	}
	defer rows.Close()

	var orgs []*model.Organization
	for rows.Next() {
		var o model.Organization
		if err := rows.Scan(&o.ID, &o.Name, &o.Slug, &o.Status, &o.SSOConfig, &o.DataRetentionPolicy, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to scan organization", err)
		}
		orgs = append(orgs, &o)
	}
	return orgs, total, nil
}
