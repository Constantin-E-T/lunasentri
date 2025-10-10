# Development Scripts Implementation Summary

## Overview

Created a set of shell scripts to simplify local development by properly loading environment variables (including `TELEGRAM_BOT_TOKEN`) from the repository root `.env` file.

## Problem Solved

The previous one-liner command for starting the API server didn't read the `.env` file, so `TELEGRAM_BOT_TOKEN` remained unset and Telegram notifications were disabled.

**Before:**

```bash
# TELEGRAM_BOT_TOKEN was never loaded
lsof -ti:80 | xargs kill -9 2>/dev/null; sleep 1; \
  (cd /Users/emiliancon/Desktop/lunasentri/apps/api-go && \
  AUTH_JWT_SECRET=dev-secret SECURE_COOKIE=false \
  LOCAL_HOST_METRICS=false CORS_ALLOWED_ORIGIN=http://localhost:3002 \
  go run cmd/api/main.go)
```

**After:**

```bash
# Properly loads .env and all environment variables
./scripts/dev/start.sh
```

## Files Created

### 1. `scripts/dev/run-api.sh`

**Purpose:** Start the Go API server with proper environment loading

**Features:**

- ✅ Kills existing processes on port 80
- ✅ Changes to `apps/api-go` directory
- ✅ Loads all environment variables from `../../.env` (repository root)
- ✅ Sets sensible defaults for development
- ✅ Uses `set -euo pipefail` for safer script execution

**Usage:**

```bash
./scripts/dev/run-api.sh
```

**Environment Variables:**

- Loaded from `.env`: `TELEGRAM_BOT_TOKEN`, etc.
- Default values (can be overridden):
  - `AUTH_JWT_SECRET=dev-secret`
  - `SECURE_COOKIE=false`
  - `LOCAL_HOST_METRICS=false`
  - `CORS_ALLOWED_ORIGIN=http://localhost:3002`

### 2. `scripts/dev/run-web.sh`

**Purpose:** Start the Next.js frontend

**Features:**

- ✅ Kills existing processes on port 3002
- ✅ Changes to `apps/web-next` directory
- ✅ Loads environment variables from `../../.env`
- ✅ Runs `npm run dev -- --port 3002`

**Usage:**

```bash
./scripts/dev/run-web.sh
```

### 3. `scripts/dev/start.sh`

**Purpose:** Master script to run both API and frontend together

**Features:**

- ✅ Starts both services in parallel
- ✅ Shows startup messages
- ✅ Checks for `.env` file existence
- ✅ Graceful shutdown with Ctrl+C (cleans up all processes)
- ✅ Trap handlers for proper cleanup

**Usage:**

```bash
./scripts/dev/start.sh
```

**Output:**

```
🚀 Starting LunaSentri development environment...

📡 Starting API server on port 80...
🌐 Starting Next.js frontend on port 3002...

✅ Development environment running!
   API:      http://localhost:80
   Frontend: http://localhost:3002

Press Ctrl+C to stop all services
```

### 4. `scripts/dev/README.md`

**Purpose:** Comprehensive documentation for the development scripts

**Contents:**

- Quick start guide
- Individual script usage
- Environment variable documentation
- Troubleshooting section
- Comparison with manual commands

## Verification

### Test 1: API Server with Telegram Enabled

```bash
$ ./scripts/dev/run-api.sh

2025/10/10 05:56:38 Database initialized at: ./data/lunasentri.db
2025/10/10 05:56:38 Admin bootstrap skipped: ADMIN_EMAIL or ADMIN_PASSWORD not set
2025/10/10 05:56:38 Auth service initialized (access token TTL: 15m0s, password reset TTL: 1h0m0s)
2025/10/10 05:56:38 Warning: Secure cookie flag disabled - only use in development
2025/10/10 05:56:38 LOCAL_HOST_METRICS disabled - metrics require registered machines
2025/10/10 05:56:38 Telegram notifications enabled  ✅
2025/10/10 05:56:38 LunaSentri API starting on port 80
```

**Success!** The log shows `Telegram notifications enabled` confirming that `TELEGRAM_BOT_TOKEN` was properly loaded.

### Test 2: Telegram Message Delivery

```bash
2025/10/10 05:57:01 [TELEGRAM] delivered to chat_id=8385128848 ✅
```

The API successfully sent a Telegram message, proving the integration works end-to-end.

## How It Works

### Environment Loading Pattern

```bash
# The key pattern used in all scripts:
if [ -f ../../.env ]; then
  set -a          # auto-export all variables
  source ../../.env  # loads TELEGRAM_BOT_TOKEN, etc.
  set +a          # stop auto-exporting
fi
```

This pattern:

1. Checks if `.env` exists
2. Enables auto-export mode (`set -a`)
3. Sources the `.env` file (all variables become exported)
4. Disables auto-export mode (`set +a`)

### Path Resolution

Scripts use relative paths from `scripts/dev/` to the project root:

- `../../.env` → Repository root `.env` file
- Changes to absolute paths for `apps/api-go` and `apps/web-next`

## Updated Documentation

### Main README.md

Updated the "Local Development" section to:

1. **Recommend the scripts first** (Option 1)
2. Keep manual setup as Option 2
3. Note about Telegram notifications requiring `TELEGRAM_BOT_TOKEN`
4. Link to `scripts/dev/README.md` for details

### Quick Start Flow

```
1. Clone repo
2. Create .env (optional)
3. pnpm install
4. ./scripts/dev/start.sh
5. Visit http://localhost:3002
```

## Benefits

✅ **Simpler workflow** - One command starts everything  
✅ **Environment consistency** - Same `.env` for all services  
✅ **Feature completeness** - Telegram notifications work out of the box  
✅ **Better DX** - Clear startup messages and status  
✅ **Safer execution** - `set -euo pipefail` catches errors early  
✅ **Graceful shutdown** - Ctrl+C cleans up all processes  
✅ **Well documented** - Comprehensive README in scripts directory  

## Migration Path

### For Existing Developers

**Old terminal commands:**

```bash
# Terminal 1
cd apps/api-go && AUTH_JWT_SECRET=... go run cmd/api/main.go

# Terminal 2  
cd apps/web-next && npm run dev
```

**New simplified workflow:**

```bash
./scripts/dev/start.sh
```

### Environment Variables

Move all configuration to `.env`:

```properties
# .env
TELEGRAM_BOT_TOKEN=your_token_here
AUTH_JWT_SECRET=your_secret_here
# ... other variables
```

## Next Steps (Optional Enhancements)

1. **Add `.env.example`** - Template for required variables
2. **Pre-commit hook** - Check if `.env` exists
3. **Health check** - Wait for API to be ready before starting frontend
4. **Log files** - Optional logging to files in `project/logs/dev/`
5. **Hot reload** - Auto-restart on `.env` changes

## Files Summary

| File | Lines | Purpose |
|------|-------|---------|
| `scripts/dev/run-api.sh` | 25 | Start API server |
| `scripts/dev/run-web.sh` | 14 | Start frontend |
| `scripts/dev/start.sh` | 42 | Start both services |
| `scripts/dev/README.md` | 100+ | Documentation |
| **Total** | **~180 lines** | Complete dev environment |

All scripts are executable (`chmod +x`) and use `#!/usr/bin/env bash` for portability.
