# Debugging Heartbeat Issues

## Current Status

✅ **Working:**

- Offline detection after 2 minutes
- Offline notification sent (Telegram)
- User logged out → still received notification

❌ **Not Working:**

1. No "back online" notification when agent restarts
2. UI requires hard refresh to see status change
3. Real-time updates missing

---

## Issue 1: Missing "Back Online" Notification

### What should happen

1. Agent restarts → sends metrics immediately
2. Backend heartbeat monitor detects on next check (within 30s)
3. Status changes from "offline" → "online"
4. Backend sends 🟢 recovery notification

### Debug steps

```bash
# Check if agent is actually running after restart
docker exec lunasentri-test-server ps aux | grep lunasentri-agent

# Check agent logs (if it has any)
docker logs lunasentri-test-server 2>&1 | tail -50

# Check if metrics are being sent
# (You'll need to check your backend logs for incoming metrics)
```

### Possible causes

**A) Agent config file doesn't exist:**

```bash
# Check if config file exists
docker exec lunasentri-test-server ls -la /etc/lunasentri/agent.yaml
```

If it doesn't exist, the agent might not be starting. Try starting WITHOUT config:

```bash
# Start agent with command-line args instead
docker exec -d lunasentri-test-server /usr/local/bin/lunasentri-agent \
  --api-url https://lunasentri.app/api \
  --api-key YOUR_API_KEY \
  --interval 60
```

**B) Agent starts but crashes immediately:**

```bash
# Run agent in foreground to see errors
docker exec -it lunasentri-test-server /usr/local/bin/lunasentri-agent --config /etc/lunasentri/agent.yaml
# (Press Ctrl+C to stop)
```

**C) Backend heartbeat monitor not detecting recovery:**

This could be a bug in the heartbeat logic. Let me check the code...

---

## Issue 2: UI Requires Hard Refresh

This is **expected behavior** - here's why:

### Current Implementation

- Frontend fetches machine list from API on page load
- No WebSocket or polling for real-time updates
- Status updates only appear when you refresh

### Why the delay in heartbeat detection?

**Timeline:**

```
T+0s   : Agent stops
T+2m0s : Heartbeat monitor checks (sees last_seen > 2m) → marks offline
T+2m0s : Offline notification sent
T+2m0s : You restart agent
T+2m30s: Agent sends first metrics batch
T+3m0s : Heartbeat monitor checks again → sees new last_seen → marks online
T+3m0s : Online notification should be sent
```

**Why not immediate?**

- Heartbeat monitor runs every **30 seconds** (not continuously)
- It only checks on each interval tick
- So there's always a 0-30s delay for detection

### Solutions

**Option A: Add real-time updates to frontend (WebSocket)**

- Backend broadcasts status changes via WebSocket
- Frontend updates UI immediately
- More complex to implement

**Option B: Add polling in frontend**

- Frontend checks API every 30s for updates
- Simpler but more API calls
- Updates within 30s max

**Option C: Keep as-is**

- Users refresh when needed
- Simplest, lowest resource usage
- Good enough for MVP

---

## Issue 3: Recovery Notification Not Sent

Let me check the heartbeat logic for a potential bug...

### Suspected Issue

Looking at the heartbeat code, the recovery notification might not be sent if:

1. **Machine status wasn't marked offline in DB**
   - Frontend might show "offline" but DB still says "online"

2. **Notification record not cleared**
   - If `ClearMachineOfflineNotification()` fails, system thinks it already notified

3. **New status same as old status**
   - If the status computation has a bug

### Debug commands

```bash
# Check database directly
# (You'll need access to your production database)

# Check machine status
SELECT id, name, status, datetime(last_seen, 'localtime') 
FROM machines;

# Check notification records
SELECT machine_id, datetime(notified_at, 'localtime')
FROM machine_offline_notifications;
```

### Expected flow

**When agent stops:**

```sql
-- Machine status should be "offline"
UPDATE machines SET status = 'offline' WHERE id = X;

-- Notification record created
INSERT INTO machine_offline_notifications VALUES (X, current_timestamp);
```

**When agent restarts:**

```sql
-- Machine status should be "online"
UPDATE machines SET status = 'online' WHERE id = X;

-- Notification record deleted
DELETE FROM machine_offline_notifications WHERE machine_id = X;
```

---

## Quick Test Script

Run this to see exactly what's happening:

```bash
#!/bin/bash

echo "=== HEARTBEAT DEBUG TEST ==="
echo ""

echo "1. Checking if agent is running..."
docker exec lunasentri-test-server ps aux | grep lunasentri-agent | grep -v grep || echo "❌ Agent NOT running"

echo ""
echo "2. Checking config file..."
docker exec lunasentri-test-server cat /etc/lunasentri/agent.yaml 2>/dev/null || echo "❌ Config file missing"

echo ""
echo "3. Stopping agent..."
docker exec lunasentri-test-server pkill -f lunasentri-agent
sleep 2

echo ""
echo "4. Verifying agent stopped..."
docker exec lunasentri-test-server ps aux | grep lunasentri-agent | grep -v grep && echo "❌ Agent still running!" || echo "✅ Agent stopped"

echo ""
echo "5. Waiting 130 seconds for offline notification..."
echo "   (Check Telegram now)"
sleep 130

echo ""
echo "6. Starting agent in foreground (will show errors if any)..."
echo "   Press Ctrl+C after a few seconds if it runs OK"
docker exec -it lunasentri-test-server /usr/local/bin/lunasentri-agent \
  --api-url https://lunasentri.app/api \
  --interval 60 \
  --machine-name test-server-01

# If the above works, start in background:
# docker exec -d lunasentri-test-server /usr/local/bin/lunasentri-agent ...
```

---

## Recommendation

1. **First**, verify agent is actually sending metrics after restart
2. **Then**, check backend logs to see if recovery is detected
3. **Finally**, we may need to fix a bug in the heartbeat recovery logic

Let me know what you find and I can help debug further!
