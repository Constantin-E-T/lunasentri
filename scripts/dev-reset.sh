#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
BACKEND_DIR="$ROOT_DIR/apps/api-go"
FRONTEND_DIR="$ROOT_DIR/apps/web-next"
DB_FILE="$BACKEND_DIR/data/lunasentri.db"
FRONTEND_PORT="${FRONTEND_PORT:-3000}"
API_URL="${NEXT_PUBLIC_API_URL:-http://localhost:8080}"
BACKEND_PORT="8080"

# Parse command line arguments
RESET_DB=false
for arg in "$@"; do
  case $arg in
    --reset-db)
      RESET_DB=true
      shift
      ;;
  esac
done

# Load .env file if it exists
if [ -f "$ROOT_DIR/.env" ]; then
  echo "Loading environment variables from .env..."
  export $(grep -v '^#' "$ROOT_DIR/.env" | xargs)
fi

command -v go >/dev/null 2>&1 || { echo "go is required"; exit 1; }
command -v pnpm >/dev/null 2>&1 || { echo "pnpm is required"; exit 1; }
command -v python3 >/dev/null 2>&1 || command -v openssl >/dev/null 2>&1 || { echo "python3 or openssl is required"; exit 1; }

# Kill existing processes by name
kill_existing() {
  pkill -f "go run main.go" >/dev/null 2>&1 || true
  pkill -f "next dev" >/dev/null 2>&1 || true
}

# Kill processes on specific ports
kill_port() {
  local port=$1
  echo "Checking for processes on port $port..."
  if lsof -ti:$port >/dev/null 2>&1; then
    echo "Killing processes on port $port..."
    lsof -ti:$port | xargs kill -9 2>/dev/null || true
    sleep 1
  fi
}

# Clean up existing processes and ports
kill_existing
kill_port $BACKEND_PORT
kill_port $FRONTEND_PORT

# Handle database reset
if [ "$RESET_DB" = true ]; then
  rm -f "$DB_FILE"
  echo "Database reset: $DB_FILE"
else
  if [ -f "$DB_FILE" ]; then
    echo "Database preserved: $DB_FILE"
  else
    echo "Database will be created: $DB_FILE"
  fi
fi

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
  
  # Ensure data directory exists
  mkdir -p data
  
  AUTH_JWT_SECRET="$AUTH_JWT_SECRET_VALUE" \
  SECURE_COOKIE=false \
  CORS_ALLOWED_ORIGIN="$CORS_ORIGIN" \
  TELEGRAM_BOT_TOKEN="${TELEGRAM_BOT_TOKEN:-}" \
  go run main.go 2>&1 | tee "$BACKEND_DIR/project/logs/backend.log"
) &
BACKEND_PID=$!

sleep 3

# Check if backend started successfully
if ! kill -0 $BACKEND_PID 2>/dev/null; then
  echo "❌ Backend failed to start. Check logs at: $BACKEND_DIR/project/logs/backend.log"
  exit 1
fi

# Verify backend is responding
if ! curl -s http://localhost:8080/health >/dev/null 2>&1; then
  echo "⚠️  Backend started but /health endpoint not responding yet..."
fi

echo "Starting frontend on $FRONTEND_PORT..."
(
  cd "$FRONTEND_DIR"
  pnpm install >/dev/null 2>&1
  pnpm dev --port "$FRONTEND_PORT" 2>&1 | tee "$FRONTEND_DIR/../api-go/project/logs/frontend.log"
) &
FRONTEND_PID=$!

sleep 2

# Check if frontend started successfully
if ! kill -0 $FRONTEND_PID 2>/dev/null; then
  echo "❌ Frontend failed to start. Check logs at: $BACKEND_DIR/project/logs/frontend.log"
  kill $BACKEND_PID 2>/dev/null || true
  exit 1
fi

echo "Backend PID: $BACKEND_PID"
echo "Frontend PID: $FRONTEND_PID"

if [ "$RESET_DB" = true ]; then
  echo "⚠️  Database was reset. First registered user will become admin."
else
  echo "Database preserved. Existing users remain available."
fi

echo "Visit http://localhost:$FRONTEND_PORT"
echo "Press Ctrl+C to stop both services."

cleanup() {
  echo "Stopping services..."
  kill "$BACKEND_PID" 2>/dev/null || true
  kill "$FRONTEND_PID" 2>/dev/null || true
}

trap cleanup EXIT INT TERM

wait
