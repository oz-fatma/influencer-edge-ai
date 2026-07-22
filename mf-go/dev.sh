#!/usr/bin/env bash
#
# dev.sh — Development runner for masterfabric-go
#
# Features:
#   - Starts Docker services (Postgres, Redis, Kafka, Kafka UI)
#   - Waits for services to become healthy
#   - Runs database migrations (Up only)
#   - Ensures Kafka topics exist
#   - Starts the Go server with hot-reload (air)
#
# Usage:
#   ./dev.sh          — Full startup (infra + migrations + hot-reload)
#   ./dev.sh server   — Hot-reload server only (skip infra)
#   ./dev.sh infra    — Start infra only (no server)
#   ./dev.sh migrate  — Run migrations only
#   ./dev.sh down     — Stop all Docker services
#   ./dev.sh logs     — Tail Docker service logs
#   ./dev.sh clean    — Stop infra, remove volumes, clean build artifacts
#
set -euo pipefail

# ─── Configuration ────────────────────────────────────────────────────────────
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/deployments/docker-compose.yml"
MIGRATION_DIR="$PROJECT_ROOT/internal/infrastructure/postgres/migrations"
DB_USER="masterfabric"
DB_NAME="masterfabric"
DB_CONTAINER="masterfabric-postgres"

# Build env — workaround for macOS sandbox permissions on /var/folders
export GOTMPDIR="$PROJECT_ROOT/tmp"
export GOCACHE="$PROJECT_ROOT/tmp/go-cache"
export TMPDIR="$PROJECT_ROOT/tmp"
export CGO_ENABLED=0

# Kafka env
export KAFKA_ENABLED=true

# ─── Colors ───────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

log_info()  { echo -e "${CYAN}[INFO]${NC}  $*"; }
log_ok()    { echo -e "${GREEN}[OK]${NC}    $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }
log_step()  { echo -e "\n${BOLD}━━━ $* ━━━${NC}"; }

# ─── Helpers ──────────────────────────────────────────────────────────────────

ensure_dirs() {
    mkdir -p "$PROJECT_ROOT/tmp"
    mkdir -p "$PROJECT_ROOT/bin"
}

check_docker() {
    if ! docker info &>/dev/null; then
        log_warn "Docker is not running. Attempting to start Docker Desktop..."
        open -a Docker 2>/dev/null || true
        local retries=30
        for i in $(seq 1 $retries); do
            docker info &>/dev/null && break
            printf "  waiting for Docker daemon... (%d/%d)\r" "$i" "$retries"
            sleep 2
        done
        echo ""
        if ! docker info &>/dev/null; then
            log_error "Docker daemon failed to start. Please start Docker manually."
            exit 1
        fi
        log_ok "Docker is ready"
    fi
}

install_air() {
    if ! command -v air &>/dev/null; then
        log_info "Installing air (hot-reload tool)..."
        GOBIN="$PROJECT_ROOT/bin" go install github.com/air-verse/air@latest 2>/dev/null \
            || go install github.com/air-verse/air@latest 2>/dev/null
        if [[ -f "$PROJECT_ROOT/bin/air" ]]; then
            export PATH="$PROJECT_ROOT/bin:$PATH"
        fi
    fi

    if ! command -v air &>/dev/null && [[ -f "$PROJECT_ROOT/bin/air" ]]; then
        export PATH="$PROJECT_ROOT/bin:$PATH"
    fi

    if ! command -v air &>/dev/null; then
        log_error "Failed to install air. Install manually: go install github.com/air-verse/air@latest"
        exit 1
    fi
    log_ok "air is available: $(which air)"
}

# ─── Infrastructure ───────────────────────────────────────────────────────────

start_infra() {
    log_step "Starting infrastructure"
    check_docker

    log_info "Starting Docker Compose services..."
    docker compose -f "$COMPOSE_FILE" up -d

    log_info "Waiting for services to become healthy..."
    local services=("masterfabric-postgres" "masterfabric-redis" "masterfabric-kafka")
    for svc in "${services[@]}"; do
        local retries=30
        for i in $(seq 1 $retries); do
            local health
            health=$(docker inspect --format='{{.State.Health.Status}}' "$svc" 2>/dev/null || echo "missing")
            if [[ "$health" == "healthy" ]]; then
                log_ok "$svc is healthy"
                break
            fi
            if [[ $i -eq $retries ]]; then
                log_warn "$svc did not become healthy (status: $health)"
            fi
            sleep 2
        done
    done

    echo ""
    docker compose -f "$COMPOSE_FILE" ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
    echo ""
}

stop_infra() {
    log_step "Stopping infrastructure"
    check_docker
    docker compose -f "$COMPOSE_FILE" down
    log_ok "All services stopped"
}

# ─── Migrations ───────────────────────────────────────────────────────────────

run_migrations() {
    log_step "Running database migrations"

    local count=0
    for f in "$MIGRATION_DIR"/0*.sql; do
        [[ -f "$f" ]] || continue
        local fname
        fname=$(basename "$f")
        # Extract only the UP part (between "-- +goose Up" and "-- +goose Down")
        local sql
        sql=$(sed -n '/^-- +goose Up$/,/^-- +goose Down$/p' "$f" | sed '1d;$d')
        if [[ -n "$sql" ]]; then
            echo "$sql" | docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -q 2>/dev/null \
                && log_ok "  $fname" \
                || log_warn "  $fname (may already exist)"
            count=$((count + 1))
        fi
    done

    log_ok "Processed $count migration files"
}

# ─── Server ───────────────────────────────────────────────────────────────────

start_server() {
	log_step "Starting server with hot-reload"

	if [[ -f "$PROJECT_ROOT/.env" ]]; then
		set -a
		# shellcheck disable=SC1091
		source "$PROJECT_ROOT/.env"
		set +a
		log_ok "Loaded $PROJECT_ROOT/.env"
	fi

	local port="${SERVER_PORT:-8080}"

	# Kill any process on the configured port
	lsof -ti:"$port" 2>/dev/null | xargs kill -9 2>/dev/null || true
	sleep 1

	install_air

	log_info "Watching for file changes..."
	log_info "Server will be available at ${BOLD}http://localhost:${port}${NC}"
	log_info "Kafka UI available at ${BOLD}http://localhost:${KAFKA_UI_PORT:-8091}${NC}"
	log_info "Prometheus metrics at ${BOLD}http://localhost:${port}/metrics${NC}"
	echo ""

	cd "$PROJECT_ROOT"
	export PATH="$PROJECT_ROOT/bin:$PATH"
	export SERVER_PORT="$port"
	exec air -c .air.toml
}

# ─── Commands ─────────────────────────────────────────────────────────────────

cmd_full() {
    ensure_dirs
    start_infra
    run_migrations
    start_server
}

cmd_server() {
    ensure_dirs
    start_server
}

cmd_infra() {
    ensure_dirs
    start_infra
    run_migrations
    echo ""
    log_ok "Infrastructure is ready. Run ${BOLD}./dev.sh server${NC} to start the app."
}

cmd_migrate() {
    run_migrations
}

cmd_down() {
    stop_infra
}

cmd_logs() {
    docker compose -f "$COMPOSE_FILE" logs -f --tail=50
}

cmd_clean() {
    log_step "Cleaning up"
    check_docker
    docker compose -f "$COMPOSE_FILE" down -v 2>/dev/null || true
    rm -rf "$PROJECT_ROOT/tmp" "$PROJECT_ROOT/bin/server" "$PROJECT_ROOT/.tmp"
    log_ok "Cleaned: Docker volumes, tmp/, bin/server"
}

cmd_help() {
    echo -e "${BOLD}masterfabric-go development runner${NC}"
    echo ""
    echo "Usage: ./dev.sh [command]"
    echo ""
    echo "Commands:"
    echo -e "  ${GREEN}(default)${NC}   Full startup: infra + migrations + hot-reload server"
    echo -e "  ${GREEN}server${NC}      Start hot-reload server only (infra must be running)"
    echo -e "  ${GREEN}infra${NC}       Start infrastructure only (Postgres, Redis, Kafka)"
    echo -e "  ${GREEN}migrate${NC}     Run database migrations"
    echo -e "  ${GREEN}down${NC}        Stop all Docker services"
    echo -e "  ${GREEN}logs${NC}        Tail Docker service logs"
    echo -e "  ${GREEN}clean${NC}       Stop infra, remove volumes, clean build artifacts"
    echo -e "  ${GREEN}help${NC}        Show this help message"
    echo ""
    echo "Environment:"
    echo "  KAFKA_ENABLED=true   (default: true)"
    echo "  DB_DSN=postgres://masterfabric:masterfabric@localhost:5432/masterfabric?sslmode=disable"
    echo ""
    echo "Endpoints (when running):"
    echo "  API:        http://localhost:8080"
    echo "  Health:     http://localhost:8080/health/ready"
    echo "  Metrics:    http://localhost:8080/metrics"
    echo "  Kafka UI:   http://localhost:8090"
}

# ─── Main ─────────────────────────────────────────────────────────────────────

case "${1:-}" in
    server)  cmd_server  ;;
    infra)   cmd_infra   ;;
    migrate) cmd_migrate ;;
    down)    cmd_down    ;;
    logs)    cmd_logs    ;;
    clean)   cmd_clean   ;;
    help|-h|--help) cmd_help ;;
    *)       cmd_full    ;;
esac
