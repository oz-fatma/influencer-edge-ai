# Changelog

All notable changes to the masterfabric-go project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

#### Development Experience
- **Hot-reload development script** (`dev.sh`) with automatic file watching and server restart
  - Full startup: `./dev.sh` (infra + migrations + hot-reload)
  - Server only: `./dev.sh server`
  - Infrastructure only: `./dev.sh infra`
  - Additional commands: `migrate`, `down`, `logs`, `clean`, `help`
- **Air configuration** (`.air.toml`) for hot-reload with:
  - Automatic rebuild on `.go`, `.toml`, `.yaml`, `.sql`, `.env` file changes
  - macOS sandbox workaround (custom `GOTMPDIR`/`GOCACHE`)
  - Port cleanup before rebuild to avoid bind errors
  - Colored output and build status

#### Postman Collection
- **Complete Postman collection** (`postman/masterfabric-go.postman_collection.json`) with 37 requests
- **Postman environment** (`postman/masterfabric-go-local.postman_environment.json`) with pre-configured variables
- **Auto-capturing scripts**:
  - JWT token automatically saved from Login response
  - User ID, Organization ID, App ID, Endpoint ID, API Key ID automatically captured
  - Variables saved to both collection and environment for persistence
  - Pre-request validation warns if token is missing
- **Test assertions** on every request (status codes, response validation)
- **Negative test cases** (unauthorized, validation errors, not found, duplicate registration)

#### Kafka Event Publishing
- **Domain events now published to Kafka** from use cases:
  - `RegisterUseCase` → `user.registered` event
  - `AssignRoleUseCase` → `role.assigned` event
  - `CreateOrgUseCase` → `organization.created` event
  - `CreateAppUseCase` → `app.created` event
  - `DefineEndpointUseCase` → `endpoint.created` event
  - `RetireEndpointUseCase` → `endpoint.retired` event
- **Event type derivation** from Go struct names (e.g., `UserRegistered` → `user.registered`)
- **EventBus injection** into all relevant use cases via dependency injection

#### Infrastructure
- **Kafka UI** added to Docker Compose (`http://localhost:8090`)
- **Apache Kafka** using KRaft mode (no Zookeeper required)
- **Auto-topic creation** at startup when `KAFKA_ENABLED=true`

### Changed

#### Kafka Integration
- **EventBus interface** now properly wired into use case layer
- **Event type mapping** fixed (was returning "unknown", now correctly derives from struct names)
- **Kafka producer** serializes events into JSON envelopes with metadata

#### Docker Compose
- **Kafka image** changed from `bitnami/kafka:3.7` to `apache/kafka:latest` (KRaft mode)
- **Removed obsolete `version` field** from docker-compose.yml

#### Development Workflow
- **Build process** now uses custom `GOTMPDIR` and `GOCACHE` to avoid macOS sandbox issues
- **Server startup** optimized with pre-build port cleanup

### Fixed

- **macOS sandbox permissions**: Build now works reliably by using project-local temp directories
- **Port binding conflicts**: Hot-reload now cleans up port 8080 before rebuild
- **Event type derivation**: Fixed Kafka event type from "unknown" to proper dot-notation (e.g., `user.registered`)
- **EventBus wiring**: Fixed missing EventBus injection in use cases (was created but unused)

### Documentation

- **README.md** updated with:
  - `dev.sh` usage instructions
  - Hot-reload development workflow
  - Postman collection information
  - Updated Kafka event publishing documentation
  - Development script commands

---

## [0.0.1] - 2026-02-09

### Added

#### Core Architecture
- Domain-Driven Design with bounded contexts:
  - IAM (Identity & Access Management)
  - Tenant & App Management
  - API Management
  - Audit & Observability
- Clean/Hexagonal Architecture with dependency injection
- Modular monolith structure (ready for service extraction)

#### HTTP Layer
- Chi router with RESTful API endpoints
- Global middleware: RequestID, Logging, Recovery, CORS
- JWT authentication middleware
- RBAC (Role-Based Access Control) middleware
- Tenant resolution middleware (header, JWT claims, subdomain)
- Audit logging middleware

#### Database
- PostgreSQL 16 with pgx driver
- 11 SQL migrations (goose format):
  - Organizations, Users, Roles, Permissions
  - Apps, API Keys, Endpoints, Policies
  - Audit Logs
- Repository pattern with PostgreSQL implementations

#### Authentication & Authorization
- JWT-based authentication (24h expiry)
- Bcrypt password hashing
- RBAC with organization/app-level roles
- Permission caching in Redis

#### API Gateway
- Endpoint definition and management
- Policy enforcement (auth, RBAC, rate limiting)
- Schema validation
- API key authentication
- Rate limiting via Redis

#### Caching
- Redis integration for:
  - Permission caching
  - Rate limiting
  - Session management

#### Observability
- Structured JSON logging (slog)
- OpenTelemetry integration
- Prometheus metrics endpoint
- Health checks (liveness, readiness)

#### Event System
- EventBus interface (in-process and Kafka implementations)
- Domain events defined for all bounded contexts
- Kafka integration with segmentio/kafka-go
- Event envelope format with metadata

#### Configuration
- Environment-based configuration
- Sensible defaults
- Support for all infrastructure components

#### Testing
- Unit tests for core components
- Test utilities and helpers

### Infrastructure

- Docker Compose setup:
  - PostgreSQL 16
  - Redis 7
  - Apache Kafka (KRaft mode)
  - Kafka UI
- Dockerfile for production builds
- Makefile with common tasks

### Documentation

- Initial README.md with:
  - Architecture overview
  - Quick start guide
  - API endpoint documentation
  - Configuration reference
  - Project structure

---

## Version History

- **0.0.1** - Initial release with core features
- **Unreleased** - Development improvements (hot-reload, Postman collection, Kafka event publishing)
