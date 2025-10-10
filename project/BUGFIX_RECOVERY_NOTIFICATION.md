# Bug Fix: Recovery Notification Not Sent

## The Problem

**User reported:**

- ✅ Offline notification works after 2 minutes
- ❌ **No "back online" notification when agent restarts**
- ❌ UI requires hard refresh to see status changes

## Root Cause

The metrics endpoint (`RecordMetrics`) was automatically setting machine status to "online" every time metrics were received:

```go
// OLD CODE (BUG)
func (s *Service) RecordMetrics(...) error {
    // ...insert metrics...
    
    // BUG: Always sets status to "online"
    s.store.UpdateMachineStatus(ctx, machineID, "online", now)
    
    return nil
}
```

**Why this caused the issue:**

1. Agent stops → `last_seen` becomes old
2. After 2 min → Heartbeat monitor: `status = "offline"` → Send offline notification ✅
3. User restarts agent → Agent sends metrics immediately
4. **Metrics endpoint: `status = "online"`** ← BUG!
5. Heartbeat monitor checks → Sees `previousStatus = "online"` AND `newStatus = "online"`
6. **No transition detected** → No recovery notification sent ❌

The heartbeat monitor logic requires a state transition:

```go
// In heartbeat.go
if previousStatus == "offline" && newStatus == "online" {
    // Send recovery notification
}
```

But because the metrics endpoint already changed status to "online", the heartbeat monitor saw:

- `previousStatus = "online"` (already changed by metrics endpoint!)
- `newStatus = "online"` (computed from last_seen)
- **No transition** = No notification

## The Fix

### 1. Added new storage method: `UpdateMachineLastSeen`

**File:** `internal/storage/interface.go`

```go
// New method - updates ONLY last_seen, not status
UpdateMachineLastSeen(ctx context.Context, id int, lastSeen time.Time) error
```

**File:** `internal/storage/machines.go`

```go
func (s *SQLiteStore) UpdateMachineLastSeen(ctx context.Context, id int, lastSeen time.Time) error {
 query := `
  UPDATE machines
  SET last_seen = ?  -- Only update last_seen
  WHERE id = ?
 `
 // ... execute query ...
}
```

### 2. Updated metrics endpoint to NOT change status

**File:** `internal/machines/service.go`

```go
// NEW CODE (FIXED)
func (s *Service) RecordMetrics(...) error {
    // Insert metrics
    s.store.InsertMetrics(ctx, machineID, ...)
    
    // FIXED: Only update last_seen, status managed by heartbeat monitor
    s.store.UpdateMachineLastSeen(ctx, machineID, now)
    
    return nil
}
```

### 3. Updated test mocks

Added `UpdateMachineLastSeen` stub to all mock stores:

- `internal/auth/service_test.go`
- `internal/notifications/http_test.go`
- `internal/notifications/webhooks_test.go`
- `internal/notifications/telegram_http_test.go`

### 4. Fixed failing tests

Updated tests that expected metrics endpoint to set status to "online":

- `internal/machines/service_test.go` - `TestMachineService/RecordMetrics`
- `internal/http/agent_handlers_test.go` - `TestAgentMetrics/successful_metrics_ingestion`

## How It Works Now

```
Timeline:
---------
T+0s   : Agent stops
T+0s   : status = "online", last_seen = old

T+2m   : Heartbeat check
T+2m   : Detects: previousStatus="online", newStatus="offline"
T+2m   : UPDATE machines SET status='offline' WHERE id=X
T+2m   : 🔴 Send offline notification

T+3m   : User restarts agent
T+3m   : Agent sends metrics
T+3m   : Metrics endpoint: UPDATE machines SET last_seen=NOW WHERE id=X
T+3m   : status = "offline" (UNCHANGED!)

T+3m30s: Heartbeat check
T+3m30s: Detects: previousStatus="offline", newStatus="online" (computed from fresh last_seen)
T+3m30s: UPDATE machines SET status='online' WHERE id=X
T+3m30s: 🟢 Send recovery notification ✅
```

## Result

✅ **Offline notifications work** (already did)  
✅ **Recovery notifications now work**  
✅ **Heartbeat monitor is the single source of truth for machine status**  
✅ **Metrics endpoint only updates last_seen timestamp**  
✅ **All tests passing**

## Files Changed

**Modified:**

1. `internal/storage/interface.go` - Added `UpdateMachineLastSeen` method
2. `internal/storage/machines.go` - Implemented `UpdateMachineLastSeen`
3. `internal/machines/service.go` - Use `UpdateMachineLastSeen` instead of `UpdateMachineStatus`
4. `internal/auth/service_test.go` - Added mock method
5. `internal/notifications/http_test.go` - Added mock method
6. `internal/notifications/webhooks_test.go` - Added mock method
7. `internal/notifications/telegram_http_test.go` - Added mock method
8. `internal/machines/service_test.go` - Fixed test expectations
9. `internal/http/agent_handlers_test.go` - Fixed test expectations

**Test Results:**

```
✅ All packages: PASS
✅ Binary compilation: SUCCESS
```

## Testing

1. **Stop agent:**

   ```bash
   docker exec lunasentri-test-server pkill -f lunasentri-agent
   ```

2. **Wait 2+ minutes** → Should receive 🔴 offline notification

3. **Start agent:**

   ```bash
   docker exec -d lunasentri-test-server /usr/local/bin/lunasentri-agent --config /etc/lunasentri/agent.yaml
   ```

4. **Wait 30-90 seconds** → Should receive 🟢 recovery notification

## Deployment

This fix is ready to deploy. The change is backwards-compatible:

- New database method (no migration needed)
- Same API behavior for agents
- Improved correctness of status management

## Additional Notes

### UI Real-time Updates

The UI still requires refresh because there's no WebSocket/polling. This is a separate feature:

**Options:**

1. **WebSocket** - Real-time push updates (more complex)
2. **Polling** - Frontend checks every 30s (simpler)
3. **Keep as-is** - Manual refresh (current behavior)

For now, manual refresh is acceptable for MVP. Real-time updates can be added in a future release.

---

**Status:** ✅ Fixed and tested  
**Ready for deployment:** Yes
