-- +goose Up
CREATE TABLE IF NOT EXISTS llm_requests (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_name    VARCHAR(128) NOT NULL,
    prompt_length INTEGER NOT NULL,
    duration_ms   BIGINT NOT NULL,
    success       BOOLEAN NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_llm_requests_created_at ON llm_requests(created_at);
CREATE INDEX idx_llm_requests_success ON llm_requests(success);
CREATE INDEX idx_llm_requests_model_name ON llm_requests(model_name);

-- +goose Down
DROP TABLE IF EXISTS llm_requests;
