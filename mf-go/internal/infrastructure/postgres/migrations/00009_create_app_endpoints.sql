-- +goose Up
CREATE TABLE IF NOT EXISTS app_endpoints (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id          UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    method          VARCHAR(10) NOT NULL,
    path            VARCHAR(500) NOT NULL,
    version         VARCHAR(20) NOT NULL DEFAULT 'v1',
    backend_service VARCHAR(255) NOT NULL,
    backend_action  VARCHAR(255) NOT NULL,
    schema          JSONB,
    audit_level     VARCHAR(50) DEFAULT 'standard',
    pii_masking     BOOLEAN NOT NULL DEFAULT FALSE,
    event_after     VARCHAR(255),
    status          VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(app_id, method, path, version)
);

CREATE INDEX idx_endpoints_app_id ON app_endpoints(app_id);
CREATE INDEX idx_endpoints_status ON app_endpoints(status);

-- +goose Down
DROP TABLE IF EXISTS app_endpoints;
