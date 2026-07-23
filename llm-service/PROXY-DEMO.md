# Reverse proxy demo (Caddy + Tunnel + Render mf-go)

Sunumda **üç katmanlı reverse proxy** zincirini göstermek için bu rehberi kullan.

## Mimari (hocanın modeli + somut Caddy katmanı)

```text
[Vercel]  Matching → Analyze
    ↓
[Render]  mf-go  POST /api/v1/llm/analyze     ← ① API reverse proxy
    ↓
[Internet]  https://xxx.trycloudflare.com     ← ② Cloudflare Tunnel (dış kapı)
    ↓
[Mac]  Caddy :8080                            ← ③ Caddy reverse proxy (somut)
    ↓
[Mac]  MLC-LLM Docker :8000                   ← model burada
```

| Katman | Rol | Sunumda ne dersin? |
|---|---|---|
| **mf-go (Render)** | Frontend isteğini LLM'e iletir | "API gateway / reverse proxy" |
| **Cloudflare Tunnel** | Mac'i internete bağlar | "Private LLM'i public HTTPS'e çıkarma" |
| **Caddy** | MLC önünde yönlendirme | "Docker üzerinde reverse proxy (Caddy)" |
| **MLC-LLM** | Inference | "Ağır model kendi makinemde" |

---

## Kurulum (3 terminal)

### Terminal 1 — MLC-LLM

```bash
docker start influencer-llm-service
curl http://localhost:8000/v1/models
```

Beklenen: JSON model listesi (`/app/model`).

### Terminal 2 — Caddy reverse proxy

```bash
cd llm-service
docker compose -f docker-compose.proxy.yml up -d
curl -I http://localhost:8080/v1/models
```

Beklenen: `HTTP/1.1 200` ve header:

```text
X-Reverse-Proxy: Caddy
X-Llm-Upstream: host.docker.internal:8000
```

**Doğrudan MLC'ye** gidersen bu header **yok**:

```bash
curl -I http://localhost:8000/v1/models
# X-Reverse-Proxy görünmez → Caddy katmanını kanıtlar
```

### Terminal 3 — Cloudflare Tunnel (Caddy'ye, MLC'ye değil)

```bash
cloudflared tunnel --url http://127.0.0.1:8080
```

URL'yi kopyala (ör. `https://abc123.trycloudflare.com`).

```bash
curl -I https://abc123.trycloudflare.com/v1/models
```

Beklenen: yine `X-Reverse-Proxy: Caddy` (tunnel → Caddy → MLC).

---

## Render bağlantısı

mf-go servisinde:

| Key | Value |
|---|---|
| `LLM_BASE_URL` | Tunnel URL (Caddy üzerinden) |
| `LLM_TIMEOUT_SECONDS` | `300` |
| `SERVER_WRITE_TIMEOUT_SECONDS` | `600` |

Redeploy sonrası mf-go loglarında: `LLM proxy enabled`.

---

## Sunumda kanıt adımları (sırayla göster)

### 1) Caddy katmanı (lokal)

```bash
# MLC direkt — proxy header yok
curl -I http://localhost:8000/v1/models

# Caddy üzerinden — proxy header var
curl -I http://localhost:8080/v1/models | grep -i x-reverse
```

### 2) Tunnel katmanı

```bash
curl -I https://YOUR-TUNNEL.trycloudflare.com/v1/models | grep -i x-reverse
```

Aynı header internetten geliyorsa: **Tunnel → Caddy → MLC** çalışıyor.

### 3) Render mf-go katmanı

```bash
curl -X POST https://influencer-edge-mfgo.onrender.com/api/v1/llm/analyze \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"influencer_name":"Demo","platform":"instagram","notes":"Beauty niche"}'
```

JSON skor dönerse: **Vercel → Render → Tunnel → Caddy → MLC** tam zincir.

### 4) Vercel UI

1. https://influencer-edge-ai.vercel.app → Matching
2. Influencer seç → **Analyze**
3. "Server-side MLC-LLM" + skorlar

---

## Sunum cümlesi (kopyala-yapıştır)

> "LLM inference kendi makinemde Docker'da çalışıyor. Caddy reverse proxy olarak MLC'nin önünde duruyor. Cloudflare Tunnel ile private servisi HTTPS'e açıyorum. Render'daki mf-go API reverse proxy olarak frontend'den gelen Analyze isteğini bu LLM servisine iletiyor. Böylece microservice ayrımı var: API Render'da, model bende, trafik Render üzerinden akıyor."

---

## Durdurma

```bash
# Tunnel: Ctrl+C
docker compose -f llm-service/docker-compose.proxy.yml down
# MLC'yi durdurmak isteğe bağlı:
# docker stop influencer-llm-service
```

---

## Sorun giderme

| Belirti | Muhtemel neden |
|---|---|
| `:8080` connection refused | Caddy container kapalı → `docker compose ... up -d` |
| `:8080` 502 | MLC kapalı → `docker start influencer-llm-service` |
| Tunnel 502 | Tunnel Caddy'ye değil `:8000`'e bağlı → `--url http://127.0.0.1:8080` kullan |
| Render 503 | `LLM_BASE_URL` boş veya tunnel kapalı |
| Analyze çok yavaş | CPU + amd64 emülasyonu; ilk istek 1–3 dk sürebilir |
