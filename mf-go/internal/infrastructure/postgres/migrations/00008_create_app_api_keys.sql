-- +goose Up
CREATE TABLE IF NOT EXISTS app_api_keys (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id      UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    key_hash    VARCHAR(255) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    scopes      JSONB,
    expires_at  TIMESTAMPTZ,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_keys_app_id ON app_api_keys(app_id);
CREATE INDEX idx_api_keys_hash ON app_api_keys(key_hash);

-- +goose Down
DROP TABLE IF EXISTS app_api_keys;
