-- +goose Up
ALTER TABLE influencer_scores
  ADD COLUMN IF NOT EXISTS niche            VARCHAR(128),
  ADD COLUMN IF NOT EXISTS audience_geo     VARCHAR(64),
  ADD COLUMN IF NOT EXISTS audience_demo    VARCHAR(64),
  ADD COLUMN IF NOT EXISTS follower_range   VARCHAR(32),
  ADD COLUMN IF NOT EXISTS engagement_rate  DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS content_formats  TEXT[] DEFAULT '{}';

-- +goose Down
ALTER TABLE influencer_scores
  DROP COLUMN IF EXISTS content_formats,
  DROP COLUMN IF EXISTS engagement_rate,
  DROP COLUMN IF EXISTS follower_range,
  DROP COLUMN IF EXISTS audience_demo,
  DROP COLUMN IF EXISTS audience_geo,
  DROP COLUMN IF EXISTS niche;
