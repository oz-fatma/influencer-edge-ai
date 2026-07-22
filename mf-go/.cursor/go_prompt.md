# Go Clean Enterprise Backend Architecture Plan  
_For multi-tenant, RBAC-driven, app/endpoint-based SaaS platforms_

---

## 1. High-level goals and constraints

**Business and technical goals**

- **Target use case:** Enterprise-grade, multi-tenant SaaS platform where:
  - Organizations manage their own **Apps**.
  - Each App exposes configurable **Endpoints**.
  - Access is controlled via **RBAC**.
- **Key non-functional requirements:**
  - **Scalability:** Horizontal scaling of stateless services.
  - **Security:** Strong tenant isolation, RBAC, audit logging, PII handling.
  - **Reliability:** High availability, graceful degradation, retries, idempotency.
  - **Governance:** API lifecycle management, versioning, policy enforcement.
  - **Observability:** Logs, metrics, traces, audit trails.

**Architectural style**

- **Domain-driven design (DDD)** with **bounded contexts**.
- **Clean/hexagonal architecture** in Go:
  - Domain-centric, infrastructure as plugins.
- **Event-driven** where possible:
  - Domain events, async processing, eventual consistency where acceptable.

---

## 2. Core domain and bounded contexts

### 2.1 Bounded contexts

1. **Identity & Access Management (IAM) Context**
   - Organizations, users, roles, permissions, authentication, RBAC.
2. **Tenant & App Management Context**
   - Organizations, Apps, API keys, App-level roles and policies.
3. **API Management Context**
   - Endpoints, policies, routing, rate limiting, schema validation.
4. **Core Business Domain Context(s)**
   - Actual business features (e.g., Orders, Inventory, Projects, etc.).
5. **Billing & Subscription Context**
   - Plans, limits, usage tracking, billing integration.
6. **Observability & Audit Context**
   - Logs, metrics, traces, audit events, compliance.

Each context can be a separate service (microservice) or a module in a modular monolith, depending on maturity.

---

## 3. Multi-tenancy and RBAC model

### 3.1 Multi-tenancy model

- **Tenant = Organization**
  - Each Organization can have multiple Apps.
- **Isolation strategy:**
  - **Logical isolation** in PostgreSQL:
    - Every business table includes `organization_id` (tenant) and often `app_id`.
  - Future-ready for **hybrid isolation**:
    - Large tenants can be moved to dedicated DB instances.

**Tenant resolution**

- Tenant is resolved from:
  - Subdomain (`orgX.company.com`), or
  - JWT claims (`organization_id`), or
  - API key metadata.

### 3.2 RBAC model

**Entities**

- **User**
- **Organization**
- **App**
- **Role**
- **Permission**
- **UserRole** (user + org/app + role)
- **RolePermission** (role + permission)

**Levels**

- **Organization-level roles:**
  - `org.owner`, `org.admin`, `org.security`, `org.billing`, `org.audit`.
- **App-level roles:**
  - `app.owner`, `app.developer`, `app.viewer`, `app.integrator`.
- **Permissions:**
  - Organization-level: `org.manage_users`, `org.manage_apps`, `org.view_billing`.
  - App-level: `app.manage_endpoints`, `app.view_logs`, `app.publish_events`.
  - Endpoint-level: `endpoint.orders.create`, `endpoint.inventory.update`, etc.

**Token contents**

- `user_id`
- `organization_id`
- `app_id` (optional, depending on context)
- `roles`
- `permissions` (flattened, cached)

RBAC enforcement happens at the **API Gateway** and optionally re-checked in domain services for critical operations.

---

## 4. App and endpoint model (API governance)

### 4.1 Conceptual model

**Organization**

- Top-level tenant.
- Owns Apps, global policies, SSO configuration, data retention rules.

**App**

- Logical application within an Organization.
- Owns:
  - Endpoints
  - API keys / OAuth clients
  - App-level RBAC roles and permissions
  - Rate limit policies
  - SLA tier

**Endpoint**

- Managed API operation within an App.
- Not just `method + path`, but a **policy object**:
  - HTTP method, path
  - Required permission
  - Rate limit
  - Authentication policy
  - Schema validation
  - PII masking
  - Audit level
  - Backend service + action
  - Event publishing rules
  - Version (v1, v2, deprecated)

### 4.2 Endpoint lifecycle

- **Design:** Defined in the admin panel by App owners/developers.
- **Deploy:** Activated and propagated to the API Gateway configuration.
- **Monitor:** Metrics, logs, traces, error rates.
- **Version:** New versions added, old ones deprecated.
- **Retire:** Endpoint disabled and eventually removed.

---

## 5. System topology and communication

### 5.1 High-level components

- **API Gateway**
  - Entry point for all external traffic.
  - Handles auth, RBAC, rate limiting, routing, schema validation, logging.
- **IAM Service**
  - Authentication, user management, roles, permissions, SSO.
- **Tenant & App Service**
  - Organizations, Apps, API keys, App-level RBAC.
- **API Management Service**
  - Endpoint definitions, policies, gateway configuration sync.
- **Core Domain Services**
  - Business logic (Orders, Inventory, etc.).
- **Billing Service**
  - Plans, usage, invoices, payment provider integration.
- **Observability Service**
  - Centralized logs, metrics, traces, audit logs.

### 5.2 Communication patterns

- **North-south traffic (client → platform):**
  - Client → API Gateway → Backend services.
- **East-west traffic (service → service):**
  - gRPC/HTTP for synchronous calls.
  - Message queue/event bus (Kafka/NATS/RabbitMQ) for async events.

**Rule:** If an operation does not strictly require synchronous behavior, prefer async via events.

---

## 6. API Gateway and policy pipeline

### 6.1 Gateway responsibilities

For each request:

1. **Tenant resolution**
   - Determine `organization_id` (and optionally `app_id`).
2. **App resolution**
   - From path, host, or token.
3. **Endpoint lookup**
   - Match `method + path` to an endpoint definition.
4. **Authentication**
   - Validate JWT, API key, OAuth token, etc.
5. **Authorization (RBAC)**
   - Check if user’s permissions include endpoint’s required permission.
6. **Rate limiting**
   - Apply tenant/app/endpoint-level rate limits.
7. **Schema validation**
   - Validate request body against JSON schema.
8. **PII masking / data policies**
   - Mask sensitive fields in logs.
9. **Audit logging**
   - Record who did what, where, and when.
10. **Backend routing**
    - Forward to the appropriate service/action.
11. **Event publishing**
    - Publish domain or integration events if configured.
12. **Response shaping**
    - Optional: transform responses, add headers, etc.

### 6.2 Policy storage

Policies are stored in the **API Management** and **Tenant & App** contexts and periodically synced or dynamically fetched by the Gateway.

---

## 7. Data model (PostgreSQL, logical multi-tenancy)

### 7.1 Core tables (simplified)

**organization**

- `id`
- `name`
- `status`
- `sso_config`
- `data_retention_policy`
- `created_at`, `updated_at`

**apps**

- `id`
- `organization_id`
- `name`
- `status`
- `sla_tier`
- `rate_limit_policy`
- `created_at`, `updated_at`

**app_endpoints**

- `id`
- `app_id`
- `method`
- `path`
- `version`
- `backend_service`
- `backend_action`
- `schema` (JSON)
- `audit_level`
- `pii_masking`
- `event_after`
- `created_at`, `updated_at`

**app_endpoint_policies**

- `id`
- `endpoint_id`
- `required_permission`
- `rate_limit`
- `auth_policy`
- `validation_policy`
- `extra_policies` (JSON)

**users**

- `id`
- `email`
- `status`
- `password_hash` (if not external IdP)
- `created_at`, `updated_at`

**organization_users**

- `organization_id`
- `user_id`
- `status`
- `invited_by`
- `created_at`, `updated_at`

**app_users**

- `app_id`
- `user_id`
- `role_id`
- `created_at`, `updated_at`

**roles**

- `id`
- `scope_type` (`organization` or `app`)
- `scope_id` (organization_id or app_id)
- `name`
- `description`

**role_permissions**

- `role_id`
- `permission`

**app_api_keys**

- `id`
- `app_id`
- `key_hash`
- `scopes`
- `expires_at`
- `created_at`

**audit_logs**

- `id`
- `organization_id`
- `app_id`
- `endpoint_id`
- `user_id`
- `request_id`
- `timestamp`
- `action`
- `metadata` (JSON)

All business domain tables (e.g., `orders`, `projects`, `items`) include at least `organization_id` and often `app_id`.

---

## 8. Caching, queues, and events

### 8.1 Caching

Use **Redis** (or similar) for:

- **Permission cache**
  - `user:{id}:org:{id}:permissions`
  - `user:{id}:app:{id}:permissions`
- **Endpoint cache**
  - `app:{id}:endpoints` (for gateway lookup)
- **Rate limiting**
  - `rate:org:{id}:app:{id}:endpoint:{id}` counters
- **Session / token blacklists** (if needed)

### 8.2 Queues and events

Use a message broker (Kafka/NATS/RabbitMQ) for:

- **Domain events**
  - `UserInvited`, `AppCreated`, `EndpointUpdated`, `OrderCreated`, etc.
- **Integration events**
  - Webhook dispatch, email sending, external system sync.
- **Outbox pattern**
  - Ensure DB transaction + event publishing consistency.

Consumers must be **idempotent** and include `organization_id` and `app_id` in event payloads.

---

## 9. Observability and audit

### 9.1 Logging

- Structured JSON logs.
- Include:
  - `trace_id`, `span_id`
  - `organization_id`, `app_id`, `user_id`
  - `endpoint_id`
  - `request_id`
- Centralized log aggregation (e.g., Loki, ELK, etc.).

### 9.2 Metrics

- Prometheus-style metrics:
  - Request count, latency (p95/p99), error rate.
  - DB query latency.
  - Cache hit/miss.
  - Queue lag.
  - Per-tenant and per-app usage metrics.

### 9.3 Tracing

- OpenTelemetry-based tracing.
- Spans across:
  - Gateway
  - IAM
  - API Management
  - Core domain services
  - External integrations

### 9.4 Audit

- Separate **audit log** stream and storage.
- Immutable, append-only style.
- Used for compliance, security investigations, and governance.

---

## 10. Go project structure (clean architecture)

### 10.1 High-level layout

For each service (e.g., `iam-service`, `api-management-service`, `gateway`, `order-service`):

```text
/service-name
  /cmd
    /service-name
      main.go
  /internal
    /domain
      /model
      /service
      /event
    /application
      /usecase
      /dto
    /infrastructure
      /db
      /cache
      /queue
      /http
      /grpc
      /gateway (for gateway service)
    /interfaces
      /api (handlers)
      /cli
  /pkg
    (shared utilities if needed)
```

### 10.2 Domain layer

- Pure Go types and interfaces.
- No external dependencies (except maybe standard library).
- Contains:
  - Entities (Organization, App, Endpoint, Role, Permission, User, etc.).
  - Domain services (e.g., `RBACService`, `TenantService`).
  - Domain events.

### 10.3 Application layer

- Use cases / application services.
- Orchestrates domain logic and infrastructure.
- Example:
  - `CreateApp`
  - `DefineEndpoint`
  - `AssignRoleToUser`
  - `CheckPermission`

### 10.4 Infrastructure layer

- Implementations of:
  - Repositories (Postgres via `pgx`).
  - Cache clients (Redis).
  - Queue producers/consumers.
  - HTTP/gRPC servers and clients.
  - Config, logging, metrics.

### 10.5 Interfaces layer

- HTTP handlers, gRPC servers, CLI commands.
- Maps external requests to application use cases and DTOs.

---

## 11. Security and compliance considerations

- **Authentication:**
  - JWT, OAuth2/OIDC, API keys.
  - Optional SSO via enterprise IdPs (SAML/OIDC).
- **Authorization:**
  - Centralized RBAC via IAM + Gateway.
  - Resource-level checks in domain services for critical operations.
- **Data protection:**
  - Encryption in transit (TLS everywhere).
  - Encryption at rest (DB, backups).
  - PII masking in logs.
- **Compliance:**
  - Data retention policies per organization.
  - Right-to-be-forgotten workflows.
  - Audit trails for admin actions.

---

## 12. Evolution path

### Phase 1: Modular monolith

- Single Go binary with clear bounded contexts as modules.
- Single Postgres instance (logical multi-tenancy).
- Redis + simple queue.

### Phase 2: Service extraction

- Extract IAM, API Management, and core domain services into separate deployable services.
- Introduce dedicated message broker.
- Introduce read replicas for Postgres.

### Phase 3: Enterprise scale

- Per-tenant or per-SLA-tier DB clusters.
- Advanced API Gateway features (WAF, advanced threat protection).
- Fine-grained policy engine (e.g., attribute-based access control on top of RBAC).
- Region-based deployments and data residency.

---

This single-file plan gives you a complete conceptual and structural blueprint for a **Go-based, clean, enterprise-grade, multi-tenant, RBAC-driven backend architecture** suitable for companies that need strong governance, scalability, and security.