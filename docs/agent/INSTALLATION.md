# LunaSentri Agent Installation Guide

The LunaSentri agent is a lightweight monitoring daemon that collects system metrics and sends them to your LunaSentri server.

## Prerequisites

- Linux system (Ubuntu, Debian, CentOS, RHEL, or similar)
- Root/sudo access for installation
- Network connectivity to your LunaSentri server
- A machine API key (generated from the LunaSentri web interface)

## Quick Install (Linux)

### 1. Get Your API Key

Before installing the agent, you need to register a machine and get its API key:

1. Log in to your LunaSentri dashboard
2. Navigate to the **Machines** page
3. Click **"Add Machine"**
4. Enter a name and description for your machine
5. Copy the generated API key (you'll only see this once!)

### 2. Download and Run Installer

```bash
# Download the installer script
curl -fsSL https://raw.githubusercontent.com/Constantin-E-T/lunasentri/main/apps/agent/scripts/install.sh -o install.sh

# Make it executable
chmod +x install.sh

# Run the installer (requires sudo)
sudo ./install.sh
```

The installer will:

- Create a dedicated `lunasentri` system user
- Install the agent binary to `/usr/local/bin`
- Create a configuration file at `/etc/lunasentri/agent.yaml`
- Set up a systemd service
- Start the agent automatically

### 3. Verify Installation

Check that the agent is running:

```bash
sudo systemctl status lunasentri-agent
```

View live logs:

```bash
sudo journalctl -u lunasentri-agent -f
```

You should see JSON-formatted log entries showing metrics being collected and sent.

## Manual Installation

If you prefer to install manually or need more control:

### 1. Build the Binary

```bash
cd apps/agent
make build
```

Or build for Linux from macOS:

```bash
make build-linux
```

### 2. Install Binary

```bash
sudo cp dist/lunasentri-agent /usr/local/bin/
sudo chmod +x /usr/local/bin/lunasentri-agent
```

### 3. Create Configuration

Create `/etc/lunasentri/agent.yaml`:

```yaml
server_url: "https://api.lunasentri.com"
api_key: "your-machine-api-key-here"
interval: "10s"
system_info_period: "1h"
max_retries: 3
retry_backoff: "5s"
```

**Important:** Secure the config file as it contains your API key:

```bash
sudo chmod 600 /etc/lunasentri/agent.yaml
sudo chown root:root /etc/lunasentri/agent.yaml
```

### 4. Create System User

```bash
sudo useradd --system --no-create-home --shell /bin/false lunasentri
```

### 5. Create Systemd Service

Create `/etc/systemd/system/lunasentri-agent.service`:

```ini
[Unit]
Description=LunaSentri Monitoring Agent
Documentation=https://github.com/Constantin-E-T/lunasentri
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=lunasentri
Group=lunasentri
ExecStart=/usr/local/bin/lunasentri-agent --config /etc/lunasentri/agent.yaml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadOnlyPaths=/
ReadWritePaths=/etc/lunasentri

[Install]
WantedBy=multi-user.target
```

### 6. Start the Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable lunasentri-agent
sudo systemctl start lunasentri-agent
```

## Docker Installation

You can also run the agent in a Docker container:

### 1. Build the Image

```bash
cd apps/agent
make docker-build
```

### 2. Run the Container

```bash
docker run -d \
  --name lunasentri-agent \
  --restart unless-stopped \
  -e LUNASENTRI_SERVER_URL=https://api.lunasentri.com \
  -e LUNASENTRI_API_KEY=your-api-key-here \
  lunasentri/agent:latest
```

### Docker Compose

Create a `docker-compose.yml`:

```yaml
version: '3.8'

services:
  lunasentri-agent:
    image: lunasentri/agent:latest
    container_name: lunasentri-agent
    restart: unless-stopped
    environment:
      - LUNASENTRI_SERVER_URL=https://api.lunasentri.com
      - LUNASENTRI_API_KEY=your-api-key-here
      - LUNASENTRI_INTERVAL=10s
```

Then run:

```bash
docker-compose up -d
```

## Configuration Options

### Command-Line Flags

```bash
lunasentri-agent [flags]

Flags:
  --server-url string         LunaSentri server URL
  --api-key string           Machine API key (required)
  --interval duration        Metrics collection interval (default 10s)
  --system-info-period duration  System info update period (default 1h)
  --max-retries int          Maximum retry attempts (default 3)
  --retry-backoff duration   Retry backoff duration (default 5s)
  --config string            Path to configuration file
```

### Environment Variables

- `LUNASENTRI_SERVER_URL` - Server URL
- `LUNASENTRI_API_KEY` - Machine API key
- `LUNASENTRI_INTERVAL` - Collection interval (e.g., "10s", "1m")
- `LUNASENTRI_SYSTEM_INFO_PERIOD` - System info refresh period
- `LUNASENTRI_MAX_RETRIES` - Maximum retry attempts
- `LUNASENTRI_RETRY_BACKOFF` - Retry backoff duration

### Configuration File

Default locations (checked in order):

1. Path specified by `--config` flag
2. `/etc/lunasentri/agent.yaml` (system-wide)
3. `~/.config/lunasentri/agent.yaml` (user-specific)

Example `agent.yaml`:

```yaml
server_url: "https://api.lunasentri.com"
api_key: "your-api-key"
interval: "10s"
system_info_period: "1h"
max_retries: 3
retry_backoff: "5s"
```

### Configuration Precedence

Settings are loaded in this order (later values override earlier ones):

1. Default values
2. Configuration file
3. Environment variables
4. Command-line flags

## Metrics Collected

The agent collects and reports:

- **CPU Usage** - Percentage of CPU utilization
- **Memory Usage** - Percentage of RAM used
- **Disk Usage** - Percentage of root filesystem used
- **Network I/O** - Cumulative bytes received/sent
- **Uptime** - System uptime in seconds

### System Information (sent periodically)

- Hostname
- Platform (OS name and version)
- Kernel version
- CPU cores
- Total memory (MB)
- Total disk space (GB)
- Last boot time

## Management Commands

### Check Status

```bash
sudo systemctl status lunasentri-agent
```

### View Logs

```bash
# Follow live logs
sudo journalctl -u lunasentri-agent -f

# Last 100 lines
sudo journalctl -u lunasentri-agent -n 100

# Logs from last hour
sudo journalctl -u lunasentri-agent --since "1 hour ago"
```

### Restart Agent

```bash
sudo systemctl restart lunasentri-agent
```

### Stop Agent

```bash
sudo systemctl stop lunasentri-agent
```

### Reload Configuration

```bash
# Edit config file
sudo nano /etc/lunasentri/agent.yaml

# Restart to apply changes
sudo systemctl restart lunasentri-agent
```

## Uninstallation

To completely remove the agent:

```bash
# Stop and disable service
sudo systemctl stop lunasentri-agent
sudo systemctl disable lunasentri-agent

# Remove files
sudo rm /etc/systemd/system/lunasentri-agent.service
sudo rm /usr/local/bin/lunasentri-agent
sudo rm -rf /etc/lunasentri

# Remove user
sudo userdel lunasentri

# Reload systemd
sudo systemctl daemon-reload
```

## Troubleshooting

### Agent Not Sending Metrics

1. **Check agent status:**

   ```bash
   sudo systemctl status lunasentri-agent
   ```

2. **View logs for errors:**

   ```bash
   sudo journalctl -u lunasentri-agent -n 50
   ```

3. **Verify API key:**
   - Ensure the API key is correct in `/etc/lunasentri/agent.yaml`
   - Check that the machine is registered in the web interface

4. **Test network connectivity:**

   ```bash
   curl -I https://api.lunasentri.com/health
   ```

### High CPU/Memory Usage

The agent is designed to be lightweight. If you notice high resource usage:

1. Increase the collection interval:

   ```yaml
   interval: "30s"  # or "1m"
   ```

2. Check for error loops in logs

### Metrics Not Appearing in Dashboard

1. **Verify machine is registered:**
   - Check the Machines page in the web interface
   - Ensure the machine shows as "Online"

2. **Check API key:**
   - The API key must match what was generated during registration

3. **Verify server URL:**
   - Ensure `server_url` points to the correct LunaSentri API

### Permission Errors

If you see permission errors:

```bash
# Ensure agent runs as correct user
sudo systemctl edit lunasentri-agent

# Add:
[Service]
User=lunasentri
Group=lunasentri
```

## Security Considerations

- The agent runs as a dedicated non-root user (`lunasentri`)
- The API key is stored in `/etc/lunasentri/agent.yaml` with 600 permissions
- All communication with the server uses HTTPS
- The systemd service has security hardening enabled (see service file)
- No privileged operations are performed during metrics collection

## Support

For issues or questions:

- GitHub Issues: <https://github.com/Constantin-E-T/lunasentri/issues>
- Documentation: <https://github.com/Constantin-E-T/lunasentri/docs>
