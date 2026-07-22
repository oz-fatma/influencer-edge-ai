-- +goose Up
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id     UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission  VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);

-- +goose Down
DROP TABLE IF EXISTS role_permissions;
