-- +goose Up
CREATE TABLE IF NOT EXISTS app_endpoint_policies (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    endpoint_id         UUID NOT NULL REFERENCES app_endpoints(id) ON DELETE CASCADE,
    required_permission VARCHAR(255),
    rate_limit          INTEGER,
    auth_policy         VARCHAR(100) NOT NULL DEFAULT 'jwt',
    validation_policy   JSONB,
    extra_policies      JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_endpoint_policies_endpoint_id ON app_endpoint_policies(endpoint_id);

-- +goose Down
DROP TABLE IF EXISTS app_endpoint_policies;
