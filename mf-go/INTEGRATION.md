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
