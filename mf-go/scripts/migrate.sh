#!/usr/bin/env bash
#
# migrate.sh - Database migration helper script
#
# Usage:
#   ./scripts/migrate.sh up      - Run all pending migrations
#   ./scripts/migrate.sh down     - Rollback last migration
#   ./scripts/migrate.sh status   - Show migration status
#   ./scripts/migrate.sh create NAME - Create new migration file
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MIGRATION_DIR="$PROJECT_ROOT/internal/infrastructure/postgres/migrations"
DB_DSN="${DB_DSN:-postgres://masterfabric:masterfabric@localhost:5432/masterfabric?sslmode=disable}"

cd "$PROJECT_ROOT"

case "${1:-}" in
    up)
        echo "🔄 Running migrations..."
        for f in "$MIGRATION_DIR"/0*.sql; do
            [[ -f "$f" ]] || continue
            fname=$(basename "$f")
            sql=$(sed -n '/^-- +goose Up$/,/^-- +goose Down$/p' "$f" | sed '1d;$d')
            if [[ -n "$sql" ]]; then
                echo "$sql" | docker exec -i masterfabric-postgres psql -U masterfabric -d masterfabric -q 2>/dev/null \
                    && echo "  ✓ $fname" \
                    || echo "  ⚠ $fname (may already exist)"
            fi
        done
        echo "✅ Migrations complete"
        ;;
    down)
        echo "⬇️  Rolling back last migration..."
        last_file=$(ls -t "$MIGRATION_DIR"/0*.sql 2>/dev/null | head -1)
        if [[ -z "$last_file" ]]; then
            echo "❌ No migrations found"
            exit 1
        fi
        sql=$(sed -n '/^-- +goose Down$/,/^-- +goose Up$/p' "$last_file" | sed '1d;$d' || sed -n '/^-- +goose Down$/,/^$/p' "$last_file" | sed '1d')
        if [[ -n "$sql" ]]; then
            echo "$sql" | docker exec -i masterfabric-postgres psql -U masterfabric -d masterfabric -q
            echo "✅ Rolled back: $(basename "$last_file")"
        else
            echo "⚠️  No down migration found in $(basename "$last_file")"
        fi
        ;;
    status)
        echo "📊 Migration status:"
        docker exec masterfabric-postgres psql -U masterfabric -d masterfabric -c "
            SELECT 
                CASE WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'goose_db_version')
                THEN (SELECT version_id FROM goose_db_version ORDER BY id DESC LIMIT 1)
                ELSE 'No migrations table'
                END as current_version;
        " 2>/dev/null || echo "  (Run migrations first)"
        ;;
    create)
        if [[ -z "${2:-}" ]]; then
            echo "❌ Usage: $0 create MIGRATION_NAME"
            exit 1
        fi
        name="${2}"
        if [[ ! "$name" =~ ^[a-zA-Z0-9_]+$ ]]; then
            echo "❌ Migration name may only contain letters, numbers, and underscores"
            exit 1
        fi
        timestamp=$(date +%Y%m%d%H%M%S)
        filename="${MIGRATION_DIR}/${timestamp}_${name}.sql"
        cat > "$filename" <<EOF
-- +goose Up
-- +goose StatementBegin
-- Add your migration SQL here
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Add your rollback SQL here
-- +goose StatementEnd
EOF
        echo "✅ Created migration: $(basename "$filename")"
        ;;
    *)
        echo "Usage: $0 {up|down|status|create NAME}"
        exit 1
        ;;
esac
