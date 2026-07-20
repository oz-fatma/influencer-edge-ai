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

## Özellikler

- **Auth** — JWT tabanlı kimlik doğrulama (access + refresh token rotation)
- **Influencer Skorlama** — 0-100 arası, 4 alt kritere bölünmüş (genel, etkileşim, kitle, marka uyumu)
- **AI Eşleştirme Paneli** — WebLLM (Gemma 2B) ile tarayıcıda çalışan influencer analizi
- **LLM Monitoring** — çağrı latency'si, hata oranı, geçmiş çağrılar için canlı dashboard

## Stack

| Katman | Teknoloji |
|---|---|
| Backend | Go, Gin, GORM |
| Veritabanı | PostgreSQL, Redis |
| Frontend | Next.js, TypeScript, Tailwind CSS |
| AI | WebLLM (@mlc-ai/web-llm, Gemma 2B) |
| Deployment | Render (backend), Vercel (frontend) |

## Kurulum

### Backend
\`\`\`bash
cd backend
docker compose up -d
go run main.go
\`\`\`

### Frontend
\`\`\`bash
cd frontend
npm install
npm run dev
\`\`\`