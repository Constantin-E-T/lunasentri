# Development Scripts

This directory contains helper scripts to run the LunaSentri development environment.

## Quick Start

### Start Everything (Recommended)

```bash
./scripts/dev/start.sh
```

This will:

- Load environment variables from `.env` (including `TELEGRAM_BOT_TOKEN`)
- Start the Go API server on port 80
- Start the Next.js frontend on port 3002
- Properly handle cleanup when you press Ctrl+C

### Start Individual Services

#### API Server Only

```bash
./scripts/dev/run-api.sh
```

Starts the Go API server on port 80 with:

- Environment variables loaded from `../../.env`
- Default development settings
- Telegram notifications enabled (if `TELEGRAM_BOT_TOKEN` is set)

#### Frontend Only

```bash
./scripts/dev/run-web.sh
```

Starts the Next.js development server on port 3002.

## Environment Variables

The scripts will automatically load variables from the repository root `.env` file:

```properties
# .env example
TELEGRAM_BOT_TOKEN=your_bot_token_here
```

### API Environment Variables

The `run-api.sh` script sets these defaults (can be overridden in `.env`):

- `AUTH_JWT_SECRET=dev-secret` - JWT signing secret
- `SECURE_COOKIE=false` - Disable secure cookies for local dev
- `LOCAL_HOST_METRICS=false` - Require registered machines for metrics
- `CORS_ALLOWED_ORIGIN=http://localhost:3002` - Frontend URL

### Adding New Environment Variables

1. Add them to `.env` in the project root
2. The scripts will automatically export them using `set -a; source .env; set +a`

## Features

✅ **Auto-cleanup**: Kills existing processes on the ports before starting  
✅ **Environment loading**: Reads `.env` from project root  
✅ **Error handling**: `set -euo pipefail` for safer script execution  
✅ **Graceful shutdown**: Ctrl+C properly stops all services  

## Troubleshooting

### Port Already in Use

The scripts automatically kill processes on ports 80 and 3002 before starting.

### Telegram Not Working

Check that your `.env` file contains a valid `TELEGRAM_BOT_TOKEN`:

```bash
# Verify .env exists and has the token
cat .env | grep TELEGRAM_BOT_TOKEN
```

When the API starts, you should see:

```
Telegram notifications enabled
```

If you see "Telegram notifier not initialized", the token is missing or invalid.

### Permission Denied

Make scripts executable:

```bash
chmod +x scripts/dev/*.sh
```

## Manual Commands (Old Way)

If you need to run commands manually:

```bash
# API
cd apps/api-go
TELEGRAM_BOT_TOKEN=your_token AUTH_JWT_SECRET=dev-secret \
  SECURE_COOKIE=false LOCAL_HOST_METRICS=false \
  CORS_ALLOWED_ORIGIN=http://localhost:3002 \
  go run cmd/api/main.go

# Frontend
cd apps/web-next
npm run dev -- --port 3002
```

**But using the scripts is much easier!** 🚀
