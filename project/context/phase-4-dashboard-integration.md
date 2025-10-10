# Phase 4 Dashboard Integration - Summary

**Date**: October 10, 2025  
**Status**: ✅ COMPLETE (Backend enablement)  
**Branch**: main

## Overview

Hooked the API into the multi-machine dashboard flow so the frontend selector can pull machine-scoped metrics, WebSocket streams, and system info. Agents can now enrich machine records with runtime/system metadata, and the UI surfaces live data once agents start reporting.

## Backend Changes

- `/metrics`, `/ws`, and `/system/info` now require `machine_id` (unless `LOCAL_HOST_METRICS=true` for local dev) and enforce per-user ownership.
- `AgentMetricsRequest` accepts optional `uptime_s` and nested `system_info` payload; values persist on the `machines` table and metrics history (`uptime_seconds`).
- Added `UpdateMachineSystemInfo` storage method + migration `012_machine_system_info` to persist platform, kernel, memory, disk, and last boot time.
- WebSocket handler polls the latest stored metrics per machine so dashboards stay live without local collectors.
- System info endpoint returns agent-provided metadata (hostname/platform/cpu/etc.) and falls back to dev-only host info when no machine is selected.

## Verification

```bash
cd apps/api-go
go test ./...
```

## Next Steps

- Agents: ensure system info payload continues to send real host metadata (consider validating unit for CPU/memory/disk).
- Dashboard: wire metrics history/alerts filtering to `machine_id` (follow-up ticket).
- Operations: schedule migration rollout with maintenance window; older SQLite files need migration `012`.

