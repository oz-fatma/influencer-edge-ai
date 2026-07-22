-- +goose Up
CREATE TABLE IF NOT EXISTS request_logs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    method       VARCHAR(16) NOT NULL,
    path         TEXT NOT NULL,
    status_code  INTEGER NOT NULL,
    duration_ms  BIGINT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_request_logs_created_at ON request_logs(created_at);
CREATE INDEX idx_request_logs_status_code ON request_logs(status_code);
CREATE INDEX idx_request_logs_method ON request_logs(method);

-- +goose Down
DROP TABLE IF EXISTS request_logs;
