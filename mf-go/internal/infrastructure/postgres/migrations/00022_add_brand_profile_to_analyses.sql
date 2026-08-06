-- +goose Up
ALTER TABLE influencer_analyses
  ADD COLUMN IF NOT EXISTS brand_profile_id UUID REFERENCES brand_profiles(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE influencer_analyses
  DROP COLUMN IF EXISTS brand_profile_id;
