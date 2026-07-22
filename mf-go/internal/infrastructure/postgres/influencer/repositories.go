package influencer

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/masterfabric-go/masterfabric/internal/domain/influencer/model"
	domainRepo "github.com/masterfabric-go/masterfabric/internal/domain/influencer/repository"
	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
)

type ScoreRepo struct {
	db *pgxpool.Pool
}

func NewScoreRepo(db *pgxpool.Pool) *ScoreRepo {
	return &ScoreRepo{db: db}
}

var _ domainRepo.ScoreRepository = (*ScoreRepo)(nil)

func (r *ScoreRepo) Create(ctx context.Context, score *model.InfluencerScore) error {
	if score.ID == uuid.Nil {
		score.ID = uuid.New()
	}
	now := time.Now().UTC()
	score.CreatedAt = now
	score.UpdatedAt = now

	_, err := r.db.Exec(ctx, `
		INSERT INTO influencer_scores (
			id, user_id, influencer_name, platform,
			overall_score, engagement_score, audience_score, brand_fit_score,
			raw_payload, notes, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		score.ID, score.UserID, score.InfluencerName, score.Platform,
		score.OverallScore, score.EngagementScore, score.AudienceScore, score.BrandFitScore,
		score.RawPayload, score.Notes, score.CreatedAt, score.UpdatedAt,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to create score", err)
	}
	return nil
}

func (r *ScoreRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*model.InfluencerScore, error) {
	var s model.InfluencerScore
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, influencer_name, platform,
			overall_score, engagement_score, audience_score, brand_fit_score,
			raw_payload, notes, created_at, updated_at
		FROM influencer_scores WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(
		&s.ID, &s.UserID, &s.InfluencerName, &s.Platform,
		&s.OverallScore, &s.EngagementScore, &s.AudienceScore, &s.BrandFitScore,
		&s.RawPayload, &s.Notes, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "score not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get score", err)
	}
	return &s, nil
}

func (r *ScoreRepo) ListByUser(ctx context.Context, userID uuid.UUID, platform string, limit, offset int) ([]model.InfluencerScore, error) {
	query := `
		SELECT id, user_id, influencer_name, platform,
			overall_score, engagement_score, audience_score, brand_fit_score,
			raw_payload, notes, created_at, updated_at
		FROM influencer_scores WHERE user_id = $1`
	args := []any{userID}
	if platform != "" {
		query += ` AND platform = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
		args = append(args, platform, limit, offset)
	} else {
		query += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "failed to list scores", err)
	}
	defer rows.Close()

	var out []model.InfluencerScore
	for rows.Next() {
		var s model.InfluencerScore
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.InfluencerName, &s.Platform,
			&s.OverallScore, &s.EngagementScore, &s.AudienceScore, &s.BrandFitScore,
			&s.RawPayload, &s.Notes, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, domainErr.New(domainErr.ErrInternal, "failed to scan score", err)
		}
		out = append(out, s)
	}
	if out == nil {
		out = []model.InfluencerScore{}
	}
	return out, nil
}

func (r *ScoreRepo) Update(ctx context.Context, score *model.InfluencerScore) error {
	score.UpdatedAt = time.Now().UTC()
	tag, err := r.db.Exec(ctx, `
		UPDATE influencer_scores SET
			influencer_name = $1, platform = $2,
			overall_score = $3, engagement_score = $4, audience_score = $5, brand_fit_score = $6,
			raw_payload = $7, notes = $8, updated_at = $9
		WHERE id = $10 AND user_id = $11`,
		score.InfluencerName, score.Platform,
		score.OverallScore, score.EngagementScore, score.AudienceScore, score.BrandFitScore,
		score.RawPayload, score.Notes, score.UpdatedAt, score.ID, score.UserID,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to update score", err)
	}
	if tag.RowsAffected() == 0 {
		return domainErr.New(domainErr.ErrNotFound, "score not found", nil)
	}
	return nil
}

func (r *ScoreRepo) Delete(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM influencer_scores WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to delete score", err)
	}
	if tag.RowsAffected() == 0 {
		return domainErr.New(domainErr.ErrNotFound, "score not found", nil)
	}
	return nil
}

type AnalysisRepo struct {
	db *pgxpool.Pool
}

func NewAnalysisRepo(db *pgxpool.Pool) *AnalysisRepo {
	return &AnalysisRepo{db: db}
}

var _ domainRepo.AnalysisRepository = (*AnalysisRepo)(nil)

func (r *AnalysisRepo) Create(ctx context.Context, analysis *model.InfluencerAnalysis) error {
	if analysis.ID == uuid.Nil {
		analysis.ID = uuid.New()
	}
	now := time.Now().UTC()
	analysis.CreatedAt = now
	analysis.UpdatedAt = now

	_, err := r.db.Exec(ctx, `
		INSERT INTO influencer_analyses (
			id, user_id, influencer_name, platform, analysis_type,
			summary, insights, raw_llm_output, score_id, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		analysis.ID, analysis.UserID, analysis.InfluencerName, analysis.Platform, analysis.AnalysisType,
		analysis.Summary, analysis.Insights, analysis.RawLLMOutput, analysis.ScoreID, analysis.CreatedAt, analysis.UpdatedAt,
	)
	if err != nil {
		return domainErr.New(domainErr.ErrInternal, "failed to create analysis", err)
	}
	return nil
}

func (r *AnalysisRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*model.InfluencerAnalysis, error) {
	var a model.InfluencerAnalysis
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, influencer_name, platform, analysis_type,
			summary, insights, raw_llm_output, score_id, created_at, updated_at
		FROM influencer_analyses WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(
		&a.ID, &a.UserID, &a.InfluencerName, &a.Platform, &a.AnalysisType,
		&a.Summary, &a.Insights, &a.RawLLMOutput, &a.ScoreID, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainErr.New(domainErr.ErrNotFound, "analysis not found", nil)
		}
		return nil, domainErr.New(domainErr.ErrInternal, "failed to get analysis", err)
	}
	return &a, nil
}

func (r *AnalysisRepo) ListByUser(ctx context.Context, userID uuid.UUID, platform string, limit, offset int) ([]model.InfluencerAnalysis, error) {
	query := `
		SELECT id, user_id, influencer_name, platform, analysis_type,
			summary, insights, raw_llm_output, score_id, created_at, updated_at
		FROM influencer_analyses WHERE user_id = $1`
	args := []any{userID}
	if platform != "" {
		query += ` AND platform = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
		args = append(args, platform, limit, offset)
	} else {
		query += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, domainErr.New(domainErr.ErrInternal, "failed to list analyses", err)
	}
	defer rows.Close()

	var out []model.InfluencerAnalysis
	for rows.Next() {
		var a model.InfluencerAnalysis
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.InfluencerName, &a.Platform, &a.AnalysisType,
			&a.Summary, &a.Insights, &a.RawLLMOutput, &a.ScoreID, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, domainErr.New(domainErr.ErrInternal, "failed to scan analysis", err)
		}
		out = append(out, a)
	}
	if out == nil {
		out = []model.InfluencerAnalysis{}
	}
	return out, nil
}
