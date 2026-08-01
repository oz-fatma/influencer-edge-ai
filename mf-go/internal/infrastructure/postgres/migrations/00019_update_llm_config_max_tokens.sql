-- +goose Up
UPDATE llm_config SET max_tokens = 400 WHERE max_tokens = 100;

ALTER TABLE llm_config ALTER COLUMN max_tokens SET DEFAULT 400;

-- +goose Down
ALTER TABLE llm_config ALTER COLUMN max_tokens SET DEFAULT 100;

UPDATE llm_config SET max_tokens = 100 WHERE max_tokens = 400;
