# masterfabric-go

<div align="center">

![masterfabric-go Banner](.github/images/banner.png)

![Version](https://img.shields.io/badge/version-0.0.1-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.26.4-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-AGPL--v3.0-green.svg)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)
![Kafka](https://img.shields.io/badge/kafka-enabled-orange.svg?logo=apache-kafka)

**Enterprise-grade, multi-tenant, RBAC-driven SaaS backend platform built with Go and clean/hexagonal architecture.**

[🚀 Quick Start](#quick-start) • [📚 Documentation](#architecture) • [🔒 Security](#security-hardening) • [🤝 Contributing](CONTRIBUTING.md) • [📄 License](LICENSE)

</div>

---

## Architecture

- **Domain-Driven Design** with bounded contexts (IAM, Tenant, API Management, Audit)
- **Clean Architecture**: domain layer has zero external dependencies
- **Phase 1 Modular Monolith**: single binary, ready for service extraction

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.26.4 |
| HTTP Router | Chi |
| Database | PostgreSQL 16 (via pgx) |
| Cache | Redis 7 |
| Message Queue | Apache Kafka (via segmentio/kafka-go) |
| Migrations | goose |
| Auth | JWT (golang-jwt) + bcrypt |
| Observability | OpenTelemetry + Prometheus |
| Logging | slog (structured JSON) |
| Validation | go-playground/validator |

## Quick Start

### Prerequisites

- Go 1.26.4+
- Docker & Docker Compose
- (Optional) `goose` CLI for manual migration management

### Option 1: Development Mode (Recommended)

Use the `dev.sh` script for hot-reload development:

```bash
# Full startup: infrastructure + migrations + hot-reload server
./dev.sh

# Or step-by-step:
./dev.sh infra      # Start Docker services + run migrations
./dev.sh server     # Start hot-reload server (infra must be running)
```

The `dev.sh` script:
- ✅ Starts Docker services (Postgres, Redis, Kafka, Kafka UI)
- ✅ Waits for services to become healthy
- ✅ Runs database migrations automatically
- ✅ Starts the server with **hot-reload** (auto-restarts on file changes)
- ✅ Auto-installs `air` (hot-reload tool) if needed

**Hot-reload**: Edit any `.go` file and save — the server automatically rebuilds and restarts (~3s).

### Option 2: Manual Setup

```bash
# 1. Start infrastructure
make docker-up

# 2. Run migrations
make migrate

# 3. Run server
make run
```

The server starts on `http://localhost:8080`.

### Verify

```bash
curl http://localhost:8080/health/live
# {"status":"alive"}

curl http://localhost:8080/health/ready
# {"status":"ready","services":{"postgres":"healthy","redis":"healthy"}}
# On failure, service entries show "unhealthy" without internal error details
```

### Development Scripts

```bash
./dev.sh            # Full startup (infra + migrations + hot-reload)
./dev.sh server     # Hot-reload server only (skip infra)
./dev.sh infra      # Start infrastructure only
./dev.sh migrate    # Run migrations only
./dev.sh down       # Stop all Docker services
./dev.sh logs       # Tail Docker service logs
./dev.sh clean      # Stop infra, remove volumes, clean artifacts
./dev.sh help       # Show help
```

## API Endpoints

### Auth (public)
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login and receive JWT

### Users (authenticated + RBAC)
- `GET /api/v1/me` - Get current user
- `GET /api/v1/users` - List users (paginated) — requires `user:read`
- `GET /api/v1/users/{id}` - Get user by ID — requires `user:read`
- `POST /api/v1/roles/assign` - Assign role to user — requires `user:write`

### Organizations (authenticated + RBAC)
- `POST /api/v1/organizations` - Create organization — requires `org:write`
- `GET /api/v1/organizations` - List organizations — requires `org:read`
- `GET /api/v1/organizations/{orgId}` - Get organization — requires `org:read`

### Apps (authenticated + RBAC)
- `POST /api/v1/organizations/{orgId}/apps` - Create app — requires `app:write`
- `GET /api/v1/organizations/{orgId}/apps` - List apps — requires `app:read`
- `GET /api/v1/organizations/{orgId}/apps/{appId}` - Get app — requires `app:read`

### API Keys (authenticated + RBAC)
- `POST /api/v1/organizations/{orgId}/apps/{appId}/keys` - Create API key — requires `app:write`
- `GET /api/v1/organizations/{orgId}/apps/{appId}/keys` - List API keys — requires `app:read`
- `DELETE /api/v1/organizations/{orgId}/apps/{appId}/keys/{keyId}` - Revoke key — requires `app:write`

### Endpoints (authenticated + RBAC)
- `POST /api/v1/organizations/{orgId}/apps/{appId}/endpoints` - Define endpoint — requires `endpoint:write`
- `GET /api/v1/organizations/{orgId}/apps/{appId}/endpoints` - List endpoints — requires `endpoint:read`
- `GET /api/v1/organizations/{orgId}/apps/{appId}/endpoints/{endpointId}` - Get endpoint — requires `endpoint:read`
- `POST /api/v1/organizations/{orgId}/apps/{appId}/endpoints/{endpointId}/retire` - Retire endpoint — requires `endpoint:write`
- `PUT /api/v1/organizations/{orgId}/apps/{appId}/endpoints/{endpointId}/policy` - Update policy — requires `endpoint:write`
- `GET /api/v1/organizations/{orgId}/apps/{appId}/endpoints/{endpointId}/policy` - Get policy — requires `endpoint:read`

### Audit Logs (authenticated + RBAC)
- `GET /api/v1/organizations/{orgId}/audit-logs` - Org audit logs — requires `org:read`
- `GET /api/v1/users/{userId}/audit-logs` - User audit logs — requires `org:read`

### Observability
- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe
- `GET /metrics` - Prometheus metrics

### Real-time (WebSocket)
- `GET /api/v1/ws?token=<jwt>` - WebSocket upgrade for live domain event delivery

**Required headers:** `X-Organization-ID`, `X-App-ID`  
**Auth:** JWT via `?token=` query parameter (browser-friendly) or `Authorization: Bearer` header  
**Permission:** `app:read`

```javascript
const ws = new WebSocket(
  "ws://localhost:8080/api/v1/ws?token=" + jwt,
  [],
);
// Set headers via a WS client library; browsers require ?token= query param
```

**Client protocol:**

```json
{ "action": "subscribe",   "channel": "api-management" }
{ "action": "unsubscribe", "channel": "tenant" }
{ "action": "ping" }
```

**Server push:**

```json
{
  "type": "endpoint.created",
  "topic": "masterfabric.api-management",
  "organization_id": "uuid",
  "app_id": "uuid",
  "data": { ... },
  "timestamp": "2026-07-03T12:00:00Z"
}
```

Architecture details: [`docs/WEBSOCKET.md`](docs/WEBSOCKET.md)

## Postman Collection

A complete Postman collection with **37 requests** and **auto-capturing scripts** is available:

- **Collection**: `postman/masterfabric-go.postman_collection.json`
- **Environment**: `postman/masterfabric-go-local.postman_environment.json`

### Features

- ✅ **Auto-capture JWT token** from Login → automatically used in all subsequent requests
- ✅ **Auto-capture IDs**: `user_id`, `org_id`, `app_id`, `endpoint_id`, `api_key_id` from responses
- ✅ **Variables persist** across sessions (saved to environment)
- ✅ **Test assertions** on every request (status codes, response validation)
- ✅ **Negative test cases** (unauthorized, validation errors, not found)

### Usage

1. Import both files into Postman
2. Select the **"MasterFabric Go - Local"** environment
3. Run **Login** → token is automatically saved
4. Run **Create Organization** → `org_id` is auto-captured
5. Run **Create App** → `app_id` is auto-captured
6. All subsequent requests use the captured variables automatically

**Endpoints covered**: Health, Auth, Users, Organizations, Apps, API Keys, Endpoints, Policies, RBAC, Audit Logs, Error Scenarios, **Invoke Defined Endpoints**.

### How to Use Defined Endpoints

After defining an endpoint (e.g., `POST /orders` or `GET /products`), you can invoke it through the API gateway:

**Required Headers:**
- `X-App-ID`: Your application ID (triggers gateway pipeline)
- `X-Organization-ID`: Your organization ID
- `Authorization: Bearer <jwt_token>`: JWT token for authenticated requests

**Example: Invoke GET /products**
```http
GET /api/v1/products
Headers:
  X-App-ID: <your-app-id>
  X-Organization-ID: <your-org-id>
  Authorization: Bearer <jwt-token>
```

**Example: Invoke POST /orders**
```http
POST /api/v1/orders
Headers:
  X-App-ID: <your-app-id>
  X-Organization-ID: <your-org-id>
  Authorization: Bearer <jwt-token>
  Content-Type: application/json
Body:
  {
    "product_id": "prod-123",
    "quantity": 2
  }
```

**Gateway Pipeline Flow:**
1. Gateway checks `X-App-ID` header
2. Looks up endpoint by method + path
3. Validates JSON schema (if defined)
4. Checks RBAC permissions (if policy requires)
5. Enforces rate limits
6. Applies interceptors (PII masking, transformations)
7. Routes to backend service

See the **"Invoke Defined Endpoints"** section in the Postman collection for complete examples including error scenarios.

## Security Hardening

A full security remediation pass was applied on the `security/hardening` branch. The goal was to close confirmed audit findings across the shared platform layer, HTTP surface, deployment defaults, and authorization coverage — without changing the public API contract.

For the complete trust model, accepted risks, and the **Security Controls Registry v0.1**, see **[SECURITY.md](SECURITY.md)**.

### Why these changes were made

| Area | Problem | Fix | Rationale |
|------|---------|-----|-----------|
| **Toolchain & dependencies** | Outdated Go stdlib and library versions carried known CVE advisories | Bumped to **Go 1.26.4**; refreshed pgx, chi, validator, and `golang.org/x/*` modules | Closes upstream vulnerability reports at the root cause rather than patching symptoms |
| **Container images** | Builder/runtime Go mismatch; EOL Alpine; process ran as root | Aligned builder to Go 1.26.4, runtime to **alpine 3.24**, dedicated **non-root** `appuser` | Reduces container escape blast radius and keeps build/runtime toolchains consistent |
| **Local compose defaults** | Postgres, Redis, and Kafka exposed on `0.0.0.0` with weak default credentials | Ports bind to **loopback** (`127.0.0.1`) by default via `*_HOST_BIND` env vars | Prevents accidental credential exposure on shared or public networks during local development |
| **5xx error responses** | `response.Error` returned `err.Error()` verbatim, leaking DB/driver details | Generic client message (`an internal error occurred`); full detail logged server-side | Stops internal infrastructure information from reaching untrusted API consumers (CWE-209) |
| **Database DSN** | Connection string built with `fmt.Sprintf`, breaking on special characters in passwords | Credentials escaped via **`net/url`** | Prevents credential parsing errors and host/db shifting when passwords contain `@`, `:`, `?`, `#`, or `%` (CWE-116) |
| **Pagination** | Unbounded `page` query param could overflow into a negative SQL `OFFSET` | `page` clamped to **`MaxPage`** (1,000,000) | Blocks integer overflow that could return unintended rows (CWE-190) |
| **Config parsing** | `DB_MAX_CONNS` / `DB_MIN_CONNS` cast from `int` to `int32` without bounds check | Dedicated **`envOrDefaultInt32`** with 32-bit parse | Prevents silent truncation flagged by static analysis (gosec G115) |
| **CORS** | `AllowedOrigins: ["*"]` combined with `AllowCredentials: true` — an invalid and unsafe combination | Configurable **`CORS_ALLOWED_ORIGINS`** allow-list; credentials auto-disabled for wildcard or empty list | Stops browsers from accepting overly permissive cross-origin credential flows (CWE-942) |
| **Request body size** | No global body limit — large payloads could exhaust server memory | **`MAX_BODY_BYTES`** middleware (default **1 MiB**) using `http.MaxBytesReader` | Mitigates memory exhaustion from oversized JSON uploads (CWE-400) |
| **Readiness probe** | `/health/ready` echoed raw Postgres/Redis error strings | Returns generic **`unhealthy`** markers; logs detail with `slog` | Health endpoints are often public; they must not disclose hostnames or connection errors (CWE-209) |
| **Outbound HTTP proxy** | Default `http.Client` followed redirects and had no timeout, risking custom header leakage | **No redirect following**, 30s timeout, response body capped at 1 MiB | Prevents `Authorization` or service tokens from being forwarded across hosts on redirect (CWE-522) |
| **RBAC coverage** | JWT was required but any authenticated user could call admin routes; wildcard permissions in seed data were not honored | **`RequirePermission`** on all admin routes; wildcard-aware matching (`*`, `org:*`, `*:read`) | Ensures state-changing operations require explicit grants, not just a valid token (CWE-306) |
| **Migration script** | `migrate.sh create NAME` did not sanitize `NAME`, allowing path traversal in filenames | Name restricted to **`[a-zA-Z0-9_]`** | Blocks `../` injection when migration files are created via automation (CWE-22) |
| **JWT secret default** | Server started silently with `change-me-in-production` | **Startup warning** when the default signing secret is detected | Makes misconfiguration visible before production exposure |
| **Gateway proxy (gosec G704)** | Intentional SSRF sink for operator-configured backend URLs | Documented as an **accepted risk** in SECURITY.md with audited `#nosec` suppressions | Proxying is a core gateway feature; risk is bounded by RBAC on endpoint creation |

### Verification

Run these checks before merging or deploying:

```bash
go build ./... && go vet ./... && go test ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
go run github.com/securego/gosec/v2/cmd/gosec@latest -quiet ./...
```

Expected results on the hardened branch:

- All tests pass
- `govulncheck`: no vulnerabilities found
- `gosec`: clean (2 intentional, documented suppressions for the gateway HTTP proxy)

### Production checklist

Before exposing the API on a production network:

1. Set a strong, random **`JWT_SECRET`** (never use the default)
2. Set explicit **`CORS_ALLOWED_ORIGINS`** (avoid `*`)
3. Enable **`DB_SSLMODE=require`** (or stricter)
4. Restrict **`/metrics`** and **`/health/*`** at the network edge
5. Replace default database credentials in any non-local deployment

## Configuration

All configuration is via environment variables with sensible defaults:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | `0.0.0.0` | Bind host |
| `SERVER_PORT` | `8080` | Bind port |
| `SERVER_READ_TIMEOUT_SECONDS` | `15` | HTTP read timeout |
| `SERVER_WRITE_TIMEOUT_SECONDS` | `15` | HTTP write timeout |
| `SERVER_IDLE_TIMEOUT_SECONDS` | `60` | HTTP idle timeout |
| `MAX_BODY_BYTES` | `1048576` | Maximum request body size (1 MiB) |
| `CORS_ALLOWED_ORIGINS` | *(empty)* | Comma-separated allowed CORS origins; credentials disabled when empty or `*` |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `masterfabric` | PostgreSQL user |
| `DB_PASSWORD` | `masterfabric` | PostgreSQL password |
| `DB_NAME` | `masterfabric` | PostgreSQL database |
| `DB_SSLMODE` | `disable` | PostgreSQL SSL mode |
| `DB_MAX_CONNS` | `25` | PostgreSQL connection pool max size |
| `DB_MIN_CONNS` | `5` | PostgreSQL connection pool min size |
| `DB_HOST_BIND` | `127.0.0.1` | Docker Compose host bind for Postgres (dev only) |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_HOST_BIND` | `127.0.0.1` | Docker Compose host bind for Redis (dev only) |
| `KAFKA_ENABLED` | `false` | Enable Kafka event bus |
| `KAFKA_BROKERS` | `localhost:9092` | Kafka broker addresses (comma-separated) |
| `KAFKA_GROUP_ID` | `masterfabric-go` | Kafka consumer group ID |
| `KAFKA_NUM_PARTITIONS` | `3` | Default partitions for auto-created topics |
| `KAFKA_REPLICATION_FACTOR` | `1` | Replication factor for auto-created topics |
| `KAFKA_HOST_BIND` | `127.0.0.1` | Docker Compose host bind for Kafka (dev only) |
| `JWT_SECRET` | `change-me-in-production` | JWT signing secret (**change before production**) |
| `JWT_EXPIRATION_HOURS` | `24` | JWT token lifetime |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `LOG_FORMAT` | `json` | Log format (json, text) |
| `WS_ENABLED` | `true` | Enable WebSocket endpoint |
| `WS_MAX_CONNECTIONS` | `1000` | Maximum concurrent WebSocket connections |
| `WS_PING_INTERVAL_SECONDS` | `30` | Server ping interval for keepalive |
| `WS_READ_BUFFER_SIZE` | `1024` | WebSocket read buffer size (bytes) |
| `WS_WRITE_BUFFER_SIZE` | `1024` | WebSocket write buffer size (bytes) |

## Kafka Event Bus

The project uses an `EventBus` interface (`internal/shared/events/bus.go`) that supports two implementations:

- **In-process bus** (default): channel-based, suitable for local dev and single-instance deployments
- **Kafka bus**: production-grade, uses `segmentio/kafka-go` with KRaft-mode Kafka (no Zookeeper)

### Enable Kafka

```bash
# Start infrastructure including Kafka
make docker-up

# Run with Kafka enabled
KAFKA_ENABLED=true make run
```

Kafka UI is available at `http://localhost:8090` for inspecting topics and messages.

### Topics

| Topic | Bounded Context | Events |
|-------|----------------|--------|
| `masterfabric.iam` | IAM | user.registered, user.invited, role.assigned, role.revoked |
| `masterfabric.tenant` | Tenant | organization.created, app.created, app.updated |
| `masterfabric.api-management` | API Management | endpoint.created, endpoint.updated, endpoint.retired |
| `masterfabric.audit` | Audit | (consumers write to audit_logs table) |

Topics are auto-created at startup when `KAFKA_ENABLED=true`.

### Publishing Events from Use Cases

**✅ Events are automatically published** from the following use cases:

- `RegisterUseCase` → `user.registered` (TopicIAM)
- `AssignRoleUseCase` → `role.assigned` (TopicIAM)
- `CreateOrgUseCase` → `organization.created` (TopicTenant)
- `CreateAppUseCase` → `app.created` (TopicTenant)
- `DefineEndpointUseCase` → `endpoint.created` (TopicAPIManagement)
- `RetireEndpointUseCase` → `endpoint.retired` (TopicAPIManagement)

The `EventBus` is injected into use cases at startup. Events are automatically serialized into JSON envelopes with metadata (ID, type, source, timestamp).

**Example**: When you create an organization via `POST /api/v1/organizations`, the `organization.created` event is published to Kafka topic `masterfabric.tenant`.

**Verify events**: Use Kafka UI at `http://localhost:8090` or consume directly:

```bash
docker exec masterfabric-kafka /opt/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server localhost:29092 \
  --topic masterfabric.tenant \
  --from-beginning
```

### Consuming Events

Register handlers at startup in `main.go`:

```go
eventBus.Subscribe(events.TopicIAM, func(ctx context.Context, event events.Event) error {
    log.Info("iam event", "event", event)
    return nil
})
```

## Project Structure

```
cmd/server/             - Application entry point
internal/
  shared/               - Cross-cutting concerns (config, middleware, errors, events)
  domain/               - Domain layer (entities, interfaces, domain events)
    iam/                - Identity & Access Management
    tenant/             - Tenant & App Management
    apimanagement/      - API Management
    audit/              - Audit & Observability
  application/          - Use cases and DTOs
  infrastructure/       - External implementations (postgres, redis, http)
  gateway/              - API Gateway policy pipeline
  domain/realtime/      - WebSocket room model and hub interface
  infrastructure/websocket/ - In-memory hub, event bridge, session pumps
deployments/            - Docker and deployment configs
docs/                   - Architecture documentation (WEBSOCKET.md)
```

## Scripts

The `scripts/` directory contains utility scripts for common development tasks:

### Database Scripts

```bash
# Run migrations
./scripts/migrate.sh up          # Apply all pending migrations
./scripts/migrate.sh down         # Rollback last migration
./scripts/migrate.sh status       # Show migration status
./scripts/migrate.sh create NAME  # Create new migration file

# Seed database with initial data
go run scripts/seed.go            # Seed roles and permissions
```

### Testing & Quality

```bash
# Run tests
./scripts/test.sh                 # Run all tests
./scripts/test.sh -cover          # Run with coverage report
./scripts/test.sh ./path          # Run tests in specific path

# Lint code
./scripts/lint.sh                 # Check code quality
./scripts/lint.sh -fix            # Auto-fix issues

# Security scans
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
go run github.com/securego/gosec/v2/cmd/gosec@latest -quiet ./...
```

## Make Targets

```bash
make build          # Build binary
make run            # Run the server
make test           # Run tests
make test-cover     # Run tests with coverage
make lint           # Run linter
make migrate        # Run migrations up
make migrate-down   # Rollback last migration
make docker-up      # Start Docker services (Postgres, Redis, Kafka, Kafka UI)
make docker-down    # Stop Docker services
make clean          # Clean build artifacts
```

**Note**: For development with hot-reload, use `./dev.sh` instead of `make run`.

## License

This project is licensed under the **GNU Affero General Public License v3.0 (AGPL v3.0)**.

See the [LICENSE](LICENSE) file for details.

### License Summary

- ✅ **Free to use** for personal and commercial projects
- ✅ **Modify** and distribute freely
- ⚠️ **Copyleft**: If you modify and run this software as a network service, you must make your modified source code available to users
- 📖 **Full License**: See [LICENSE](LICENSE) file

### For Commercial Use

If you need to use this software in a commercial product without the AGPL copyleft requirements, please contact us for licensing options.

---

**Copyright © 2025 MasterFabric. All rights reserved.**
