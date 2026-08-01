package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

var postgresSchemaName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

const (
	migrationUpMarker   = "-- +goose Up"
	migrationDownMarker = "-- +goose Down"
)

// RunMigrations applies pending SQL migrations tracked in schema_migrations.
func RunMigrations(db *sql.DB, schema string) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}
	if schema != "" && !postgresSchemaName.MatchString(schema) {
		return fmt.Errorf("invalid db schema name: %q", schema)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := ensureMigrationsTable(ctx, db, schema); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	files, err := listMigrationFiles()
	if err != nil {
		return fmt.Errorf("list migration files: %w", err)
	}

	for _, filename := range files {
		applied, err := isMigrationApplied(ctx, db, schema, filename)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", filename, err)
		}
		if applied {
			continue
		}

		upSQL, err := readMigrationUpSQL(filename)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filename, err)
		}
		if strings.TrimSpace(upSQL) == "" {
			continue
		}

		if err := applyMigration(ctx, db, schema, filename, upSQL); err != nil {
			return fmt.Errorf("apply migration %s: %w", filename, err)
		}
	}

	return nil
}

func listMigrationFiles() ([]string, error) {
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)
	return files, nil
}

func readMigrationUpSQL(filename string) (string, error) {
	content, err := migrationFS.ReadFile(path.Join("migrations", filename))
	if err != nil {
		return "", err
	}
	return extractUpSQL(string(content)), nil
}

func extractUpSQL(content string) string {
	upIdx := strings.Index(content, migrationUpMarker)
	downIdx := strings.Index(content, migrationDownMarker)
	if upIdx < 0 || downIdx < 0 || downIdx <= upIdx {
		return ""
	}

	body := content[upIdx+len(migrationUpMarker) : downIdx]
	return strings.TrimSpace(body)
}

func ensureMigrationsTable(ctx context.Context, db *sql.DB, schema string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if schema != "" {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS "%s"`, schema)); err != nil {
			return err
		}
		if err := setSearchPath(ctx, tx, schema); err != nil {
			return err
		}
	}

	const ddl = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    id         SERIAL PRIMARY KEY,
    filename   VARCHAR(255) NOT NULL UNIQUE,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`
	if _, err := tx.ExecContext(ctx, ddl); err != nil {
		return err
	}

	return tx.Commit()
}

func isMigrationApplied(ctx context.Context, db *sql.DB, schema, filename string) (bool, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	if err := setSearchPath(ctx, tx, schema); err != nil {
		return false, err
	}

	var exists bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)`, filename).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func applyMigration(ctx context.Context, db *sql.DB, schema, filename, upSQL string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := setSearchPath(ctx, tx, schema); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, upSQL); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (filename) VALUES ($1)`, filename); err != nil {
		return err
	}

	return tx.Commit()
}

func setSearchPath(ctx context.Context, tx *sql.Tx, schema string) error {
	if schema == "" {
		return nil
	}
	_, err := tx.ExecContext(ctx, fmt.Sprintf(`SET search_path TO "%s", public`, schema))
	return err
}
