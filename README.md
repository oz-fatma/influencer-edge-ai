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

AI destekli influencer-ajans eşleştirme platformu. Ajansların influencer adaylarını 
filtreleyip kampanyaya uygunluklarını otomatik olarak puanlamasını sağlar.

MasterFabric Academy Agentic AI Developer Training — Cohort 1, OOP fazı projesi.

## 🔗 Canlı Demo

- **Frontend:** https://influencer-edge-ai.vercel.app
- **Backend API:** https://influencer-edge-ai.onrender.com

## Özellikler

- **Auth** — JWT tabanlı kimlik doğrulama (access + refresh token rotation)
- **Influencer Havuzu** — influencer ekleme, listeleme, skorlama
- **Influencer Skorlama** — 0-100 arası, 4 alt kritere bölünmüş (genel, etkileşim, kitle, marka uyumu)
- **AI Eşleştirme Paneli** — WebLLM (Gemma 2B) ile tarayıcıda çalışan influencer analizi
- **LLM Monitoring** — çağrı latency'si, hata oranı, geçmiş çağrılar için canlı dashboard

## Influencer Ekleme

"+ Influencer Ekle" formuyla havuza yeni bir influencer eklenirken şu bilgiler girilir:

| Alan | Açıklama |
|---|---|
| `influencer_name` | Influencer'ın adı (zorunlu) |
| `platform` | instagram, tiktok, youtube, twitter, linkedin, other |
| `notes` | Kısa açıklama/niş bilgisi (opsiyonel) |
| `overall_score`, `engagement_score`, `audience_score`, `brand_fit_score` | 0-100 arası manuel skorlar (opsiyonel, boş bırakılırsa 0 kaydedilir) |

Skorlar manuel girilebilir veya Eşleştirme Paneli'nde WebLLM analiziyle otomatik üretilebilir.

## Stack

| Katman | Teknoloji |
|---|---|
| Backend | Go, Gin, GORM |
| Veritabanı | PostgreSQL, Redis |
| Frontend | Next.js, TypeScript, Tailwind CSS |
| AI | WebLLM (@mlc-ai/web-llm, Gemma 2B — tarayıcıda çalışır) |
| Deployment | Render (backend), Vercel (frontend) |

## API Endpoint'leri (22)

| Grup | Endpoint'ler |
|---|---|
| Common | `GET /health`, `GET /version` |
| Config | `GET /config`, `GET /health-config` |
| Auth | `POST /auth/register`, `/login`, `/logout`, `/refresh-token`, `GET/PUT /auth/profile`, `PUT /auth/change-password`, `DELETE /auth/account` |
| Scores | `POST/GET /api/scores`, `GET/PUT/DELETE /api/scores/:id` |
| Analyses | `POST/GET /api/analyses`, `GET /api/influencer-analysis/:id` |
| Monitoring | `POST /api/llm-metrics`, `GET /api/monitoring/stats` |

## Kurulum (Yerel Geliştirme)

### Backend
\`\`\`bash
cd backend
docker compose up -d   # Postgres + Redis
go run main.go
\`\`\`

### Frontend
\`\`\`bash
cd frontend
npm install
npm run dev
\`\`\`

Frontend, backend'e `http://localhost:8080` üzerinden bağlanır (`NEXT_PUBLIC_API_URL` ile değiştirilebilir).

## Mimari Notlar

- Şifreler **bcrypt** ile hash'lenir, JWT refresh token'lar rotation ile yenilenir
- Her API isteğinde kullanıcı verisi izolasyonu sağlanır (IDOR koruması)
- LLM analizi tamamen **tarayıcı tarafında** (client-side) çalışır — sunucuya influencer verisi dışında hassas veri gönderilmez
- Monitoring verileri Redis sorted set üzerinde tutulur (zaman damgalı, hızlı erişim)