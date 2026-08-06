-- +goose Up
CREATE TABLE IF NOT EXISTS brand_profiles (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name             VARCHAR(255) NOT NULL,
    industry         VARCHAR(128) NOT NULL,
    target_audience  TEXT NOT NULL,
    budget_range     VARCHAR(64),
    brand_values     TEXT NOT NULL,
    campaign_goal    VARCHAR(128) NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_brand_profiles_user_name UNIQUE (user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_brand_profiles_user_id ON brand_profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_brand_profiles_created_at ON brand_profiles(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS brand_profiles;
