<div align="center">
<a href="https://academy.masterfabric.co">
  <img src="https://academy.masterfabric.co/academy-badge.png" width="120" alt="MasterFabric Academy">
</a>
<p>
  <sub>
    academy.masterfabric.co is a
    <a href="https://masterfabric.co">MasterFabric</a>
    subsidiary.
  </sub>
</p>
</div>

# InfluencerEdge AI

**InfluencerEdge AI** is an influencer–brand matching platform. Agencies manage influencer pools, define brand profiles, and run AI-powered fit analysis using a **PEFT/LoRA fine-tuned Gemma 2B** model served via **Hugging Face Inference Endpoints**.

## Live Demo

| Service | URL |
|---------|-----|
| Frontend (Vercel) | https://influencer-edge-ai.vercel.app |
| Backend API (Render) | https://influencer-edge-mfgo.onrender.com |

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Client (Browser)                                │
└───────────────────────────────────┬─────────────────────────────────────────┘
                                    │ HTTPS
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                    Vercel — Next.js Frontend (TypeScript)                    │
│   Dashboard · Influencer Pool · Brand Profiles · Matching Panel · Admin      │
└───────────────────────────────────┬─────────────────────────────────────────┘
                                    │ REST /api/v1 + JWT
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│              Render — Go Backend (mf-go, Chi router)                         │
│   Auth · Scores · Analyses · Brand Profiles · MCP · Admin LLM config       │
└───────┬─────────────────┬─────────────────────┬───────────────────────────┘
        │                 │                     │
        │ SQL             │ cache / metrics     │ events (optional)
        ▼                 ▼                     ▼
┌───────────────┐  ┌──────────────┐      ┌──────────────┐
│  PostgreSQL   │  │    Redis     │      │    Kafka     │
│   (Render)    │  │   (Render)   │      │  (local/dev) │
└───────────────┘  └──────────────┘      └──────────────┘
        │
        │ LLM analyze (server-side)
        ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│        Hugging Face Inference Endpoint — fine-tuned Gemma 2B (PEFT/LoRA)   │
│        LLM_ENDPOINT_TYPE=hf-inference · OpenAI-compatible /v1/chat/completions│
└─────────────────────────────────────────────────────────────────────────────┘

Optional local path: llm-service/ (Ollama + Caddy proxy) for development.
Browser fallback: WebLLM (@mlc-ai/web-llm) runs Gemma 2B client-side in Matching Panel.
```

## Features

- **Influencer profile management** — create, list, update, and delete influencers with platform metadata and manual scores (overall, engagement, audience, brand fit).
- **Brand Profile** — define brand context (name, industry, target audience, values, campaign goals) and inject it into MCP/LLM analysis prompts.
- **MCP (Model Context Protocol) integration** — `POST /api/v1/mcp/request` with `analyze_influencer` and optional `brand_profile_id`.
- **Admin panel** — manage system prompt, model selection, and max tokens (`/admin` UI, admin-only API).
- **Monitoring** — LLM call latency, error rate, and recent logs (`/monitoring`, admin-only).
- **Analysis History** — view past analyses per influencer in the Matching Panel (`GET /api/v1/analyses`).
- **Automatic database migrations** — embedded SQL migrations run on server startup (`schema_migrations` tracking).

## Tech Stack

| Layer | Technology |
|-------|------------|
| Frontend | Next.js 16, React 19, TypeScript, Tailwind CSS |
| Backend | Go 1.26, Chi router, pgx, clean/hexagonal architecture |
| Legacy backend | Go, Gin, GORM (`backend/` — reference / parallel dev) |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Message queue | Apache Kafka (local dev via Docker; optional in production) |
| LLM | Gemma 2B, PEFT/LoRA fine-tuning, Hugging Face Inference Endpoints |
| Client LLM fallback | WebLLM (`@mlc-ai/web-llm`) |
| Deployment | Vercel (frontend), Render (mf-go API + Postgres + Redis) |

## Local Setup

### Prerequisites

- **Go** 1.26+ (mf-go)
- **Node.js** 20+ and npm (frontend)
- **Docker** and Docker Compose (Postgres, Redis, Kafka for local infra)

### 1. Backend (mf-go)

```bash
cd mf-go
cp .env.example.integration .env   # adjust ports and secrets as needed
./dev.sh                           # infra + hot-reload server (migrations on startup)
```

`./dev.sh` starts Docker services (Postgres `:5434`, Redis `:6380`, Kafka), waits for health checks, and runs the API with **air** hot-reload. Database migrations apply automatically when the server starts.

Other commands:

```bash
./dev.sh infra    # Docker only
./dev.sh server   # API only (infra must be running)
./dev.sh down     # stop Docker services
```

Health check:

```bash
curl http://localhost:8081/health/ready
```

### 2. Frontend

```bash
cd frontend
npm install
npm run dev
```

Open http://localhost:3000. The app calls the API at `http://localhost:8081` by default (override with `NEXT_PUBLIC_API_URL`).

### 3. Environment variables

**mf-go** (`.env` — copy from `.env.example.integration`):

```bash
# Server
SERVER_HOST=
SERVER_PORT=
PORT=                          # Render uses PORT; local dev often 8081
CORS_ALLOWED_ORIGINS=

# Database (DATABASE_URL or individual vars)
DATABASE_URL=
DB_HOST=
DB_PORT=
DB_USER=
DB_PASSWORD=
DB_NAME=
DB_SSLMODE=
DB_SCHEMA=                     # e.g. mf on shared Render DB

# Redis (REDIS_URL or individual vars)
REDIS_URL=
REDIS_HOST=
REDIS_PORT=
REDIS_PASSWORD=
REDIS_DB=

# Auth
JWT_SECRET=
JWT_EXPIRATION_HOURS=
JWT_ISSUER=

# Kafka
KAFKA_ENABLED=
KAFKA_BROKERS=
KAFKA_GROUP_ID=

# LLM / MCP
LLM_BASE_URL=                  # HF endpoint or Ollama proxy URL
LLM_ENDPOINT_TYPE=             # chat | hf-inference
LLM_API_KEY=
LLM_MODEL=
LLM_TIMEOUT_SECONDS=
MCP_MODEL=

# Logging
LOG_LEVEL=
LOG_FORMAT=
```

**frontend**:

```bash
NEXT_PUBLIC_API_URL=           # default http://localhost:8081
```

For Hugging Face Inference in production, set `LLM_ENDPOINT_TYPE=hf-inference`, `LLM_BASE_URL`, `LLM_API_KEY`, and `LLM_MODEL` on Render.

## Project Structure

```
influencer-edge-ai/
├── frontend/              # Next.js app (Vercel)
│   ├── app/                 # Routes: dashboard, influencers, brand-profiles, matching, admin, monitoring
│   ├── components/
│   └── lib/                 # API client, MCP helpers, auth
├── mf-go/                   # Primary Go backend (Render)
│   ├── cmd/                 # Application entrypoint
│   ├── internal/            # Domain, application, infrastructure (HTTP, postgres, llm, mcp)
│   ├── deployments/         # docker-compose.yml
│   ├── dev.sh               # Local dev runner
│   └── .env.example.integration
├── backend/               # Legacy Gin backend (reference)
├── llm-service/             # Local Ollama + Caddy proxy for LLM dev
├── influencer_training_data.json
└── README.md
```

## API Overview (InfluencerEdge)

All protected routes require `Authorization: Bearer <token>`.

| Group | Endpoints |
|-------|-----------|
| Health | `GET /health/live`, `GET /health/ready`, `GET /metrics` |
| Auth | `POST /api/v1/auth/register`, `POST /api/v1/auth/login`, `GET /api/v1/auth/me` |
| Scores | `POST/GET /api/v1/scores`, `GET/PUT/DELETE /api/v1/scores/{id}` |
| Analyses | `POST/GET /api/v1/analyses`, `GET /api/v1/influencer-analysis/{id}` |
| Brand Profiles | `POST/GET /api/v1/brand-profiles`, `GET/PUT/DELETE /api/v1/brand-profiles/{id}` |
| LLM | `POST /api/v1/llm/analyze`, `POST /api/v1/mcp/request`, `POST /api/v1/llm-metrics` |
| Admin | `GET/PUT /api/v1/admin/llm-config`, `GET /api/v1/admin/llm-logs`, `GET /api/v1/admin/llm-models` |
| Monitoring | `GET /api/v1/monitoring/stats` (admin-only) |

## Developer

Developed by **Fatma Öz** as part of the **MasterFabric Academy Agentic AI Developer Training** program.
