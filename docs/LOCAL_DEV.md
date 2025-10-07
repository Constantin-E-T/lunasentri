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

# Optional: Password reset token time-to-live (default: 1h)
export PASSWORD_RESET_TTL=1h

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

**Registration Flow**:

- Navigate to `/register` or click "Create account" link from login page
- Enter email, password (8+ characters), and confirm password
- Client-side validation ensures passwords match and meet requirements
- On successful registration:
  - User account is created via `POST /auth/register`
  - First user to register becomes an administrator automatically
  - User is automatically logged in (session cookie set)
  - Redirected to dashboard `/`
- Already authenticated users are redirected to dashboard if they visit `/register`
- Admin users see "Manage Users" link in dashboard header; regular users don't

**Manage Users**:

- Click "Manage Users" link in dashboard header to access `/users` page
- **List Users**: View all users with email and creation date
- **Add User**:
  - Enter email (required)
  - Optionally enter password, or leave empty for auto-generated temp password
  - If temp password is generated, it will be displayed once in a dismissible alert
  - Share the temp password with the new user securely
- **Delete User**:
  - Click "Delete" button next to any user (except yourself)
  - Confirm deletion in dialog
  - Cannot delete your own account or the last remaining user

**Settings**:

- Click "Settings" link in dashboard header to access `/settings` page (available to all authenticated users)
- **Change Password**:
  - Enter current password for verification
  - Enter new password (8+ characters, different from current)
  - Confirm new password
  - On success, password is updated and user remains logged in
  - Common errors: incorrect current password, weak new password, password mismatch

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
- `POST /auth/forgot-password` - Request password reset token
- `POST /auth/reset-password` - Reset password using token

#### Protected Endpoints (require authentication)

- `GET /auth/me` - Get current user profile
- `POST /auth/change-password` - Change password (logged-in user)
- `GET /auth/users` - List all users
- `POST /auth/users` - Create a new user
- `DELETE /auth/users/{id}` - Delete a user by ID
- `GET /metrics` - System metrics (CPU, memory, disk, uptime)
- `WebSocket /ws` - Real-time metrics streaming (sends JSON every ~3 seconds)

#### Alert System Endpoints (require authentication)

- `GET /alerts/rules` - List all alert rules
- `POST /alerts/rules` - Create a new alert rule
- `PUT /alerts/rules/{id}` - Update an existing alert rule
- `DELETE /alerts/rules/{id}` - Delete an alert rule
- `GET /alerts/events` - List all alert events (with optional limit query parameter)
- `PUT /alerts/events/{id}/ack` - Acknowledge an alert event

#### Registration and Authentication

**POST /auth/register** (Public)

Self-serve user registration. The first user to register automatically becomes an admin.

```bash
# Register first user (becomes admin)
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"SecurePassword123"}'

# Response: {"id":1,"email":"admin@example.com","is_admin":true,"created_at":"2025-10-06T20:00:00Z"}

# Register additional users (non-admin)
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"AnotherPassword123"}'

# Response: {"id":2,"email":"user@example.com","is_admin":false,"created_at":"2025-10-06T20:01:00Z"}
```

**Registration Validation:**

- Email: Must be non-empty and contain `@`
- Password: Must be at least 8 characters long
- Duplicate emails return 409 Conflict

**First User Admin Promotion:**

- The first user to register automatically receives admin privileges
- This ensures the system always has at least one administrator
- Subsequent users are created as regular users (non-admin)

**POST /auth/login**

Login with email and password. Returns user profile (including `is_admin` flag) and sets HttpOnly session cookie.

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

**POST /auth/forgot-password**

Request a password reset token. Always returns 202 Accepted (doesn't reveal if user exists). In development, the reset token is returned in the response and logged to stdout. In production, this would be sent via email.

```bash
curl -X POST http://localhost:8080/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@yourdomain.com"}'

# Response (dev only): {"reset_token":"base64-encoded-token-here"}
# Status: 202 Accepted
```

**POST /auth/reset-password**

Reset password using a valid reset token. Token must not be expired or previously used.

```bash
curl -X POST http://localhost:8080/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{"token":"your-reset-token","password":"newpassword123"}'

# Response: 204 No Content (success)
# Or 400 Bad Request with error message
```

**Password Reset Flow (Development):**

1. Request reset token:

   ```bash
   curl -X POST http://localhost:8080/auth/forgot-password \
     -H "Content-Type: application/json" \
     -d '{"email":"admin@yourdomain.com"}'
   ```

2. Copy the `reset_token` from the response (also logged in server output)

3. Reset password with the token:

   ```bash
   curl -X POST http://localhost:8080/auth/reset-password \
     -H "Content-Type: application/json" \
     -d '{"token":"<token-from-step-1>","password":"mynewpassword"}'
   ```

4. Login with new password:

   ```bash
   curl -X POST http://localhost:8080/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"admin@yourdomain.com","password":"mynewpassword"}' \
     -c cookies.txt
   ```

**Password Reset Security:**

- Tokens expire after `PASSWORD_RESET_TTL` (default: 1 hour)
- Tokens are hashed before storage (SHA256)
- Tokens can only be used once
- Endpoint doesn't reveal whether email exists (timing-safe)
- Password must be at least 8 characters
- Old password stops working immediately after reset

**POST /auth/change-password**

Change password for the currently logged-in user. Requires authentication and knowledge of current password (self-service password change).

```bash
# Login first to get session cookie
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"oldpassword123"}' \
  -c cookies.txt

# Change password (requires current password)
curl -X POST http://localhost:8080/auth/change-password \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"current_password":"oldpassword123","new_password":"newpassword456"}'

# Response: 204 No Content (success)
# Or 401 Unauthorized (wrong current password)
# Or 400 Bad Request (weak new password)
```

**Password Change Requirements:**

- User must be logged in (valid session cookie)
- Must provide correct current password
- New password must be at least 8 characters
- New password must be different from current password
- Outstanding password reset tokens are automatically invalidated

**Error Responses:**

- `400 Bad Request`: Invalid request body or weak new password
- `401 Unauthorized`: Incorrect current password
- `404 Not Found`: User account not found (rare edge case)

**Security Notes:**

- Self-service: Users can change their own password without admin intervention
- Verifies current password before allowing change (prevents unauthorized password changes if session is compromised)
- New password is hashed with bcrypt (cost 12) before storage
- Old password stops working immediately after successful change
- All password reset tokens for the user are invalidated (prevents reset token reuse)
- Action is logged with user ID for audit trail

**Complete Change Password Flow:**

1. Login with current credentials:

   ```bash
   curl -X POST http://localhost:8080/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"user@example.com","password":"oldpassword123"}' \
     -c cookies.txt
   ```

2. Change password:

   ```bash
   curl -X POST http://localhost:8080/auth/change-password \
     -H "Content-Type: application/json" \
     -b cookies.txt \
     -d '{"current_password":"oldpassword123","new_password":"newpassword456"}'
   ```

3. Verify old password no longer works:

   ```bash
   curl -X POST http://localhost:8080/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"user@example.com","password":"oldpassword123"}'
   # Should return 401 Unauthorized
   ```

4. Login with new password:

   ```bash
   curl -X POST http://localhost:8080/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"user@example.com","password":"newpassword456"}' \
     -c cookies.txt
   # Should succeed
   ```

#### User Management Endpoints

**GET /auth/users**

List all users. Requires authentication.

```bash
curl -b cookies.txt http://localhost:8080/auth/users

# Response: [{"id":1,"email":"admin@example.com","created_at":"2024-01-01T12:00:00Z"},...]
```

**POST /auth/users**

Create a new user. If password is not provided, a secure temporary password is generated and returned. Requires authentication.

```bash
# Create user with password
curl -b cookies.txt -X POST http://localhost:8080/auth/users \
  -H "Content-Type: application/json" \
  -d '{"email":"newuser@example.com","password":"securepassword"}'

# Response: {"id":2,"email":"newuser@example.com","created_at":"2024-01-01T12:00:00Z"}

# Create user without password (temp password generated)
curl -b cookies.txt -X POST http://localhost:8080/auth/users \
  -H "Content-Type: application/json" \
  -d '{"email":"tempuser@example.com"}'

# Response: {"id":3,"email":"tempuser@example.com","created_at":"...","temp_password":"base64-encoded-password"}
```

**DELETE /auth/users/{id}**

Delete a user by ID. Cannot delete yourself or the last admin. Requires authentication.

```bash
curl -b cookies.txt -X DELETE http://localhost:8080/auth/users/2

# Response: 204 No Content (success)
# Or 403 Forbidden (cannot delete self or last admin)
# Or 404 Not Found (user doesn't exist)
```

**User Management Notes:**

- First user can register via `/auth/register` and becomes admin automatically
- Initial admin user can also be created from `ADMIN_EMAIL` and `ADMIN_PASSWORD` environment variables
- Additional users can be created via public registration or by authenticated users via the API
- Users cannot delete their own account (prevents accidental lockout)
- System prevents deletion of the last admin (ensures administrative access)
- Regular users (non-admin) can be deleted even if they're the only user
- Email format validation (must contain @)
- Duplicate emails are rejected with 409 Conflict
- All user responses now include the `is_admin` boolean field

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
| `PASSWORD_RESET_TTL` | No | `1h` | Password reset token lifetime (Go duration format: `30m`, `1h`, `2h`) |
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

## Alert System

LunaSentri includes a comprehensive alert system that monitors system metrics and triggers alerts when thresholds are breached. The alert system consists of two main components: **Alert Rules** (what to monitor) and **Alert Events** (when alerts are triggered).

### Alert Rules

Alert Rules define the conditions that trigger alerts. Each rule monitors a specific metric (CPU, memory, or disk usage) and triggers events when the metric crosses a defined threshold.

**Rule Properties:**

- **Name**: Human-readable identifier for the rule
- **Metric Type**: `cpu`, `memory`, or `disk`
- **Threshold**: Percentage value (0-100) that triggers the alert
- **Condition**: `above` or `below` the threshold
- **Consecutive Samples**: Number of consecutive metric samples that must meet the condition before triggering
- **Enabled**: Whether the rule is actively monitoring

**Example Alert Rules:**

- High CPU: Trigger when CPU usage is above 80% for 3 consecutive samples
- Low disk space: Trigger when disk usage is above 90% for 2 consecutive samples
- Memory pressure: Trigger when memory usage is above 85% for 5 consecutive samples

### Alert Events

Alert Events are generated when Alert Rules are triggered. Each event represents a specific instance of a rule being breached.

**Event Properties:**

- **Rule**: Reference to the Alert Rule that was triggered
- **Triggered At**: Timestamp when the event was created
- **Metric Value**: The actual metric value that caused the trigger
- **Acknowledged**: Whether an admin has acknowledged the event
- **Acknowledged At**: Timestamp of acknowledgment (if applicable)

### Real-time Monitoring

The alert system evaluates rules in real-time:

- **HTTP /metrics endpoint**: Evaluates all active rules on each request
- **WebSocket /ws endpoint**: Evaluates rules on each streamed metric sample (~3 seconds)
- **Consecutive Logic**: Tracks consecutive breach counts per rule to prevent false positives
- **Thread Safety**: Uses read/write locks for concurrent access to rule state

### Frontend Integration

The web interface provides complete alert management:

**Dashboard Integration:**

- Alert badge appears in navigation when unacknowledged events exist
- Badge shows count of unacknowledged events
- Quick access link to alerts page

**Alerts Page (`/alerts`):**

- **Create Rules**: Modal form for creating new alert rules with validation
- **Manage Rules**: Table view with edit/delete actions and enabled/disabled toggle
- **View Events**: List of recent alert events with timestamps and values
- **Acknowledge Events**: One-click acknowledgment to mark events as handled

**API Integration:**

- Uses custom `useAlerts` hook for state management
- Optimistic updates for better UX
- Automatic refresh and error handling
- SWR-style data fetching patterns

### Database Schema

The alert system uses two SQLite tables with foreign key constraints:

**alert_rules table:**

```sql
CREATE TABLE alert_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    metric_type TEXT NOT NULL CHECK (metric_type IN ('cpu', 'memory', 'disk')),
    threshold REAL NOT NULL CHECK (threshold >= 0 AND threshold <= 100),
    condition TEXT NOT NULL CHECK (condition IN ('above', 'below')),
    consecutive_samples INTEGER NOT NULL CHECK (consecutive_samples > 0),
    enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**alert_events table:**

```sql
CREATE TABLE alert_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rule_id INTEGER NOT NULL,
    triggered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    metric_value REAL NOT NULL,
    acknowledged BOOLEAN NOT NULL DEFAULT 0,
    acknowledged_at DATETIME,
    FOREIGN KEY (rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE
);
```

### Testing

The alert system includes comprehensive test coverage:

- **Unit Tests**: Service logic, storage operations, and HTTP handlers
- **Integration Tests**: End-to-end API workflows
- **Frontend Tests**: React hooks and component behavior
- **Test Coverage**: 100% of alert-related code paths

Run alert tests:

```bash
# Backend tests
cd apps/api-go
go test ./internal/alerts/... -v
go test ./internal/storage/... -v -run "Alert"

# Frontend tests  
cd apps/web-next
npm test -- useAlerts
```

- **Race Detection**: Go tests run with `-race` flag to catch concurrency bugs
- **Type Safety**: TypeScript strict mode checks enforce type correctness
- **Monorepo Support**: Uses pnpm workspace filtering for selective builds

### Viewing CI Results

- Check the "Actions" tab in GitHub to see workflow runs
- Each job provides detailed logs and timing information
- Failed checks block PR merges (if branch protection is enabled)
