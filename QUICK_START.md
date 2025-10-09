# Quick Start - Development Workflow

## TL;DR

```bash
# Normal development (keeps your data)
./scripts/dev-reset.sh

# Fresh start (clears database)
./scripts/dev-reset.sh --reset-db
```

## What Changed

✅ **Database now persists by default** - No more losing your test data!  
✅ **Port conflicts auto-resolved** - No more manual cleanup  
✅ **Better error messages** - Logs captured automatically  

## Common Commands

### Start Dev Server (Keep Data)

```bash
./scripts/dev-reset.sh
```

Your users, alerts, webhooks, and settings are preserved.

### Start Fresh (Reset Database)

```bash
./scripts/dev-reset.sh --reset-db
```

Deletes the database. First user becomes admin.

### Stop Services

Press `Ctrl+C` in the terminal where you ran the script.

### View Logs

```bash
# Backend logs
tail -f apps/api-go/project/logs/backend.log

# Frontend logs  
tail -f apps/api-go/project/logs/frontend.log
```

### Manual Port Cleanup (if needed)

```bash
# Kill backend
lsof -ti:8080 | xargs kill -9

# Kill frontend
lsof -ti:3000 | xargs kill -9
```

## Testing Registration

### Via Frontend

1. Start server: `./scripts/dev-reset.sh --reset-db`
2. Visit: <http://localhost:3000/register>
3. Register with email + password (8+ chars)
4. First user automatically becomes admin

### Via curl

```bash
# Register
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"admin123456"}'

# Login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"email":"admin@test.com","password":"admin123456"}'

# Get profile
curl -X GET http://localhost:8080/auth/profile \
  -b cookies.txt
```

## Troubleshooting

### "Port already in use"

The script now handles this automatically. If it doesn't work:

```bash
lsof -ti:8080 | xargs kill -9
lsof -ti:3000 | xargs kill -9
./scripts/dev-reset.sh
```

### "Database is locked"

```bash
pkill -9 go
rm -f apps/api-go/data/lunasentri.db-*
./scripts/dev-reset.sh
```

### Registration fails with 500 error

```bash
# Check backend logs
tail -f apps/api-go/project/logs/backend.log

# Common fix: reset database
./scripts/dev-reset.sh --reset-db
```

### 401 Unauthorized errors

- CORS is configured correctly
- Make sure cookies are enabled in your browser
- Check that `SECURE_COOKIE=false` for localhost

## Environment Variables

Create `.env` in project root:

```bash
# Required
AUTH_JWT_SECRET=your-secret-key-here

# Optional
TELEGRAM_BOT_TOKEN=your-bot-token
NEXT_PUBLIC_API_URL=http://localhost:8080
FRONTEND_PORT=3000
```

## Development Workflow Example

```bash
# Monday morning
./scripts/dev-reset.sh
# Work on features, create test users, alerts...

# Stop for lunch (Ctrl+C)

# After lunch
./scripts/dev-reset.sh
# All your test data is still there!

# Friday - need clean state for testing
./scripts/dev-reset.sh --reset-db
# Fresh database, ready for integration tests
```

## Success Indicators

When everything works correctly:

```
✅ Checking for processes on port 8080...
✅ Checking for processes on port 3000...
✅ Database preserved: apps/api-go/data/lunasentri.db
✅ Starting backend on 8080...
✅ Starting frontend on 3000...
✅ Backend PID: 12345
✅ Frontend PID: 12346
✅ Database preserved. Existing users remain available.
✅ Visit http://localhost:3000
```

## Additional Resources

- Full details: [DEV_WORKFLOW_FIXES.md](./DEV_WORKFLOW_FIXES.md)
- Backend API docs: [apps/api-go/README.md](./apps/api-go/README.md)
- Frontend docs: [apps/web-next/README.md](./apps/web-next/README.md)
- Agent guidelines: [docs/AGENT_GUIDELINES.md](./docs/AGENT_GUIDELINES.md)
