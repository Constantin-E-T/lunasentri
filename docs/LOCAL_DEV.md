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

#### Backend (apps/api-go)

The backend requires authentication configuration and supports optional CORS, database path, and admin user bootstrapping:

```bash
# Required: JWT secret for session token signing
export AUTH_JWT_SECRET=your-secret-key-min-32-chars-recommended

# Optional: Disable secure cookie for local development (default: true)
export SECURE_COOKIE=false

# Optional: Access token time-to-live (default: 15m)
export ACCESS_TOKEN_TTL=15m

# Optional: Set allowed CORS origin (defaults to http://localhost:3000)
export CORS_ALLOWED_ORIGIN=http://localhost:3000

# Optional: Set database path (defaults to ./data/lunasentri.db)
export DB_PATH=./data/lunasentri.db

# Optional: Bootstrap admin user on startup (for first-run seeding)
export ADMIN_EMAIL=admin@yourdomain.com
export ADMIN_PASSWORD=your_secure_password
```

**Authentication Configuration:**

- `AUTH_JWT_SECRET` is **required** - the server will not start without it
- Use a strong random string (32+ characters recommended)
- Generate with: `openssl rand -base64 32` or `python3 -c "import secrets; print(secrets.token_urlsafe(32))"`
- `SECURE_COOKIE=false` is **required for local development over HTTP** - allows cookies to work on `localhost`
  - Default is `true` (production mode - requires HTTPS)
  - Set to `false` in development to enable login over `http://localhost`
  - **Warning**: Never set to `false` in production - cookies would be vulnerable to interception
- `ACCESS_TOKEN_TTL` accepts Go duration format: `15m`, `1h`, `24h`, etc.
- Session cookies are HttpOnly and SameSite=Lax for security

**Admin User Bootstrapping Notes:**

- If both `ADMIN_EMAIL` and `ADMIN_PASSWORD` are set, an admin user will be created or updated on startup
- Useful for first-run initialization or password rotation
- The password is hashed with bcrypt (cost 12) before storage
- **Security**: Use strong passwords and consider rotating credentials after initial setup
- **Never log**: Raw passwords are never logged; only email and user ID are logged

**Authentication Flow:**

1. Set `ADMIN_EMAIL` and `ADMIN_PASSWORD` environment variables
2. Start the backend - admin user will be created automatically
3. Login via `POST /auth/login` with your credentials
4. Session cookie (`lunasentri_session`) will be set automatically
5. Frontend must use `credentials: 'include'` in fetch calls to send cookies
6. Protected endpoints (`/metrics`, `/ws`, `/auth/me`) require valid session

If not set, the Go API defaults to allowing requests from `http://localhost:3000` and stores the SQLite database at `./data/lunasentri.db`.

### 3. Start Backend (Terminal 1)

```bash
cd apps/api-go
go run main.go
```

The API will be available at `http://localhost:8080` with CORS enabled for the frontend.

Expected output:

```
Database initialized at: ./data/lunasentri.db
LunaSentri API starting on port 8080 (endpoints: /, /health, /metrics, /ws) with CORS origin: http://localhost:3000
```

### 4. Start Frontend (Terminal 2)

```bash
# From root directory
pnpm dev:web
```

The web interface will be available at `http://localhost:3000`

### 5. Verify Setup and Login

1. Open `http://localhost:3000` in your browser
2. You will be redirected to `http://localhost:3000/login`
3. Login with the admin credentials you set via environment variables:
   - Email: Value from `ADMIN_EMAIL`
   - Password: Value from `ADMIN_PASSWORD`
4. After successful login, you'll be redirected to the dashboard
5. The metrics card should display CPU, Memory, Disk percentages, and uptime
6. Metrics should auto-update in real-time via WebSocket or polling every 5 seconds

**Login Flow**:
- First visit: → Redirected to `/login`
- Enter credentials → Session cookie set
- Redirected to dashboard `/`
- Logout button in header clears session

**Troubleshooting Authentication**:

- **"Invalid email or password"**: Check that `ADMIN_EMAIL` and `ADMIN_PASSWORD` match what you're entering
- **Login succeeds but immediately redirects to login**:
  - Ensure `SECURE_COOKIE=false` is set in backend environment variables
  - Check browser DevTools → Application → Cookies to verify `lunasentri_session` cookie is set
  - Verify both frontend and backend are running on `localhost` (not mixing `127.0.0.1` and `localhost`)
- **Metrics show "Please log in"**: Session expired or invalid, logout and login again

**Other Troubleshooting**:

- **CORS errors**:
  - Ensure the Go backend is running on port 8080
  - Check that `.env.local` exists in `apps/web-next/` with the correct API URL
  - Verify the backend logs show CORS origin matching your frontend URL
- **WebSocket connection fails**: Normal fallback to polling will occur automatically

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

#### Public Endpoints

- `GET /` - API welcome message
- `GET /health` - Health check endpoint (returns `{"status":"healthy"}`)
- `POST /auth/login` - Login with credentials (sets session cookie)
- `POST /auth/logout` - Logout (clears session cookie)

#### Protected Endpoints (require authentication)

- `GET /auth/me` - Get current user profile
- `GET /metrics` - System metrics (CPU, memory, disk, uptime)
- `WebSocket /ws` - Real-time metrics streaming (sends JSON every ~3 seconds)

#### Authentication Endpoints

**POST /auth/login**

Login with email and password. Returns user profile and sets HttpOnly session cookie.

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@yourdomain.com","password":"your_secure_password"}' \
  -c cookies.txt

# Response: {"id":1,"email":"admin@yourdomain.com"}
```

**POST /auth/logout**

Clears the session cookie.

```bash
curl -X POST http://localhost:8080/auth/logout \
  -b cookies.txt

# Response: 204 No Content
```

**GET /auth/me**

Get current authenticated user's profile. Requires valid session cookie.

```bash
curl http://localhost:8080/auth/me \
  -b cookies.txt

# Response: {"id":1,"email":"admin@yourdomain.com"}
```

**Frontend Usage:**

```javascript
// Login
const response = await fetch('http://localhost:8080/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  credentials: 'include', // Important: sends cookies
  body: JSON.stringify({ email, password })
});

// Access protected endpoint
const metrics = await fetch('http://localhost:8080/metrics', {
  credentials: 'include' // Important: sends cookies
});

// Logout
await fetch('http://localhost:8080/auth/logout', {
  method: 'POST',
  credentials: 'include'
});
```

#### WebSocket Usage

The `/ws` endpoint provides real-time streaming of system metrics via WebSocket connection. **Authentication required**.

```javascript
// Connect to WebSocket (from frontend)
// Note: Session cookie must be set (login first)
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = (event) => {
  const metrics = JSON.parse(event.data);
  console.log('Received metrics:', metrics);
  // Example: {"cpu_pct":15.2,"mem_used_pct":67.1,"disk_used_pct":23.4,"uptime_s":120.5}
};

ws.onopen = () => console.log('WebSocket connected');
ws.onclose = () => console.log('WebSocket disconnected');
ws.onerror = (error) => console.error('WebSocket error:', error);
```

**WebSocket Features:**

- Sends metrics JSON every 3 seconds automatically
- Validates Origin header against `CORS_ALLOWED_ORIGIN` (default: `http://localhost:3000`)
- **Requires valid session cookie** - validates authentication during upgrade
- Graceful handling of client disconnections
- Ping/pong frames for connection health
- Read/write timeouts for robustness

**Testing WebSocket (with authentication):**

```bash
# First login to get session cookie
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@yourdomain.com","password":"your_secure_password"}' \
  -c cookies.txt

# Using websocat (install: brew install websocat)
# Note: WebSocket auth requires special handling - use frontend or authenticated tool
websocat ws://localhost:8080/ws --origin http://localhost:3000

# Using wscat (install: npm install -g wscat)
wscat -c ws://localhost:8080/ws --origin http://localhost:3000
```

### Frontend (port 3000)

- Serves the Next.js application with live metrics dashboard

## Project Structure

```text
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

### Backend Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `AUTH_JWT_SECRET` | **Yes** | - | Secret key for JWT token signing (32+ characters recommended) |
| `SECURE_COOKIE` | **Yes for dev** | `true` | Set to `false` for local HTTP development, `true` for production HTTPS |
| `ACCESS_TOKEN_TTL` | No | `15m` | Session token lifetime (Go duration format: `15m`, `1h`, `24h`) |
| `CORS_ALLOWED_ORIGIN` | No | `http://localhost:3000` | Allowed CORS origin for API requests |
| `DB_PATH` | No | `./data/lunasentri.db` | Path to SQLite database file (directory will be created if needed) |
| `ADMIN_EMAIL` | No | - | Admin user email for bootstrap (requires `ADMIN_PASSWORD`) |
| `ADMIN_PASSWORD` | No | - | Admin user password for bootstrap (requires `ADMIN_EMAIL`) |

**Authentication Configuration:**

- `AUTH_JWT_SECRET` is **required** - server will not start without it
- Generate strong secret: `openssl rand -base64 32`
- `SECURE_COOKIE=false` is **required for localhost** - browsers reject secure cookies over HTTP
- Session cookies are HttpOnly and SameSite=Lax

**Admin Bootstrap Notes:**

- When both `ADMIN_EMAIL` and `ADMIN_PASSWORD` are set, an admin user is created/updated on startup
- Passwords are hashed with bcrypt (cost 12) for secure storage
- Useful for first-run initialization and password rotation
- Raw passwords are never logged to maintain security

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

## Continuous Integration (CI)

### GitHub Actions Workflow

The project uses GitHub Actions for automated testing on every push and pull request.

**Workflow file**: `.github/workflows/ci.yml`

### Jobs

#### 1. Backend (Go)

- **Triggers**: Push to `main`/`develop`, PRs to `main`
- **Runner**: Ubuntu latest
- **Steps**:
  1. Checkout code
  2. Set up Go 1.23 with caching
  3. Install pnpm 10.18.0 (for monorepo dependencies)
  4. Install root dependencies (`pnpm install --frozen-lockfile`)
  5. Verify Go dependencies (`go mod verify`)
  6. Build binary (`go build -v ./...`)
  7. Run tests with race detector (`go test -race ./...`)
  8. Run static analysis (`go vet ./...`)

#### 2. Frontend (Next.js)

- **Triggers**: Push to `main`/`develop`, PRs to `main`
- **Runner**: Ubuntu latest
- **Steps**:
  1. Checkout code
  2. Set up pnpm 10.18.0
  3. Set up Node.js 20 with pnpm caching
  4. Install dependencies (`pnpm install --frozen-lockfile`)
  5. Build production bundle (`pnpm --filter web-next build`)
  6. Type check TypeScript (`npx tsc --noEmit`)

### Running CI Checks Locally

Before pushing, you can run the same checks locally:

**Backend:**

```bash
cd apps/api-go
go mod verify
go build -v ./...
go test -race ./...
go vet ./...
```

**Frontend:**

```bash
pnpm install --frozen-lockfile
pnpm --filter web-next build
cd apps/web-next && npx tsc --noEmit
```

### CI Features

- **Caching**: Both jobs use caching to speed up builds
  - Go: Automatic module caching via `setup-go@v5`
  - pnpm: Automatic store caching via `setup-node@v4`
- **Race Detection**: Go tests run with `-race` flag to catch concurrency bugs
- **Type Safety**: TypeScript strict mode checks enforce type correctness
- **Monorepo Support**: Uses pnpm workspace filtering for selective builds

### Viewing CI Results

- Check the "Actions" tab in GitHub to see workflow runs
- Each job provides detailed logs and timing information
- Failed checks block PR merges (if branch protection is enabled)
