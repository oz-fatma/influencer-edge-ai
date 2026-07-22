-- +goose Up
CREATE TABLE IF NOT EXISTS audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    app_id          UUID,
    endpoint_id     UUID,
    user_id         UUID,
    request_id      VARCHAR(255),
    action          VARCHAR(255) NOT NULL,
    resource_type   VARCHAR(255),
    resource_id     VARCHAR(255),
    metadata        JSONB,
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_org_id ON audit_logs(organization_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

-- +goose Down
DROP TABLE IF EXISTS audit_logs;
