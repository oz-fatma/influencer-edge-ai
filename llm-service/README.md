# LLM service (Ollama)

Local [Ollama](https://ollama.com/) stack for server-side Gemma 2 inference. Render mf-go reaches it via Cloudflare Tunnel → Caddy (see [PROXY-DEMO.md](./PROXY-DEMO.md)).

## Quick start

```bash
cd llm-service
docker compose up -d
docker exec -it ollama ollama pull gemma2:2b
```

Verify:

```bash
curl http://localhost:11434/v1/models
```

## Reverse proxy + tunnel (demo)

```bash
docker compose -f docker-compose.proxy.yml up -d
./demo-proxy.sh
cloudflared tunnel --url http://127.0.0.1:8080
```

Set Render `LLM_BASE_URL` to the tunnel URL and `LLM_MODEL=gemma2:2b`.

## Stop

```bash
docker compose down
docker compose -f docker-compose.proxy.yml down
```

Model data persists in the `ollama_data` Docker volume.
