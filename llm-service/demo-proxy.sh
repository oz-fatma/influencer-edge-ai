#!/usr/bin/env bash
# Quick checks for the Caddy + MLC reverse-proxy demo stack.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

echo "=== 1) MLC direct (:8000) ==="
if curl -sf -m 5 http://localhost:8000/v1/models >/dev/null 2>&1; then
  echo "OK — MLC responding on :8000"
  if curl -sI -m 5 http://localhost:8000/v1/models | grep -qi 'x-reverse-proxy'; then
    echo "WARN — unexpected X-Reverse-Proxy on direct MLC"
  else
    echo "OK — no Caddy header on direct MLC (expected)"
  fi
else
  echo "FAIL — start MLC: docker start influencer-llm-service"
  exit 1
fi

echo ""
echo "=== 2) Caddy reverse proxy (:8080) ==="
if ! docker ps --format '{{.Names}}' | grep -q '^influencer-llm-caddy$'; then
  echo "Starting Caddy..."
  docker compose -f docker-compose.proxy.yml up -d
  sleep 2
fi

HEADERS="$(curl -sI -m 5 http://localhost:8080/v1/models || true)"
if echo "$HEADERS" | grep -qi 'x-reverse-proxy: caddy'; then
  echo "OK — Caddy reverse proxy active (X-Reverse-Proxy: Caddy)"
else
  echo "FAIL — Caddy not proxying correctly"
  echo "$HEADERS"
  exit 1
fi

echo ""
echo "=== 3) Next: Cloudflare Tunnel ==="
echo "Run in another terminal:"
echo "  cloudflared tunnel --url http://127.0.0.1:8080"
echo ""
echo "Then set Render LLM_BASE_URL to the tunnel URL and redeploy mf-go."
