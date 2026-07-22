-- +goose Up
CREATE TABLE IF NOT EXISTS influencer_scores (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    influencer_name   VARCHAR(255) NOT NULL,
    platform          VARCHAR(64) NOT NULL,
    overall_score     DOUBLE PRECISION NOT NULL DEFAULT 0,
    engagement_score  DOUBLE PRECISION NOT NULL DEFAULT 0,
    audience_score    DOUBLE PRECISION NOT NULL DEFAULT 0,
    brand_fit_score   DOUBLE PRECISION NOT NULL DEFAULT 0,
    raw_payload       TEXT NOT NULL DEFAULT '',
    notes             TEXT NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_influencer_scores_user_id ON influencer_scores(user_id);
CREATE INDEX IF NOT EXISTS idx_influencer_scores_platform ON influencer_scores(platform);
CREATE INDEX IF NOT EXISTS idx_influencer_scores_created_at ON influencer_scores(created_at DESC);

CREATE TABLE IF NOT EXISTS influencer_analyses (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    influencer_name   VARCHAR(255) NOT NULL,
    platform          VARCHAR(64) NOT NULL,
    analysis_type     VARCHAR(64) NOT NULL,
    summary           TEXT NOT NULL,
    insights          TEXT NOT NULL DEFAULT '',
    raw_llm_output    TEXT NOT NULL DEFAULT '',
    score_id          UUID REFERENCES influencer_scores(id) ON DELETE SET NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_influencer_analyses_user_id ON influencer_analyses(user_id);
CREATE INDEX IF NOT EXISTS idx_influencer_analyses_score_id ON influencer_analyses(score_id);
CREATE INDEX IF NOT EXISTS idx_influencer_analyses_created_at ON influencer_analyses(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS influencer_analyses;
DROP TABLE IF EXISTS influencer_scores;
