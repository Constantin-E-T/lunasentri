# Local Development Guide

## Prerequisites

- **Go**: v1.21 or higher
- **Node.js**: v20 or higher
- **pnpm**: v10 or higher

## Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd lunasentri
pnpm install
```

### 2. Configure Environment Variables

#### Frontend (apps/web-next)

Create `apps/web-next/.env.local`:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
```

This tells the Next.js frontend where to find the Go API. The `NEXT_PUBLIC_` prefix makes the variable available to the browser.

#### Backend (apps/api-go) - Optional

The backend supports CORS configuration via environment variable:

```bash
# Optional: Set allowed CORS origin (defaults to http://localhost:3000)
export CORS_ALLOWED_ORIGIN=http://localhost:3000
```

If not set, the Go API defaults to allowing requests from `http://localhost:3000`.

### 3. Start Backend (Terminal 1)

```bash
cd apps/api-go
go run main.go
```

The API will be available at `http://localhost:8080` with CORS enabled for the frontend.

Expected output:
```
LunaSentri API starting on port 8080 (endpoints: /, /health, /metrics) with CORS origin: http://localhost:3000
```

### 4. Start Frontend (Terminal 2)

```bash
# From root directory
pnpm dev:web
```

The web interface will be available at `http://localhost:3000`

### 5. Verify Setup

1. Open `http://localhost:3000` in your browser
2. You should see the LunaSentri dashboard with a live metrics card
3. The metrics card should display CPU, Memory, Disk percentages, and uptime
4. Metrics should auto-update every 5 seconds

**Troubleshooting**: If you see a CORS error or the metrics card shows an error:
- Ensure the Go backend is running on port 8080
- Check that `.env.local` exists in `apps/web-next/` with the correct API URL
- Verify the backend logs show CORS origin matching your frontend URL

## Development Commands

### Root Workspace Commands

```bash
# Install all dependencies
pnpm install

# Start frontend development server
pnpm dev:web

# Build frontend for production
pnpm build:web
```

### Backend (apps/api-go)

```bash
# Run development server
make run
# or
go run main.go

# Build binary
make build

# Run tests
make test

# Format code
make fmt

# Run all checks (format, vet, test)
make check

# Clean build artifacts
make clean
```

### Frontend (direct commands in apps/web-next)

```bash
# Install dependencies (from root preferred)
pnpm install

# Run development server (with Turbopack)
pnpm dev

# Build for production (with Turbopack)
pnpm build

# Start production server
pnpm start
```

## API Endpoints

### Backend (port 8080)

- `GET /` - API welcome message
- `GET /health` - Health check endpoint (returns `{"status":"healthy"}`)
- `GET /metrics` - System metrics (CPU, memory, disk, uptime)

### Frontend (port 3000)

- Serves the Next.js application with live metrics dashboard

## Project Structure

```
lunasentri/
├── pnpm-workspace.yaml   # Workspace configuration
├── package.json          # Root package with workspace scripts
├── pnpm-lock.yaml        # Dependency lockfile
├── .npmrc                # pnpm configuration
├── apps/
│   ├── api-go/           # Go backend
│   │   ├── main.go       # Main server file
│   │   ├── Makefile      # Build commands
│   │   └── Dockerfile    # Container build
│   └── web-next/         # Next.js frontend
│       ├── app/          # App Router pages
│       ├── Dockerfile    # Container build (uses pnpm)
│       └── package.json  # Dependencies
├── deploy/               # Deployment configs
└── docs/                 # Documentation
```

## Docker Development

### Build and Run Backend

```bash
cd apps/api-go
docker build -t lunasentri-api .
docker run -p 8080:8080 lunasentri-api
```

### Build and Run Frontend

```bash
cd apps/web-next
docker build -t lunasentri-web .
docker run -p 3000:3000 lunasentri-web
```

## Environment Variables Reference

### Frontend (`apps/web-next/.env.local`)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | Yes | `http://localhost:8080` | Base URL for the Go API backend |

**Note**: The `.env.local` file is gitignored. You must create it manually on each development machine.

### Backend (Optional Environment Variables)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CORS_ALLOWED_ORIGIN` | No | `http://localhost:3000` | Allowed CORS origin for API requests |

## Troubleshooting

### Common Issues

1. **Port already in use**: Make sure no other services are running on ports 3000 or 8080
2. **Node.js version**: Ensure you're using Node.js v20 or higher
3. **Go version**: Ensure you're using Go v1.21 or higher
4. **CORS errors**:
   - Verify `.env.local` exists in `apps/web-next/` with `NEXT_PUBLIC_API_URL=http://localhost:8080`
   - Check backend logs show CORS origin: `http://localhost:3000`
   - Ensure both servers are running on their default ports
5. **Metrics card shows error**:
   - Confirm Go backend is running: `curl http://localhost:8080/metrics`
   - Check browser DevTools console for network errors
   - Verify environment variables are set correctly

### Logs

- Backend logs appear in the terminal where `go run main.go` is executed
- Frontend logs appear in the terminal where `npm run dev` is executed
- Browser DevTools Network tab shows API requests to `/metrics` endpoint
