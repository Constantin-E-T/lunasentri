# Testing Heartbeat Monitoring in Production

## Your Test Server: lunasentri-test-server

Container ID: `074cb7d12014`  
Status: Currently **Running** ✅

---

## Test 1: Trigger Offline Notification

### Stop the agent to simulate machine going offline

```bash
# Stop the LunaSentri agent inside the container
docker exec lunasentri-test-server pkill -f lunasentri-agent

# Or if that doesn't work, try:
docker exec lunasentri-test-server killall lunasentri-agent
```

### What happens next

1. **Immediately**: Agent stops sending metrics
2. **After 2 minutes**: Backend heartbeat monitor detects `last_seen > 2m`
3. **Notification sent**: You should receive:
   - 🔴 Telegram message: "Machine Offline - test-server-01"
   - Webhook event (if configured): `machine.offline`

### Monitor the backend logs

```bash
# Watch for offline detection (adjust if using different logging)
# You'll see a log line like:
# "Machine X (test-server-01) went offline (last seen: 2m15s ago)"
```

---

## Test 2: Trigger Recovery Notification

### Restart the agent

```bash
# Start the agent again
docker exec -d lunasentri-test-server /usr/local/bin/lunasentri-agent --config /etc/lunasentri/agent.yaml

# Or if the path is different, find it first:
docker exec lunasentri-test-server which lunasentri-agent
```

### What happens next

1. **Within 60s**: Agent reports first metrics batch
2. **Next heartbeat check** (within 30s): Backend detects recovery
3. **Notification sent**: You should receive:
   - 🟢 Telegram message: "Machine Back Online - test-server-01"
   - Webhook event (if configured): `machine.online`

---

## Test 3: Check if agent is running

```bash
# Check if the agent process is running
docker exec lunasentri-test-server ps aux | grep lunasentri-agent

# Check agent logs
docker exec lunasentri-test-server journalctl -u lunasentri-agent -f
# (if using systemd inside container)

# Or check plain logs
docker logs lunasentri-test-server 2>&1 | grep -i lunasentri
```

---

## Test 4: Verify notifications were sent

### Check backend logs for heartbeat activity

You should see logs like:

```
[INFO] Heartbeat monitor started (interval: 30s, threshold: 2m0s)
[INFO] Machine 1 (test-server-01) went offline (last seen: 2m15s ago)
[INFO] Machine 1 (test-server-01) came back online
```

### Check Telegram

Look for messages in your connected Telegram chat.

### Check Webhooks (if configured)

Look at your webhook endpoint logs for POST requests with:

- Event: `machine.offline`
- Event: `machine.online`

---

## Quick Test Sequence

Run this complete test in one go:

```bash
# 1. Stop agent
echo "🛑 Stopping agent..."
docker exec lunasentri-test-server pkill -f lunasentri-agent

# 2. Verify it stopped
echo "✓ Checking if stopped..."
docker exec lunasentri-test-server ps aux | grep lunasentri-agent | grep -v grep

# 3. Wait for offline detection (2 minutes)
echo "⏱️  Waiting 2 minutes for offline detection..."
echo "   Check your Telegram for offline notification..."
sleep 130

# 4. Restart agent
echo "🔄 Restarting agent..."
docker exec -d lunasentri-test-server /usr/local/bin/lunasentri-agent --config /etc/lunasentri/agent.yaml

# 5. Verify it started
echo "✓ Checking if started..."
sleep 5
docker exec lunasentri-test-server ps aux | grep lunasentri-agent | grep -v grep

# 6. Wait for recovery notification (30-90 seconds)
echo "⏱️  Waiting for recovery notification..."
echo "   Check your Telegram for online notification..."
sleep 90

echo "✅ Test complete! Check your notifications."
```

---

## Adding More Test Servers

You have two options:

### Option 1: Create another Docker container (Quick & Easy)

```bash
# Create a second test container
docker run -d \
  --name lunasentri-test-server-02 \
  --hostname test-server-02 \
  ubuntu:22.04 \
  sleep infinity

# Install the agent in it
docker exec lunasentri-test-server-02 bash -c "
  apt-get update && apt-get install -y curl && \
  curl -O https://lunasentri.app/downloads/lunasentri-agent && \
  chmod +x lunasentri-agent && \
  mv lunasentri-agent /usr/local/bin/
"

# Configure and start (you'll need an API key from the UI first)
docker exec -d lunasentri-test-server-02 /usr/local/bin/lunasentri-agent \
  --api-url https://lunasentri.app/api \
  --api-key YOUR_API_KEY_HERE \
  --interval 60
```

### Option 2: Use a real Linux VM/server

If you have access to a Linux machine (DigitalOcean, AWS, your laptop with Linux VM, etc.):

```bash
# SSH into the machine
ssh your-server

# Download and install agent
curl -O https://lunasentri.app/downloads/lunasentri-agent
chmod +x lunasentri-agent
sudo mv lunasentri-agent /usr/local/bin/

# Run the agent
lunasentri-agent \
  --api-url https://lunasentri.app/api \
  --api-key YOUR_API_KEY_HERE \
  --interval 60
```

### Get API Keys for new machines

1. **Go to**: <https://lunasentri.app>
2. **Login** with your account
3. **Navigate to**: Machines page
4. **Click**: "Add Machine" or "Register Machine"
5. **Copy**: The generated API key
6. **Use**: That key in the agent commands above

---

## Expected Timeline

| Time | Event |
|------|-------|
| T+0s | Stop agent |
| T+2m | Backend detects offline → 🔴 Offline notification sent |
| T+2m | Restart agent |
| T+3m | Agent sends first metrics |
| T+3m30s | Backend detects recovery → 🟢 Online notification sent |

---

## Troubleshooting

**Agent won't stop:**

```bash
# Force kill
docker exec lunasentri-test-server pkill -9 lunasentri-agent
```

**Agent won't start:**

```bash
# Check if binary exists
docker exec lunasentri-test-server ls -la /usr/local/bin/lunasentri-agent

# Run manually to see errors
docker exec -it lunasentri-test-server /usr/local/bin/lunasentri-agent --help
```

**No notifications received:**

- Check Telegram bot is configured
- Check webhook endpoints are active
- Verify backend logs show heartbeat monitor running
- Wait the full 2+ minutes for offline detection

**Container not found:**

```bash
# List all containers
docker ps -a

# Use the actual container name from the list
docker exec <actual-name> ...
```
