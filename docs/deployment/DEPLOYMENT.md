# =€ LunaSentri Production Deployment Guide

## Overview
This guide covers deploying LunaSentri to CapRover (or any Docker-based platform).

## Architecture
- **Backend (Go)**: Port 8080 - REST API + WebSocket + Metrics Collection
- **Frontend (Next.js)**: Port 3000 - Dashboard UI

---

## Prerequisites
- CapRover instance running
- Domain names configured (e.g., `api.lunasentri.com`, `app.lunasentri.com`)
- Telegram Bot Token (from @BotFather)

---

## Environment Variables

### Backend (api-go)

#### Required:
```bash
# Authentication
AUTH_JWT_SECRET=<random-32-char-string>   # Generate with: openssl rand -base64 32

# Admin Bootstrap (first-time setup only)
ADMIN_EMAIL=admin@yourdomain.com
ADMIN_PASSWORD=<secure-password>

# Security
SECURE_COOKIE=true  # Must be true in production
```

#### Optional:
```bash
# Telegram Notifications
TELEGRAM_BOT_TOKEN=<your-bot-token>

# CORS (if frontend on different domain)
CORS_ORIGIN=https://app.lunasentri.com

# Database path (default: ./data/lunasentri.db)
DB_PATH=/app/data/lunasentri.db
```

### Frontend (web-next)

#### Required:
```bash
# Backend API URL (must be accessible from browser)
NEXT_PUBLIC_API_URL=https://api.lunasentri.com
```

---

## CapRover Deployment

### 1. Deploy Backend

```bash
cd deploy/caprover/backend
tar -czf deploy.tar.gz -C ../../.. \
  apps/api-go \
  go.mod \
  go.sum

# Deploy via CapRover CLI
caprover deploy -a lunasentri-api
```

**CapRover App Settings:**
- **Persistent Directory**: `/app/data` (for SQLite database)
- **Environment Variables**: Add all backend env vars above
- **Port Mapping**: 8080
- **Enable HTTPS**: Yes
- **Custom Domain**: api.lunasentri.com

### 2. Deploy Frontend

```bash
cd deploy/caprover/frontend
tar -czf deploy.tar.gz -C ../../.. \
  apps/web-next \
  package.json \
  pnpm-lock.yaml

# Deploy via CapRover CLI
caprover deploy -a lunasentri-web
```

**CapRover App Settings:**
- **Environment Variables**: `NEXT_PUBLIC_API_URL`
- **Port Mapping**: 3000
- **Enable HTTPS**: Yes
- **Custom Domain**: app.lunasentri.com

---

## Database Persistence

### CapRover Persistent Storage
1. Go to your backend app in CapRover
2. Navigate to **App Configs > Persistent Directories**
3. Add mapping: `/app/data` ’ `/app/data`
4. Save & Restart

This ensures your SQLite database survives container restarts.

---

## Post-Deployment Steps

### 1. Initialize Admin User
On first deployment, the backend will create an admin user with credentials from `ADMIN_EMAIL` and `ADMIN_PASSWORD`.

### 2. Login & Configure Telegram
1. Visit https://app.lunasentri.com
2. Login with admin credentials
3. Go to **Notifications ’ Telegram**
4. Click **"Connect Telegram"**
5. Start a chat with your bot: `t.me/YourBotUsername`
6. Send `/start` to the bot
7. Click **"Test Connection"** in LunaSentri

### 3. Set Up Alert Rules
1. Go to **Alerts** page
2. Create rules for CPU, Memory, Disk usage
3. Alerts will trigger Telegram notifications automatically

---

## Security Checklist

- [ ] `AUTH_JWT_SECRET` is a random 32+ character string
- [ ] `SECURE_COOKIE=true` in production
- [ ] HTTPS enabled for both frontend and backend
- [ ] Admin password is strong and unique
- [ ] Telegram bot token kept secret
- [ ] CORS origin matches your frontend domain

---

## Monitoring & Logs

### View Backend Logs
```bash
# CapRover
caprover logs -a lunasentri-api -f

# Docker
docker logs -f <container-id>
```

### Health Check
```bash
curl https://api.lunasentri.com/health
# Expected: {"status":"healthy"}
```

---

## Troubleshooting

### Database Issues
**Problem**: Database resets on every deployment
**Solution**: Ensure persistent directory `/app/data` is configured in CapRover

### Telegram Not Working
**Problem**: Telegram test works but alerts don't trigger notifications
**Solution**: Check backend logs for errors. Ensure `TELEGRAM_BOT_TOKEN` is set.

### CORS Errors
**Problem**: Frontend can't connect to backend API
**Solution**:
1. Check `NEXT_PUBLIC_API_URL` in frontend
2. Add `CORS_ORIGIN=https://app.lunasentri.com` to backend
3. Ensure both apps use HTTPS

### WebSocket Connection Fails
**Problem**: Real-time metrics not updating
**Solution**: Ensure WebSocket support is enabled in your reverse proxy (CapRover handles this automatically)

---

## Alternative Deployment Platforms

### Docker Compose
See `docker-compose.yml` in project root for local/self-hosted deployment.

### Railway / Render / Fly.io
Use the Dockerfiles directly:
- Backend: `apps/api-go/Dockerfile`
- Frontend: `apps/web-next/Dockerfile`

Set environment variables in platform dashboard.

---

## Backup & Restore

### Backup Database
```bash
# Copy from CapRover persistent volume
docker cp <container-id>:/app/data/lunasentri.db ./backup.db
```

### Restore Database
```bash
# Copy to CapRover persistent volume
docker cp ./backup.db <container-id>:/app/data/lunasentri.db
```

---

## Support

For issues or questions, check:
- GitHub Issues: [your-repo-url]
- Documentation: `/README.md`
