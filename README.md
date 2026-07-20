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

AI-powered influencer–agency matching platform. Helps agencies filter influencer candidates and automatically score their campaign fit.

MasterFabric Academy Agentic AI Developer Training — Cohort 1, OOP phase project.

## 🔗 Live Demo

- **Frontend:** https://influencer-edge-ai.vercel.app
- **Backend API:** https://influencer-edge-ai.onrender.com

## Features

- **Auth** — JWT-based authentication (access + refresh token rotation)
- **Influencer Pool** — add, list, and score influencers
- **Influencer Scoring** — 0–100 scale split across 4 sub-criteria (overall, engagement, audience, brand fit)
- **AI Matching Panel** — browser-based influencer analysis with WebLLM (Gemma 2B)
- **LLM Monitoring** — live dashboard for call latency, error rate, and call history

## Adding Influencers

When adding a new influencer to the pool via the **"+ Add Influencer"** form, the following fields are used:

| Field | Description |
|---|---|
| `influencer_name` | Influencer name (required) |
| `platform` | instagram, tiktok, youtube, twitter, linkedin, other |
| `notes` | Short description/niche info (optional) |
| `overall_score`, `engagement_score`, `audience_score`, `brand_fit_score` | Manual scores from 0–100 (optional; defaults to 0 if left blank) |

Scores can be entered manually or generated automatically via WebLLM analysis in the Matching Panel.

## Stack

| Layer | Technology |
|---|---|
| Backend | Go, Gin, GORM |
| Database | PostgreSQL, Redis |
| Frontend | Next.js, TypeScript, Tailwind CSS |
| AI | WebLLM (@mlc-ai/web-llm, Gemma 2B — runs in the browser) |
| Deployment | Render (backend), Vercel (frontend) |

## API Endpoints (22)

| Group | Endpoints |
|---|---|
| Common | `GET /health`, `GET /version` |
| Config | `GET /config`, `GET /health-config` |
| Auth | `POST /auth/register`, `/login`, `/logout`, `/refresh-token`, `GET/PUT /auth/profile`, `PUT /auth/change-password`, `DELETE /auth/account` |
| Scores | `POST/GET /api/scores`, `GET/PUT/DELETE /api/scores/:id` |
| Analyses | `POST/GET /api/analyses`, `GET /api/influencer-analysis/:id` |
| Monitoring | `POST /api/llm-metrics`, `GET /api/monitoring/stats` |

## Setup (Local Development)

### Backend
```bash
cd backend
docker compose up -d   # Postgres + Redis
go run main.go
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

The frontend connects to the backend at `http://localhost:8080` (configurable via `NEXT_PUBLIC_API_URL`).

## Architecture Notes

- Passwords are hashed with **bcrypt**; JWT refresh tokens are rotated on renewal
- User data isolation on every API request (IDOR protection)
- LLM analysis runs entirely **client-side** — no sensitive data is sent to the server beyond influencer metadata
- Monitoring data is stored in Redis sorted sets (timestamped, fast access)
