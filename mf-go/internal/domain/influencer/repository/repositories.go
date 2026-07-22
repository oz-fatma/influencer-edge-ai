package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/influencer/model"
)

type ScoreRepository interface {
	Create(ctx context.Context, score *model.InfluencerScore) error
	GetByID(ctx context.Context, userID, id uuid.UUID) (*model.InfluencerScore, error)
	ListByUser(ctx context.Context, userID uuid.UUID, platform string, limit, offset int) ([]model.InfluencerScore, error)
	Update(ctx context.Context, score *model.InfluencerScore) error
	Delete(ctx context.Context, userID, id uuid.UUID) error
}

type AnalysisRepository interface {
	Create(ctx context.Context, analysis *model.InfluencerAnalysis) error
	GetByID(ctx context.Context, userID, id uuid.UUID) (*model.InfluencerAnalysis, error)
	ListByUser(ctx context.Context, userID uuid.UUID, platform string, limit, offset int) ([]model.InfluencerAnalysis, error)
}
