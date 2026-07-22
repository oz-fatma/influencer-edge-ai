# MasterFabric Go - Agent Conventions

## Project Overview

**masterfabric_go** is an enterprise-grade, multi-tenant, event-driven Go platform with IAM, tenant management, API management, gateway enforcement, dynamic CRUD handlers, and observability.

## Architecture

- **Pattern**: Domain-Driven Design (DDD) with Clean/Hexagonal Architecture
- **Language**: Go 1.22+
- **Router**: chi (go-chi/chi/v5)
- **Database**: PostgreSQL via pgx/v5 + pgxpool
- **Cache**: Redis via go-redis/v9
- **Events**: Kafka via segmentio/kafka-go
- **Auth**: JWT (golang-jwt/jwt/v5)
- **Observability**: slog (structured logging), Prometheus metrics

## Project Structure

```
cmd/server/main.go                          # Entry point, dependency injection
internal/
  domain/                                    # Domain layer (pure Go, no external deps)
    iam/model/                              # User, Role, Permission entities
    tenant/model/                           # Organization, App, APIKey entities
    apimanagement/model/                    # Endpoint, EndpointPolicy entities
    apimanagement/event/                    # Domain events
    apimanagement/repository/               # Repository interfaces
  application/                              # Application layer (use cases, DTOs)
    iam/usecase/                            # Register, Login, AssignRole
    tenant/usecase/                         # CreateOrg, CreateApp, ManageAPIKeys
    apimanagement/usecase/                  # DefineEndpoint, UpdatePolicy, RetireEndpoint, ActivateEndpoint
    apimanagement/dto/                      # Request/Response DTOs
  infrastructure/                           # Infrastructure layer (implementations)
    http/router/                            # Chi router setup
    http/handler/                           # HTTP handlers per domain
    postgres/                               # PostgreSQL repositories
    postgres/migrations/                    # Goose migrations
    redis/                                  # Redis client
    kafka/                                  # Kafka event bus
    gateway/interceptors/                   # Schema validation, PII masking
  gateway/                                  # API Gateway
    pipeline.go                             # Policy enforcement pipeline
    dynamic_handler.go                      # Dynamic CRUD handler
    backend_handler.go                      # Handler registry
    handlers/                               # Custom backend handlers
  shared/                                   # Shared utilities
    middleware/                             # Auth, RequestID, OrgID middleware
    errors/                                 # Domain error types
    events/                                 # Event bus interface
    pagination/                             # Pagination helpers
    response/                               # HTTP response helpers
    validator/                              # Request validation
```

## Key Conventions

### Naming
- Files: `snake_case.go`
- Packages: `lowercase` (single word preferred)
- Interfaces: Descriptive name (not `I` prefix) — `EndpointRepository`, not `IEndpointRepository`
- Constructors: `NewXxx()` functions
- Errors: `ErrXxx` variables or `domainErr.New()`

### Layers (Dependency Rule)
- **Domain** depends on nothing (pure Go)
- **Application** depends on Domain
- **Infrastructure** depends on Domain + Application
- **Never** import infrastructure from domain

### Error Handling
- Return `error` as last return value
- Use `domainErr.New(code, message, cause)` for domain errors
- Wrap errors with context: `fmt.Errorf("failed to X: %w", err)`
- Check errors immediately after function calls

### Multi-Tenancy
- Every business table has `organization_id` and `app_id`
- Extract from context: `middleware.OrgIDFromContext(ctx)`
- Set via headers: `X-Organization-ID`, `X-App-ID`

### Event-Driven
- Publish events after successful operations
- Use `events.EventBus` interface
- Topics: `masterfabric.iam`, `masterfabric.tenant`, `masterfabric.api-management`

### Dynamic Handler System
- Endpoints defined via API with `backend_service` and `backend_action`
- Actions: `list`, `get`, `create`, `update`, `patch`, `delete`
- Table derived from service name: `order-service` → `orders`
- Custom table via `schema.table_name`

## Build Commands

```bash
go build -o /tmp/masterfabric-server ./cmd/server
go test ./...
go vet ./...
goose -dir internal/infrastructure/postgres/migrations postgres "..." up
```

## Testing

- Unit tests: `*_test.go` next to source
- Use table-driven tests
- Mock interfaces, not implementations
- Test files use `package xxx_test` for black-box testing

## Common Patterns

### Adding a New Feature
1. Define model in `internal/domain/<context>/model/`
2. Define repository interface in `internal/domain/<context>/repository/`
3. Create use case in `internal/application/<context>/usecase/`
4. Create DTO in `internal/application/<context>/dto/`
5. Implement repository in `internal/infrastructure/postgres/<context>/`
6. Create HTTP handler in `internal/infrastructure/http/handler/<context>/`
7. Add route in `internal/infrastructure/http/router/router.go`
8. Wire dependencies in `cmd/server/main.go`

### Adding a Dynamic Endpoint
1. POST to `/api/v1/organizations/{orgId}/apps/{appId}/endpoints`
2. Set `backend_service` and `backend_action`
3. Gateway automatically handles CRUD operations
