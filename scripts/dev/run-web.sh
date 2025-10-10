#!/usr/bin/env bash
set -euo pipefail

# Kill anything listening on port 3002 (ignore if nothing is there)
if pids=$(lsof -ti:3002); then
  kill -9 $pids
fi

cd /Users/emiliancon/Desktop/lunasentri/apps/web-next

# Export everything from the repo-level .env
if [ -f ../../.env ]; then
  set -a          # auto-export
  source ../../.env
  set +a
fi

npm run dev -- --port 3002
