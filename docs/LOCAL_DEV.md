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

### 2. Start Backend (Terminal 1)

```bash
cd apps/api-go
go run main.go
```

The API will be available at `http://localhost:8080`

### 3. Start Frontend (Terminal 2)

```bash
# From root directory
pnpm dev:web
```

The web interface will be available at `http://localhost:3000`

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
- `GET /health` - Health check endpoint

### Frontend (port 3000)

- Serves the Next.js application

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

## Troubleshooting

### Common Issues

1. **Port already in use**: Make sure no other services are running on ports 3000 or 8080
2. **Node.js version**: Ensure you're using Node.js v20 or higher
3. **Go version**: Ensure you're using Go v1.21 or higher

### Logs

- Backend logs appear in the terminal where `go run main.go` is executed
- Frontend logs appear in the terminal where `npm run dev` is executed
