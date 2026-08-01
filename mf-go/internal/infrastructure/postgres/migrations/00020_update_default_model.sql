-- +goose Up
UPDATE llm_config SET model = 'tgi' WHERE model = 'gemma-influencer-ft';

ALTER TABLE llm_config ALTER COLUMN model SET DEFAULT 'tgi';

-- +goose Down
ALTER TABLE llm_config ALTER COLUMN model SET DEFAULT 'gemma-influencer-ft';

UPDATE llm_config SET model = 'gemma-influencer-ft' WHERE model = 'tgi';
