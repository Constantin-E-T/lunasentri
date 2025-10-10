#!/usr/bin/env bash
set -euo pipefail

# Kill anything listening on port 80 (ignore if nothing is there)
if pids=$(lsof -ti:80); then
  kill -9 $pids
fi

cd /Users/emiliancon/Desktop/lunasentri/apps/api-go

# Export everything from the repo-level .env
if [ -f ../../.env ]; then
  set -a          # auto-export
  source ../../.env  # loads TELEGRAM_BOT_TOKEN, etc.
  set +a
fi

# Default local settings (overridden if present in .env)
export AUTH_JWT_SECRET="${AUTH_JWT_SECRET:-dev-secret}"
export SECURE_COOKIE="${SECURE_COOKIE:-false}"
export LOCAL_HOST_METRICS="${LOCAL_HOST_METRICS:-false}"
export CORS_ALLOWED_ORIGIN="${CORS_ALLOWED_ORIGIN:-http://localhost:3002}"

go run cmd/api/main.go
