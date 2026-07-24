#!/usr/bin/env bash
# Sends repeated requests through the scale-demo nginx LB and verifies traffic
# hits more than one mf-go app replica (via X-Instance-ID).
# Compatible with macOS default bash 3.2 (no associative arrays).
set -euo pipefail

BASE_URL="${LB_URL:-http://127.0.0.1:8082}"
REQUESTS="${REQUESTS:-15}"
MIN_UNIQUE="${MIN_UNIQUE:-2}"

IDS_FILE="$(mktemp)"
trap 'rm -f "$IDS_FILE"' EXIT

echo "Load balancer URL: $BASE_URL"
echo "Sending $REQUESTS requests to /health/live ..."

prev_unique=0

for i in $(seq 1 "$REQUESTS"); do
  headers="$(mktemp)"
  body="$(mktemp)"
  if ! curl -sf -D "$headers" -o "$body" "$BASE_URL/health/live"; then
    echo "FAIL — request $i failed (is nginx + scaled app running?)"
    rm -f "$headers" "$body"
    exit 1
  fi

  instance_id=""
  if grep -qi '^x-instance-id:' "$headers"; then
    instance_id="$(grep -i '^x-instance-id:' "$headers" | tail -1 | cut -d' ' -f2- | tr -d '\r')"
  fi
  if [[ -z "$instance_id" ]] && command -v python3 >/dev/null 2>&1; then
    instance_id="$(python3 -c "import json; print(json.load(open('$body')).get('instance_id',''))" 2>/dev/null || true)"
  fi

  rm -f "$headers" "$body"

  if [[ -z "$instance_id" ]]; then
    echo "FAIL — no instance_id in response $i"
    exit 1
  fi

  printf '%s\n' "$instance_id" >>"$IDS_FILE"
  unique="$(sort -u "$IDS_FILE" | wc -l | tr -d ' ')"
  if [[ "$unique" -gt "$prev_unique" ]]; then
    echo "  hit instance: $instance_id (unique so far: $unique)"
    prev_unique="$unique"
  fi
done

echo ""
if [[ "$unique" -ge "$MIN_UNIQUE" ]]; then
  echo "OK — $unique distinct instance(s) across $REQUESTS requests (load balancing works)"
  exit 0
fi

echo "FAIL — only $unique distinct instance(s); expected at least $MIN_UNIQUE"
echo "Tip: docker compose --profile scale-demo ps  (need app scaled to 3)"
exit 1
