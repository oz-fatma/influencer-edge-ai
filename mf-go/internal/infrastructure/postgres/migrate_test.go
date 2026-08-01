package postgres

import (
	"strings"
	"testing"
)

func TestExtractUpSQL(t *testing.T) {
	content := `-- +goose Up
CREATE TABLE IF NOT EXISTS example (id INT);

-- +goose Down
DROP TABLE IF EXISTS example;
`
	got := extractUpSQL(content)
	want := "CREATE TABLE IF NOT EXISTS example (id INT);"
	if got != want {
		t.Fatalf("extractUpSQL() = %q, want %q", got, want)
	}
}

func TestExtractUpSQL_withStatementBlocks(t *testing.T) {
	content := `-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS workspaces (id UUID PRIMARY KEY);
ALTER TABLE apps ADD COLUMN IF NOT EXISTS workspace_id UUID;
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS workspaces;
`
	got := extractUpSQL(content)
	if got == "" {
		t.Fatal("expected non-empty up SQL")
	}
	if !strings.Contains(got, "CREATE TABLE IF NOT EXISTS workspaces") {
		t.Fatalf("missing workspaces DDL: %q", got)
	}
	if !strings.Contains(got, "ALTER TABLE apps ADD COLUMN") {
		t.Fatalf("missing apps alter: %q", got)
	}
}

func TestListMigrationFiles_sorted(t *testing.T) {
	files, err := listMigrationFiles()
	if err != nil {
		t.Fatalf("listMigrationFiles: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected embedded migration files")
	}
	for i := 1; i < len(files); i++ {
		if files[i-1] >= files[i] {
			t.Fatalf("files not sorted: %q before %q", files[i-1], files[i])
		}
	}
	if files[0] != "00001_create_organizations.sql" {
		t.Fatalf("first migration = %q", files[0])
	}
}
