# LunaSentri - AI Coding Agent Instructions

## Project Overview
LunaSentri is a lightweight server monitoring dashboard designed for solo developers. It's built as a monorepo with two main applications that communicate via REST API.

## Architecture
- **Monorepo structure**: Two independent apps in `apps/` directory
- **Backend (`apps/api-go/`)**: Minimal Go HTTP server using standard library
- **Frontend (`apps/web-next/`)**: Next.js 15 with App Router and Tailwind CSS v4
- **Communication**: REST API between Go backend (port 8080) and Next.js frontend (port 3000)

## Key Technologies & Versions
- **Go**: v1.21 with standard `net/http` library (no frameworks)
- **Next.js**: v15.5.4 with Turbopack enabled for dev/build
- **React**: v19.1.0 (latest)
- **Tailwind CSS**: v4 with PostCSS inline theme configuration
- **TypeScript**: Strict mode enabled with Next.js path aliases (`@/*`)

## Development Workflows

### Starting the stack
```bash
# Backend (from apps/api-go/)
go run main.go

# Frontend (from apps/web-next/)
npm run dev  # Uses --turbopack flag
```

### Building
```bash
# Backend: Standard Go build
go build

# Frontend: Uses Turbopack
npm run build  # Includes --turbopack flag
```

## Project-Specific Patterns

### Go Backend Conventions
- **No external frameworks**: Uses only standard library `net/http`
- **JSON responses**: Set `Content-Type: application/json` headers manually
- **Module path**: `github.com/Constantin-E-T/lunasentri/apps/api-go`
- **Health check endpoint**: Always implement `/health` returning `{"status":"healthy"}`

### Next.js Frontend Conventions
- **App Router**: All pages in `app/` directory (not `pages/`)
- **Turbopack**: Enabled by default for both dev and build
- **Styling**: Dark theme with slate color palette (`from-slate-900 to-slate-800`)
- **UI Pattern**: Gradient backgrounds with backdrop-blur glass effects
- **Typography**: Large headings (text-6xl) with emoji icons (🌙 for LunaSentri)

### Tailwind CSS v4 Specific
- **Import**: Use `@import "tailwindcss"` (not v3 syntax)
- **Theme configuration**: Inline themes with `@theme inline` directive
- **CSS variables**: Custom properties defined in `:root` for background/foreground
- **PostCSS**: Uses `@tailwindcss/postcss` plugin

## Code Style & Standards
- **Go**: Standard formatting, simple handler functions
- **TypeScript**: Strict mode, use `Readonly<>` for props interfaces
- **React**: Functional components only, no class components
- **Imports**: Use `@/*` path aliases for internal modules

## Development Status
This is an early-stage project with basic scaffolding. When adding features:
- Maintain the minimal, lightweight philosophy
- Keep the Go backend simple (standard library preferred)
- Follow the established dark theme and glass-morphism design patterns
- Ensure both apps can run independently for development

## File Locations
- Main Go server: `apps/api-go/main.go`
- Next.js entry: `apps/web-next/app/page.tsx`
- Global styles: `apps/web-next/app/globals.css`
- Type definitions: Use built-in Next.js and React types