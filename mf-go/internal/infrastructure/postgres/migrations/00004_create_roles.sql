-- +goose Up
CREATE TABLE IF NOT EXISTS roles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scope_type      VARCHAR(50) NOT NULL CHECK (scope_type IN ('organization', 'app')),
    scope_id        UUID NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(scope_type, scope_id, name)
);

CREATE INDEX idx_roles_scope ON roles(scope_type, scope_id);

-- +goose Down
DROP TABLE IF EXISTS roles;
