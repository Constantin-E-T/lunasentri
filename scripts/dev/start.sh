#!/usr/bin/env bash
set -euo pipefail

echo "🚀 Starting LunaSentri development environment..."
echo ""

# Get the script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." && pwd )"

# Check if .env exists
if [ ! -f "$PROJECT_ROOT/.env" ]; then
  echo "⚠️  Warning: .env file not found at $PROJECT_ROOT/.env"
  echo "   Some features (like Telegram notifications) may be disabled."
  echo ""
fi

# Function to cleanup on exit
cleanup() {
  echo ""
  echo "🛑 Stopping all services..."
  # Kill all child processes
  jobs -p | xargs -r kill 2>/dev/null || true
  exit 0
}

trap cleanup SIGINT SIGTERM

# Start API server in background
echo "📡 Starting API server on port 80..."
"$SCRIPT_DIR/run-api.sh" &
API_PID=$!

# Wait a moment for API to start
sleep 2

# Start frontend in background
echo "🌐 Starting Next.js frontend on port 3002..."
"$SCRIPT_DIR/run-web.sh" &
WEB_PID=$!

echo ""
echo "✅ Development environment running!"
echo "   API:      http://localhost:80"
echo "   Frontend: http://localhost:3002"
echo ""
echo "Press Ctrl+C to stop all services"
echo ""

# Wait for any process to exit
wait -n

# Cleanup will be called by trap
