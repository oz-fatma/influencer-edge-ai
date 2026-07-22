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

// AppRepo implements repository.AppRepository with PostgreSQL.
type AppRepo struct {
	db *pgxpool.Pool
}

// NewAppRepo creates a new AppRepo.
func NewAppRepo(db *pgxpool.Pool) *AppRepo {
	return &AppRepo{db: db}
}

func (r *AppRepo) Create(ctx context.Context, app *model.App) error {
	if app.ID == uuid.Nil {
		app.ID = uuid.New()
	}
	now := time.Now().UTC()
	app.CreatedAt = now
	app.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO apps (id, organization_id, name, slug, status, sla_tier, rate_limit_policy, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		app.ID, app.OrganizationID, app.Name, app.Slug, app.Status, app.SLATier, app.RateLimitPolicy, app.CreatedAt, app.UpdatedAt,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to create app", err)
	}
	return nil
}

func (r *AppRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.App, error) {
	var a model.App
	err := r.db.QueryRow(ctx,
		`SELECT id, organization_id, name, slug, status, sla_tier, rate_limit_policy, created_at, updated_at
		 FROM apps WHERE id = $1`, id,
	).Scan(&a.ID, &a.OrganizationID, &a.Name, &a.Slug, &a.Status, &a.SLATier, &a.RateLimitPolicy, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "app not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get app", err)
	}
	return &a, nil
}

func (r *AppRepo) GetBySlug(ctx context.Context, orgID uuid.UUID, slug string) (*model.App, error) {
	var a model.App
	err := r.db.QueryRow(ctx,
		`SELECT id, organization_id, name, slug, status, sla_tier, rate_limit_policy, created_at, updated_at
		 FROM apps WHERE organization_id = $1 AND slug = $2`, orgID, slug,
	).Scan(&a.ID, &a.OrganizationID, &a.Name, &a.Slug, &a.Status, &a.SLATier, &a.RateLimitPolicy, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "app not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get app by slug", err)
	}
	return &a, nil
}

func (r *AppRepo) Update(ctx context.Context, app *model.App) error {
	app.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx,
		`UPDATE apps SET name=$1, status=$2, sla_tier=$3, rate_limit_policy=$4, updated_at=$5 WHERE id=$6`,
		app.Name, app.Status, app.SLATier, app.RateLimitPolicy, app.UpdatedAt, app.ID,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to update app", err)
	}
	return nil
}

func (r *AppRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM apps WHERE id=$1`, id)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to delete app", err)
	}
	return nil
}

func (r *AppRepo) ListByOrg(ctx context.Context, orgID uuid.UUID, offset, limit int) ([]*model.App, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM apps WHERE organization_id=$1`, orgID).Scan(&total)
	if err != nil {
		return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to count apps", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, organization_id, name, slug, status, sla_tier, rate_limit_policy, created_at, updated_at
		 FROM apps WHERE organization_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, orgID, limit, offset,
	)
	if err != nil {
		return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to list apps", err)
	}
	defer rows.Close()

	var apps []*model.App
	for rows.Next() {
		var a model.App
		if err := rows.Scan(&a.ID, &a.OrganizationID, &a.Name, &a.Slug, &a.Status, &a.SLATier, &a.RateLimitPolicy, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, domainErr.New(domainErr.ErrInternal, "failed to scan app", err)
		}
		apps = append(apps, &a)
	}
	return apps, total, nil
}
