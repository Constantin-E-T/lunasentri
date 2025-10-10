# LunaSentri Agent - Quick Start Guide

Get your server monitoring up and running in under 2 minutes! 🚀

## Prerequisites

- Linux server (Ubuntu, Debian, CentOS, RHEL, or similar)
- `curl` or `wget` installed
- Root or sudo access
- Your LunaSentri API key

## Getting Your API Key

1. Log in to your LunaSentri dashboard: https://lunasentri-web.serverplus.org
2. Navigate to **Machines** → **Add Machine**
3. Enter your server's hostname
4. Copy the generated API key

## One-Line Installation

### Method 1: Using curl (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/Constantin-E-T/lunasentri/main/apps/agent/scripts/install.sh | \
  sudo LUNASENTRI_API_KEY="your-api-key-here" \
  LUNASENTRI_SERVER_URL="https://lunasentri-api.serverplus.org" \
  bash
```

### Method 2: Using wget

```bash
wget -qO- https://raw.githubusercontent.com/Constantin-E-T/lunasentri/main/apps/agent/scripts/install.sh | \
  sudo LUNASENTRI_API_KEY="your-api-key-here" \
  LUNASENTRI_SERVER_URL="https://lunasentri-api.serverplus.org" \
  bash
```

## Manual Installation

If you prefer to inspect the installer first:

```bash
# Download the installer
curl -fsSL https://raw.githubusercontent.com/Constantin-E-T/lunasentri/main/apps/agent/scripts/install.sh -o install.sh

# Review it (recommended!)
less install.sh

# Make it executable
chmod +x install.sh

# Run with your API key
sudo LUNASENTRI_API_KEY="your-api-key-here" \
     LUNASENTRI_SERVER_URL="https://lunasentri-api.serverplus.org" \
     ./install.sh
```

## What the Installer Does

The installation script automatically:

1. ✅ Detects your Linux distribution
2. ✅ Creates a dedicated `lunasentri` system user
3. ✅ Installs the agent binary to `/usr/local/bin/lunasentri-agent`
4. ✅ Creates configuration file at `/etc/lunasentri/agent.yaml`
5. ✅ Sets up a systemd service for automatic startup
6. ✅ Starts the agent and enables it on boot

## Verify Installation

After installation, check that the agent is running:

```bash
# Check service status
sudo systemctl status lunasentri-agent

# View live logs
sudo journalctl -u lunasentri-agent -f

# Or check the log file directly
sudo tail -f /var/log/lunasentri/agent.log
```

You should see JSON logs like:
```json
{"level":"info","msg":"LunaSentri agent starting","version":"1.0.0"}
{"level":"info","msg":"System info collected","hostname":"your-server"}
{"level":"info","msg":"Metrics sent successfully","status_code":202}
```

## Check Your Dashboard

Go to https://lunasentri-web.serverplus.org/machines and you should see your server appear within 10 seconds! 🎉

## Common Commands

```bash
# Start the agent
sudo systemctl start lunasentri-agent

# Stop the agent
sudo systemctl stop lunasentri-agent

# Restart the agent
sudo systemctl restart lunasentri-agent

# View status
sudo systemctl status lunasentri-agent

# View logs
sudo journalctl -u lunasentri-agent -n 50

# Disable automatic startup
sudo systemctl disable lunasentri-agent

# Enable automatic startup
sudo systemctl enable lunasentri-agent
```

## Configuration

The agent configuration is stored at `/etc/lunasentri/agent.yaml`:

```yaml
server_url: https://lunasentri-api.serverplus.org
api_key: your-api-key-here
interval: 10s              # How often to send metrics
retry_backoff: 5s          # Wait time between retries
max_retries: 3             # Maximum retry attempts
system_info_period: 1h     # How often to refresh system info
```

After changing the configuration, restart the agent:
```bash
sudo systemctl restart lunasentri-agent
```

## Uninstallation

To remove the agent:

```bash
# Stop and disable the service
sudo systemctl stop lunasentri-agent
sudo systemctl disable lunasentri-agent

# Remove files
sudo rm -f /usr/local/bin/lunasentri-agent
sudo rm -rf /etc/lunasentri
sudo rm -f /etc/systemd/system/lunasentri-agent.service
sudo rm -rf /var/log/lunasentri
sudo rm -rf /var/lib/lunasentri

# Remove user (optional)
sudo userdel lunasentri

# Reload systemd
sudo systemctl daemon-reload
```

## Troubleshooting

### Agent won't start

1. Check the logs:
   ```bash
   sudo journalctl -u lunasentri-agent -n 50
   ```

2. Verify your API key is correct in `/etc/lunasentri/agent.yaml`

3. Test network connectivity:
   ```bash
   curl -I https://lunasentri-api.serverplus.org/health
   ```

### Metrics not appearing in dashboard

1. Check agent is running:
   ```bash
   sudo systemctl status lunasentri-agent
   ```

2. Look for errors in logs:
   ```bash
   sudo tail -f /var/log/lunasentri/agent.log
   ```

3. Verify API key matches the one in your dashboard

### Permission denied errors

The installer needs root access. Make sure to run with `sudo`:
```bash
sudo LUNASENTRI_API_KEY="..." ./install.sh
```

## Security Notes

- The agent runs as a dedicated `lunasentri` system user (not root)
- API key is stored in `/etc/lunasentri/agent.yaml` with 0644 permissions
- The agent only sends metrics data (read-only system access)
- All communication uses HTTPS encryption
- The systemd service includes security hardening (ProtectSystem, PrivateTmp, etc.)

## What Gets Monitored

The agent collects and sends:

- **CPU Usage** - Overall CPU percentage
- **Memory Usage** - RAM utilization percentage  
- **Disk Usage** - Root filesystem usage percentage
- **Network Traffic** - Bytes sent/received
- **System Info** - Hostname, OS, CPU cores, total RAM, total disk
- **Uptime** - How long the system has been running

## Support

For issues or questions:
- GitHub Issues: https://github.com/Constantin-E-T/lunasentri/issues
- Documentation: https://github.com/Constantin-E-T/lunasentri/tree/main/docs

---

**That's it!** Your server is now being monitored by LunaSentri. 🌙✨
