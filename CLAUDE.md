# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
LunaSentri is a lightweight server monitoring dashboard for solo developers, built as a monorepo with a Go backend and Next.js frontend communicating via REST API.

## Architecture

### Monorepo Structure
```
lunasentri/
├── apps/
│   ├── api-go/        # Go backend (port 8080)
│   └── web-next/      # Next.js frontend (port 3000)
```

### Backend (apps/api-go)
- Minimal Go HTTP server using **standard library only** (`net/http`)
- Module path: `github.com/Constantin-E-T/lunasentri/apps/api-go`
- No external frameworks - keep it simple
- Always implement `/health` endpoint returning `{"status":"healthy"}`
- Manually set `Content-Type: application/json` headers

### Frontend (apps/web-next)
- Next.js 15 with App Router (pages in `app/` directory, not `pages/`)
- React 19 with functional components only
- Tailwind CSS v4 with PostCSS inline theme configuration
- TypeScript strict mode with `@/*` path aliases

### Communication
- Backend runs on port 80 (production) or 8080 (dev)
- Frontend runs on port 3000 (dev) or 80 (production)
- REST API for inter-app communication
- WebSocket for real-time metrics streaming

## Common Commands

### Development (Run Both Apps)
```bash
# Backend (from apps/api-go/)
go run main.go

# Frontend (from apps/web-next/)
npm run dev  # Uses Turbopack
```

### Building
```bash
# Backend
cd apps/api-go && go build

# Frontend
cd apps/web-next && npm run build  # Uses Turbopack
```

### Testing
```bash
# Backend
cd apps/api-go && go test ./...

# Frontend
cd apps/web-next && npm test  # If tests are added
```

## Technology-Specific Patterns

### Go Backend Conventions
- Use standard library only - no frameworks
- Simple handler functions registered with `http.HandleFunc`
- Go version: 1.21
- Standard Go formatting

### Next.js Frontend Conventions
- Turbopack enabled for dev and build (`--turbopack` flag)
- App Router architecture (not Pages Router)
- Geist font family from Vercel (`next/font`)

### Tailwind CSS v4 Specifics
- Import: `@import "tailwindcss"` (not v3 `@tailwind` directives)
- Theme configuration: `@theme inline` directive in globals.css
- CSS variables in `:root` for background/foreground colors
- PostCSS plugin: `@tailwindcss/postcss`
- Dark theme design with slate color palette and glass-morphism effects

### Design Patterns
- Dark gradients: `from-slate-900 to-slate-800`
- Glass effects: backdrop-blur with transparency
- Large headings: `text-6xl` typography
- Emoji icons: 🌙 for LunaSentri branding

## Code Standards
- **Go**: Standard formatting, minimal dependencies
- **TypeScript**: Strict mode, `Readonly<>` for props interfaces
- **React**: Functional components only
- **Imports**: Use `@/*` path aliases for Next.js internal modules

## Key File Locations
- Go server entry: `apps/api-go/main.go`
- Next.js home page: `apps/web-next/app/page.tsx`
- Global styles: `apps/web-next/app/globals.css`
- Layout wrapper: `apps/web-next/app/layout.tsx`

## Development Philosophy
This is an early-stage project - maintain minimal, lightweight approach:
- Prefer standard library solutions over external dependencies
- Keep backend simple and framework-free
- Both apps must run independently for development
- Follow established dark theme and modern UI patterns

## Authentication & Admin Users

### Admin User Creation
LunaSentri provides multiple ways to create admin users:

1. **Bootstrap via Environment Variables** (Production recommended)
   - Set `ADMIN_EMAIL` and `ADMIN_PASSWORD` environment variables
   - Admin user is created/updated on server startup
   - Implemented in `apps/api-go/internal/auth/bootstrap.go`

2. **First User Auto-Promotion** (Development)
   - First registered user automatically becomes admin
   - Implemented in `apps/api-go/internal/auth/user_management.go:59-82`
   - Uses `CountUsers()` to detect first user and `PromoteToAdmin()` to elevate privileges

3. **Manual Database Update** (Emergency)
   - Direct SQLite access to update `is_admin` flag
   - SQL: `UPDATE users SET is_admin = 1 WHERE email = 'user@example.com';`

### Key Files
- `apps/api-go/internal/auth/bootstrap.go` - Admin bootstrapping
- `apps/api-go/internal/auth/user_management.go` - User CRUD and first-user logic
- `apps/api-go/internal/storage/sqlite.go:316-361` - `UpsertAdmin()` function
- `apps/api-go/internal/storage/sqlite.go:516-524` - `CountUsers()` function
- `apps/api-go/internal/storage/sqlite.go:526-541` - `PromoteToAdmin()` function

### Security Notes
- Passwords are hashed using bcrypt
- JWT tokens for session management (configurable TTL)
- Secure cookies in production (HTTPOnly, Secure flags)
- Admin role required for user management and alert configuration
