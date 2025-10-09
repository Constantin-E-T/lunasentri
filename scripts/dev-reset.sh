#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
BACKEND_DIR="$ROOT_DIR/apps/api-go"
FRONTEND_DIR="$ROOT_DIR/apps/web-next"
DB_FILE="$BACKEND_DIR/data/lunasentri.db"
FRONTEND_PORT="${FRONTEND_PORT:-3000}"
API_URL="${NEXT_PUBLIC_API_URL:-http://localhost:8080}"

# Load .env file if it exists
if [ -f "$ROOT_DIR/.env" ]; then
  echo "Loading environment variables from .env..."
  export $(grep -v '^#' "$ROOT_DIR/.env" | xargs)
fi

command -v go >/dev/null 2>&1 || { echo "go is required"; exit 1; }
command -v pnpm >/dev/null 2>&1 || { echo "pnpm is required"; exit 1; }
command -v python3 >/dev/null 2>&1 || command -v openssl >/dev/null 2>&1 || { echo "python3 or openssl is required"; exit 1; }

kill_existing() {
  pkill -f "go run main.go" >/dev/null 2>&1 || true
  pkill -f "next dev" >/dev/null 2>&1 || true
}

kill_existing

rm -f "$DB_FILE"
echo "Database reset: $DB_FILE"

generate_secret() {
  if command -v python3 >/dev/null 2>&1; then
    python3 -c 'import secrets, sys; sys.stdout.write(secrets.token_urlsafe(32))'
  else
    openssl rand -base64 32 | tr -d '\n'
  fi
}

AUTH_JWT_SECRET_VALUE="${AUTH_JWT_SECRET:-$(generate_secret)}"
CORS_ORIGIN="http://localhost:$FRONTEND_PORT"

export NEXT_PUBLIC_API_URL="$API_URL"

echo "Starting backend on 8080..."
(
  cd "$BACKEND_DIR"
  AUTH_JWT_SECRET="$AUTH_JWT_SECRET_VALUE" \
  SECURE_COOKIE=false \
  CORS_ALLOWED_ORIGIN="$CORS_ORIGIN" \
  TELEGRAM_BOT_TOKEN="${TELEGRAM_BOT_TOKEN:-}" \
  go run main.go
) &
BACKEND_PID=$!

sleep 2

echo "Starting frontend on $FRONTEND_PORT..."
(
  cd "$FRONTEND_DIR"
  pnpm install >/dev/null
  pnpm dev --port "$FRONTEND_PORT"
) &
FRONTEND_PID=$!

echo "Backend PID: $BACKEND_PID"
echo "Frontend PID: $FRONTEND_PID"

echo "First registered user will become admin. Visit http://localhost:$FRONTEND_PORT"

echo "Press Ctrl+C to stop both services."

cleanup() {
  echo "Stopping services..."
  kill "$BACKEND_PID" 2>/dev/null || true
  kill "$FRONTEND_PID" 2>/dev/null || true
}

trap cleanup EXIT INT TERM

wait
