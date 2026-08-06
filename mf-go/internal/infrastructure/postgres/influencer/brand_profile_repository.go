package influencer

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/masterfabric-go/masterfabric/internal/domain/influencer/model"
	domainRepo "github.com/masterfabric-go/masterfabric/internal/domain/influencer/repository"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
)

type BrandProfileRepo struct {
	db *pgxpool.Pool
}

func NewBrandProfileRepo(db *pgxpool.Pool) *BrandProfileRepo {
	return &BrandProfileRepo{db: db}
}

var _ domainRepo.BrandProfileRepository = (*BrandProfileRepo)(nil)

const brandProfileSelectColumns = `
	id, user_id, name, industry, target_audience, budget_range,
	brand_values, campaign_goal, created_at, updated_at`

func scanBrandProfile(row pgx.Row) (model.BrandProfile, error) {
	var p model.BrandProfile
	var budgetRange pgtype.Text
	err := row.Scan(
		&p.ID, &p.UserID, &p.Name, &p.Industry, &p.TargetAudience, &budgetRange,
		&p.BrandValues, &p.CampaignGoal, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return p, err
	}
	if budgetRange.Valid {
		p.BudgetRange = budgetRange.String
	}
	return p, nil
}

func (r *BrandProfileRepo) Create(ctx context.Context, profile *model.BrandProfile) error {
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}
	now := time.Now().UTC()
	profile.CreatedAt = now
	profile.UpdatedAt = now

	_, err := r.db.Exec(ctx, `
		INSERT INTO brand_profiles (
			id, user_id, name, industry, target_audience, budget_range,
			brand_values, campaign_goal, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		profile.ID, profile.UserID, profile.Name, profile.Industry, profile.TargetAudience,
		nullIfEmpty(profile.BudgetRange), profile.BrandValues, profile.CampaignGoal,
		profile.CreatedAt, profile.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainErr.New(domainErr.ErrAlreadyExists, "brand profile with this name already exists", err)
		}
		return domainErr.New(domainErr.ErrInternal, "failed to create brand profile", err)
	}
	return nil
}

func (r *BrandProfileRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*model.BrandProfile, error) {
	p, err := scanBrandProfile(r.db.QueryRow(ctx, `
		SELECT`+brandProfileSelectColumns+`
		FROM brand_profiles WHERE id = $1 AND user_id = $2`, id, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "brand profile not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get brand profile", err)
	}
	return &p, nil
}

func (r *BrandProfileRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.BrandProfile, error) {
	rows, err := r.db.Query(ctx, `
		SELECT`+brandProfileSelectColumns+`
		FROM brand_profiles WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "failed to list brand profiles", err)
	}
	defer rows.Close()

	var out []model.BrandProfile
	for rows.Next() {
		p, err := scanBrandProfile(rows)
		if err != nil {
			return nil, domainErr.New(domainErr.ErrInternal, "failed to scan brand profile", err)
		}
		out = append(out, p)
	}
	if out == nil {
		out = []model.BrandProfile{}
	}
	return out, nil
}

func (r *BrandProfileRepo) Update(ctx context.Context, profile *model.BrandProfile) error {
	profile.UpdatedAt = time.Now().UTC()
	tag, err := r.db.Exec(ctx, `
		UPDATE brand_profiles SET
			name = $1, industry = $2, target_audience = $3, budget_range = $4,
			brand_values = $5, campaign_goal = $6, updated_at = $7
		WHERE id = $8 AND user_id = $9`,
		profile.Name, profile.Industry, profile.TargetAudience, nullIfEmpty(profile.BudgetRange),
		profile.BrandValues, profile.CampaignGoal, profile.UpdatedAt, profile.ID, profile.UserID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainErr.New(domainErr.ErrAlreadyExists, "brand profile with this name already exists", err)
		}
		return domainErr.New(domainErr.ErrInternal, "failed to update brand profile", err)
	}
	if tag.RowsAffected() == 0 {
		return domainErr.New(domainErr.ErrNotFound, "brand profile not found", nil)
	}
	return nil
}

func (r *BrandProfileRepo) Delete(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM brand_profiles WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to delete brand profile", err)
	}
	if tag.RowsAffected() == 0 {
		return domainErr.New(domainErr.ErrNotFound, "brand profile not found", nil)
	}
	return nil
}
