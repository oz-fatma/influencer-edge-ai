-- +goose Up
CREATE TABLE IF NOT EXISTS llm_config (
    id            UUID PRIMARY KEY DEFAULT '00000000-0000-0000-0000-000000000001',
    system_prompt TEXT NOT NULL,
    temperature   DOUBLE PRECISION NOT NULL DEFAULT 0.1,
    max_tokens    INTEGER NOT NULL DEFAULT 100,
    model         VARCHAR(128) NOT NULL DEFAULT 'gemma-influencer-ft',
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by    UUID REFERENCES users(id) ON DELETE SET NULL
);

INSERT INTO llm_config (id, system_prompt, temperature, max_tokens, model)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'You are an expert influencer marketing analyst. ONLY return valid JSON. No markdown, no explanation, no code fences.',
    0.1,
    100,
    'gemma-influencer-ft'
)
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS llm_config;
