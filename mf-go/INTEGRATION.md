# mf-go in InfluencerEdge AI

This folder contains [masterfabric-go](https://github.com/gurkanfikretgunak/masterfabric-go) integrated as a **parallel backend** for InfluencerEdge. The legacy Gin backend in `../backend/` is unchanged and keeps running on port **8080**.

## Ports

| Service | Legacy `backend/` | `mf-go/` |
|---------|-------------------|----------|
| API | `:8080` | `:8081` |
| Postgres | `:5433` | `:5434` |
| Redis | `:6379` | `:6380` |

## Quick start

```bash
cd mf-go
cp .env.example.integration .env
./dev.sh infra      # Postgres + Redis + Kafka (Kafka optional)
./dev.sh server     # hot-reload API on :8081
```

Health:

```bash
curl http://localhost:8081/health/ready
```

Auth + influencer flow:

```bash
# Register
curl -X POST http://localhost:8081/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"password123","first_name":"Demo","last_name":"User"}'

# Login
TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"password123"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")

# Create score
curl -X POST http://localhost:8081/api/v1/scores \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"influencer_name":"Ada Lovelace","platform":"instagram","overall_score":88}'
```

## Influencer domain endpoints

Mirrors legacy `backend/` API under `/api/v1`:

- `POST/GET /api/v1/scores`, `GET/PUT/DELETE /api/v1/scores/{id}`
- `POST/GET /api/v1/analyses`, `GET /api/v1/influencer-analysis/{id}`
- `POST /api/v1/llm-metrics`, `GET /api/v1/monitoring/stats`

## Frontend switch (optional)

Set a separate env when ready to try mf-go:

```env
NEXT_PUBLIC_MF_API_URL=http://localhost:8081
```

Legacy frontend continues using `NEXT_PUBLIC_API_URL=http://localhost:8080`.

## Tests

```bash
cd mf-go
go test ./...
```

## Render: shared Postgres with legacy backend

Use the **same Render Postgres database** as the classic Gin backend (`influencer_edge_db`). Do not create or switch to a separate `masterfabric` database.

### 1. Link Postgres to both services

In Render, attach the existing Postgres instance to:

- classic backend (`backend/`) — already uses `DATABASE_URL`
- mf-go service — link the **same** database

When linked, Render injects `DATABASE_URL` pointing at `influencer_edge_db`. mf-go reads `DATABASE_URL` automatically (same as the legacy backend).

Alternatively set explicit vars:

| Key | Value |
|---|---|
| `DB_HOST` | Render Postgres host |
| `DB_PORT` | `5432` |
| `DB_USER` | Render Postgres user |
| `DB_PASSWORD` | Render Postgres password |
| `DB_NAME` | `influencer_edge_db` |
| `DB_SSLMODE` | `require` |

Also set:

| Key | Value |
|---|---|
| `KAFKA_ENABLED` | `false` |
| `JWT_SECRET` | same secret as classic backend (or a new one if frontend only talks to mf-go) |

Remove `SERVER_PORT` if present — Render sets `PORT`.

### 2. Safe migration on shared DB

The legacy backend owns these tables in `public`:

- `users`, `refresh_tokens`, `influencer_scores`, `influencer_analyses`

mf-go expects different schemas for some of the same table names (UUID ids vs legacy `uint` ids). **Do not run full mf-go migrations in `public`** on `influencer_edge_db`.

**Recommended:** put mf-go tables in PostgreSQL schema **`mf`** and set on Render:

| Key | Value |
|---|---|
| `DB_SCHEMA` | `mf` |

#### One-time migration (from your Mac)

Render Dashboard → Postgres → **Connect** → copy **External Database URL**, then:

```bash
cd mf-go
export DATABASE_URL='postgres://USER:PASS@HOST/influencer_edge_db?sslmode=require'
export DB_SCHEMA=mf
chmod +x scripts/migrate_render_schema.sh
./scripts/migrate_render_schema.sh
```

Then on **influencer-edge-mfgo** Render service → Environment:

| Key | Value |
|---|---|
| `DB_SCHEMA` | `mf` |
| `CORS_ALLOWED_ORIGINS` | `https://influencer-edge-ai.vercel.app` |

Save → **Manual Deploy**.

Safe to run in `public` only (optional observability):

```bash
psql "$DATABASE_URL" -f mf-go/internal/infrastructure/postgres/migrations/00014_create_request_logs.sql
```

Or skip — `request_logs` is also created inside `mf` when you run the full schema migration.

### 3. What works today on shared DB

| Feature | Shared `influencer_edge_db` + `DB_SCHEMA=mf` |
|---|---|
| Classic backend | Works — unchanged in `public.*` |
| mf-go `/health/ready` | Works after DB env is set |
| mf-go auth (register/login) | Works after `migrate_render_schema.sh` + `DB_SCHEMA=mf` |
| mf-go influencer API | Works after schema migration |
| mf-go request logging / Grafana | Works (`mf.request_logs` or public migration) |

### 4. Next step for full mf-go on same DB

Pick one:

1. **PostgreSQL schema `mf`** — ✅ recommended (see above)
2. **Legacy-only data** — adapt mf-go to read/write legacy GORM table shapes
3. **Cutover** — migrate users/scores to mf-go schema, retire classic backend writes

Until `DB_SCHEMA=mf` is set and migrations are applied, keep classic backend on `influencer_edge_db` for auth; mf-go health checks work but register returns 500.

## Render: Redis (Key Value)

mf-go uses Redis for RBAC permission cache, rate limiting, and LLM metrics. Render provides this as **Key Value** (Redis-compatible).

### 1. Create Key Value instance

1. Render Dashboard → **New** → **Key Value**
2. **Region:** same as Postgres (e.g. Frankfurt)
3. **Plan:** Free or Starter
4. Create

### 2. Link to mf-go (and optionally classic backend)

1. Open mf-go web service → **Environment**
2. **Add from Render Dashboard** (or link Key Value resource)
3. Select your Key Value instance → property **`REDIS_URL`** (internal URL, e.g. `redis://red-xxxxx:6379`)
4. Save → redeploy

Classic backend also accepts `REDIS_URL` — link the **same** Key Value instance to both services if you want shared cache/metrics.

### 3. Verify

After deploy:

```bash
curl https://influencer-edge-mfgo.onrender.com/health/ready
```

Expected:

```json
{"status":"ready","services":{"postgres":"healthy","redis":"healthy"}}
```

### Manual env (without linking)

From Key Value → **Connect** → Internal URL:

| Key | Value |
|---|---|
| `REDIS_URL` | `redis://red-xxxxxxxx:6379` |

Or split:

| Key | Value |
|---|---|
| `REDIS_HOST` | `red-xxxxxxxx` |
| `REDIS_PORT` | `6379` |
