# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.0.1   | :white_check_mark: |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

- **Email**: security@masterfabric.co
- **GitHub Security Advisory**: Use the [Security tab](https://github.com/masterfabric-go/masterfabric-go/security/advisories/new) to create a private security advisory

### What to Include

When reporting a security vulnerability, please include:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity (typically 30-90 days)

## Trust Model

masterfabric-go is a multi-tenant API management platform. The server is trusted to enforce authentication, authorization, tenant isolation, and policy rules. Clients are untrusted. Administrative operators are trusted to configure secrets, CORS origins, and infrastructure bindings correctly.

## Security Controls Registry v0.1

Baseline security controls implemented on the `security/hardening` branch (July 2026). Each control maps to a confirmed audit finding or a portable hardening pattern from the shared platform layer.

| ID | Category | Control | Implementation | CWE | Status |
| -- | -------- | ------- | -------------- | --- | ------ |
| SC-01 | Toolchain | Pin Go stdlib to a patched release | `go 1.26.4` in `go.mod` | CWE-94 | ✅ Implemented |
| SC-02 | Dependencies | Refresh vulnerable direct modules | pgx v5.10.0, chi v5.3.0, validator v10.30.3, `golang.org/x/crypto` v0.53.0 | CWE-400 | ✅ Implemented |
| SC-03 | Container | Non-root runtime with current base images | `golang:1.26.4-alpine` builder, `alpine:3.24` runtime, `appuser` | CWE-250 | ✅ Implemented |
| SC-04 | Infrastructure | Bind dev service ports to loopback | `DB_HOST_BIND`, `REDIS_HOST_BIND`, `KAFKA_HOST_BIND` default to `127.0.0.1` in `deployments/docker-compose.yml` | CWE-1392 | ✅ Implemented |
| SC-05 | Error handling | Sanitize internal server error responses | `internal/shared/response/json.go` — generic 5xx message, detail via `slog` | CWE-209 | ✅ Implemented |
| SC-06 | Configuration | Escape database DSN credentials | `DatabaseConfig.DSN()` uses `net/url.UserPassword` | CWE-116 | ✅ Implemented |
| SC-07 | Input validation | Clamp pagination page overflow | `MaxPage = 1_000_000` in `internal/shared/pagination/pagination.go` | CWE-190 | ✅ Implemented |
| SC-08 | Configuration | Bounded int32 environment parsing | `envOrDefaultInt32` for `DB_MAX_CONNS` / `DB_MIN_CONNS` | G115 | ✅ Implemented |
| SC-09 | HTTP surface | Safe CORS allow-list | `CORS_ALLOWED_ORIGINS` env + `middleware.CORSOptions`; credentials off for `*` or empty | CWE-942 | ✅ Implemented |
| SC-10 | HTTP surface | Global request body size cap | `MAX_BODY_BYTES` (default 1 MiB) via `middleware.MaxBodyBytes` | CWE-400 | ✅ Implemented |
| SC-11 | Observability | Generic readiness probe responses | `internal/infrastructure/http/handler/health/handler.go` — no raw error strings | CWE-209 | ✅ Implemented |
| SC-12 | Egress | Harden outbound HTTP proxy client | No redirect following, 30s timeout, 1 MiB response cap in `internal/gateway/dynamic_handler.go` | CWE-522 | ✅ Implemented |
| SC-13 | Authentication | Detect default JWT signing secret | Startup warning in `cmd/server/main.go` when `JWT_SECRET` is unchanged | CWE-798 | ✅ Implemented |
| SC-14 | Authorization | Enforce RBAC on administrative routes | `RequirePermission` on all `/api/v1` admin routes in `router.go` | CWE-306 | ✅ Implemented |
| SC-15 | Authorization | Wildcard-aware permission matching | `matchesPermission` in `internal/infrastructure/auth/rbac_service.go` (`*`, `org:*`, `*:read`) | CWE-285 | ✅ Implemented |
| SC-16 | Input validation | Sanitize migration script names | `scripts/migrate.sh create` — `[a-zA-Z0-9_]` charset only | CWE-22 | ✅ Implemented |
| SC-17 | Gateway | Suppress internal DB errors in dynamic handler | Generic `"an internal error occurred"` to clients; detail in logs | CWE-209 | ✅ Implemented |
| SC-18 | Gateway | Document intentional proxy SSRF sink | `#nosec G704` on admin-configured outbound proxy; accepted risk entry below | CWE-918 | ✅ Documented |
| SC-19 | Verification | Automated vulnerability scanning gate | `govulncheck` clean; `gosec` clean with 2 audited suppressions | — | ✅ Verified |

### Control summary

| Metric | Value |
| ------ | ----- |
| Registry version | **v0.1** |
| Total controls | **19** |
| Implemented | **18** |
| Documented accepted risk | **1** (SC-18) |
| Go toolchain | **1.26.4** |
| Verification | `go test ./...`, `govulncheck`, `gosec` |

### Environment variables

| Variable | Purpose | Production guidance |
| -------- | ------- | ------------------- |
| `JWT_SECRET` | HS256 signing key | Required; never use the default value |
| `CORS_ALLOWED_ORIGINS` | Comma-separated browser origins | Set explicit origins; avoid `*` |
| `MAX_BODY_BYTES` | Request body cap | Keep at or below gateway policy limits |
| `DB_SSLMODE` | PostgreSQL TLS mode | Use `require` or stricter |
| `DB_HOST_BIND` | Compose host bind for Postgres | Keep `127.0.0.1` outside isolated dev machines |

### Security Best Practices

When using masterfabric-go in production:

- Change default `JWT_SECRET` to a strong, random value
- Use SSL/TLS for database connections (`DB_SSLMODE=require`)
- Set `CORS_ALLOWED_ORIGINS` to explicit trusted origins
- Enable rate limiting for production workloads via endpoint policies
- Regularly update dependencies (`go get -u ./...`) and run `govulncheck`
- Review and rotate API keys regularly
- Monitor audit logs for suspicious activity
- Use environment variables for sensitive configuration
- Keep Docker images updated
- Restrict `/metrics` and health endpoints at the network edge

### Accepted Risks

| Risk | Rationale | Mitigation |
| ---- | --------- | ---------- |
| Unauthenticated `/metrics` and `/health/*` | Required for orchestrator probes and Prometheus scraping | Restrict by network policy or reverse-proxy auth |
| HS256 JWT | Simplicity for single-tenant deployments | Rotate secrets; prefer external identity for large fleets |
| Dynamic SQL gateway handler | Admin-defined table names via endpoint configuration | RBAC on endpoint creation; audit endpoint changes |
| Gateway HTTP proxy (gosec G704) | Managed endpoints may proxy to operator-configured backends | RBAC on endpoint creation; redirect refusal; response size cap |
| Development compose credentials | Convenience for local bootstrap | Loopback bind + documented dev-only posture |

### Verification Commands

```bash
go build ./... && go vet ./... && go test ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
go run github.com/securego/gosec/v2/cmd/gosec@latest -quiet ./...
```

### Security Updates

Security updates will be:

- Documented in CHANGELOG.md
- Tagged with security labels
- Released as patch versions

Thank you for helping keep masterfabric-go secure!
