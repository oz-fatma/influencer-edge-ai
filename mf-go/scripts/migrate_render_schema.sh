#!/usr/bin/env bash
#
# migrate_render_schema.sh — Run mf-go migrations in a dedicated PostgreSQL schema
# on the shared Render database (influencer_edge_db) without touching legacy public.* tables.
#
# Usage (from mf-go/):
#   export DATABASE_URL='postgres://USER:PASS@HOST:5432/influencer_edge_db?sslmode=require'
#   export DB_SCHEMA=mf
#   ./scripts/migrate_render_schema.sh
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MIGRATION_DIR="$PROJECT_ROOT/internal/infrastructure/postgres/migrations"

DB_SCHEMA="${DB_SCHEMA:-mf}"

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "ERROR: DATABASE_URL is required (Render Postgres external connection string)."
  exit 1
fi

if [[ ! "$DB_SCHEMA" =~ ^[a-zA-Z_][a-zA-Z0-9_]*$ ]]; then
  echo "ERROR: invalid DB_SCHEMA: $DB_SCHEMA"
  exit 1
fi

if ! command -v psql >/dev/null 2>&1; then
  echo "ERROR: psql not found. Install PostgreSQL client (brew install libpq)."
  exit 1
fi

echo "Creating schema '$DB_SCHEMA' if needed..."
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c "CREATE SCHEMA IF NOT EXISTS \"$DB_SCHEMA\";"

echo "Running mf-go migrations in schema '$DB_SCHEMA'..."
for f in "$MIGRATION_DIR"/0*.sql; do
  [[ -f "$f" ]] || continue
  fname=$(basename "$f")
  sql=$(sed -n '/^-- +goose Up$/,/^-- +goose Down$/p' "$f" | sed '1d;$d')
  if [[ -n "$sql" ]]; then
    if printf 'SET search_path TO %s, public;\n%s\n' "$DB_SCHEMA" "$sql" | psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -q; then
      echo "  ✓ $fname"
    else
      echo "  ⚠ $fname (may already exist — check error above)"
    fi
  fi
done

echo "Done. Set DB_SCHEMA=$DB_SCHEMA on Render mf-go service and redeploy."
