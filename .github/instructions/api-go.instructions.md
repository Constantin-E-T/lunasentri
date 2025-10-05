---
applyTo: "apps/api-go/**"
---
- Keep port 8080.
- Add `/metrics` and `/ws` as separate handlers.
- Use gopsutil via a `metrics.Service` interface for testability.
- Run `make fmt vet` and `go build ./...` before proposing a PR.
