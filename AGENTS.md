# AGENTS.md — How to work in LunaSentri

## Dev commands
- Install: `pnpm install`
- API: `make -C apps/api-go run` | checks: `make -C apps/api-go fmt vet`
- Web: `pnpm --filter web-next dev` | build: `pnpm --filter web-next build`

## Conventions
- Go: net/http, ports fixed, graceful shutdown preferred, JSON responses.
- Web: typed components, avoid heavy deps, read API URL from `NEXT_PUBLIC_API_URL`.

## CI / PR
- CI must pass (Go build+vet; Next build).
- Update docs when commands/env change.
- Do not touch Dockerfiles/CapRover unless the task says so.

## Roadmap hints (for agents)
- `/metrics`: {cpu_pct, mem_used_pct, disk_used_pct, uptime_s}
- `/ws`: stream metrics JSON every 3s
- Frontend: live charts (CPU, RAM, uptime) using Recharts
