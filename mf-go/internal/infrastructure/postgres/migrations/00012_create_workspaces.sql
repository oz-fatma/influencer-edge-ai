-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'archived')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, slug)
);

CREATE INDEX idx_workspaces_organization_id ON workspaces(organization_id);
CREATE INDEX idx_workspaces_slug ON workspaces(slug);
CREATE INDEX idx_workspaces_status ON workspaces(status);

-- Add workspace_id column to apps table (nullable for backward compatibility)
ALTER TABLE apps ADD COLUMN IF NOT EXISTS workspace_id UUID REFERENCES workspaces(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_apps_workspace_id ON apps(workspace_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_apps_workspace_id;
ALTER TABLE apps DROP COLUMN IF EXISTS workspace_id;
DROP INDEX IF EXISTS idx_workspaces_status;
DROP INDEX IF EXISTS idx_workspaces_slug;
DROP INDEX IF EXISTS idx_workspaces_organization_id;
DROP TABLE IF EXISTS workspaces;
-- +goose StatementEnd
