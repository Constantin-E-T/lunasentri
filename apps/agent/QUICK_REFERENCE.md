# Quick Reference - LunaSentri Agent

## Development

```bash
# Build
cd apps/agent
make build                  # Build for current platform
make build-linux            # Cross-compile for Linux (from macOS)

# Test
make test                   # Run all tests
go test -v ./...           # Verbose test output
go test -cover ./...       # With coverage

# Run locally
export LUNASENTRI_SERVER_URL=http://localhost:8080
export LUNASENTRI_API_KEY=your-test-key
go run main.go

# Or with flags
go run main.go --server-url=http://localhost:8080 --api-key=test-key --interval=5s
```

## Docker

```bash
# Build image
make docker-build

# Run
docker run --rm \
  -e LUNASENTRI_SERVER_URL=http://host.docker.internal:8080 \
  -e LUNASENTRI_API_KEY=your-key \
  lunasentri/agent:latest

# Or using make
LUNASENTRI_API_KEY=your-key make docker-run
```

## Installation (Linux)

```bash
# Quick install
curl -fsSL https://raw.githubusercontent.com/Constantin-E-T/lunasentri/main/apps/agent/scripts/install.sh | sudo bash

# Manual install
sudo cp dist/lunasentri-agent /usr/local/bin/
sudo mkdir -p /etc/lunasentri
sudo vim /etc/lunasentri/agent.yaml
# ... set up systemd service ...
```

## Management Commands

```bash
# Status
sudo systemctl status lunasentri-agent

# Logs
sudo journalctl -u lunasentri-agent -f              # Follow logs
sudo journalctl -u lunasentri-agent -n 100          # Last 100 lines
sudo journalctl -u lunasentri-agent --since "1h ago"  # Last hour

# Control
sudo systemctl start lunasentri-agent
sudo systemctl stop lunasentri-agent
sudo systemctl restart lunasentri-agent

# Edit config
sudo vim /etc/lunasentri/agent.yaml
sudo systemctl restart lunasentri-agent
```

## Configuration

### File: `/etc/lunasentri/agent.yaml`

```yaml
server_url: "https://api.lunasentri.com"
api_key: "your-api-key-here"
interval: "10s"
system_info_period: "1h"
max_retries: 3
retry_backoff: "5s"
```

### Environment Variables

```bash
LUNASENTRI_SERVER_URL=https://api.lunasentri.com
LUNASENTRI_API_KEY=your-key
LUNASENTRI_INTERVAL=10s
LUNASENTRI_SYSTEM_INFO_PERIOD=1h
LUNASENTRI_MAX_RETRIES=3
LUNASENTRI_RETRY_BACKOFF=5s
```

### Command-Line Flags

```bash
lunasentri-agent \
  --server-url=https://api.lunasentri.com \
  --api-key=your-key \
  --interval=10s \
  --system-info-period=1h \
  --max-retries=3 \
  --retry-backoff=5s
```

## Troubleshooting

### Check if agent is sending metrics

```bash
# Watch logs
sudo journalctl -u lunasentri-agent -f

# Look for these patterns:
# ✅ Good: {"level":"info","msg":"Metrics sent successfully",...}
# ❌ Bad:  {"level":"error","msg":"Failed to send metrics",...}
```

### Test connectivity

```bash
# Test API endpoint
curl -I https://api.lunasentri.com/health

# Test with API key
curl -X POST https://api.lunasentri.com/agent/metrics \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"cpu_pct":50,"mem_used_pct":60,"disk_used_pct":70}'
```

### Debug mode

```bash
# Run in foreground with verbose output
sudo -u lunasentri /usr/local/bin/lunasentri-agent \
  --config /etc/lunasentri/agent.yaml
```

## Project Structure

```
apps/agent/
├── main.go                    # Entry point
├── main_test.go              # Integration tests
├── Makefile                  # Build automation
├── Dockerfile                # Docker image
├── README.md                 # Documentation
├── internal/
│   ├── config/              # Configuration loading
│   │   ├── config.go
│   │   └── config_test.go
│   ├── collector/           # Metrics collection
│   │   └── collector.go
│   └── transport/           # API client
│       └── client.go
└── scripts/
    └── install.sh           # Linux installer
```

## API Endpoints

### Register Machine (Web UI only)

```
POST /agent/register
Authorization: Bearer <session_token>

Request:
{
  "name": "production-server",
  "hostname": "web-01.example.com",
  "description": "Main web server"
}

Response:
{
  "id": 1,
  "name": "production-server",
  "api_key": "lunasentri_abcd1234...",  # Only shown once!
  "created_at": "2025-10-10T12:00:00Z"
}
```

### Send Metrics (Agent)

```
POST /agent/metrics
Authorization: Bearer <api_key>

Request:
{
  "cpu_pct": 45.5,
  "mem_used_pct": 67.8,
  "disk_used_pct": 23.4,
  "net_rx_bytes": 1024000,
  "net_tx_bytes": 512000,
  "uptime_s": 12345.0,
  "system_info": {
    "hostname": "web-01",
    "platform": "ubuntu",
    "cpu_cores": 4,
    "memory_total_mb": 8192
  }
}

Response: 202 Accepted
```

## Common Issues

| Problem | Solution |
|---------|----------|
| "Unauthorized" errors | Check API key is correct in config |
| Metrics not appearing | Verify machine is registered in web UI |
| High CPU usage | Increase `interval` to reduce collection frequency |
| Connection refused | Check `server_url` points to correct API endpoint |
| Permission denied | Ensure agent runs as `lunasentri` user, not root |

## Performance

- **Memory**: ~10-20 MB RSS
- **CPU**: <0.1% (idle), ~0.5% during collection
- **Network**: ~500 bytes per metric submission
- **Disk**: Binary is ~10 MB

## Security Checklist

- [ ] API key stored with 600 permissions
- [ ] Agent runs as non-root user (`lunasentri`)
- [ ] Config file owned by root:root
- [ ] HTTPS used for all communication
- [ ] Systemd security hardening enabled
- [ ] No privileged operations performed

## Release Checklist

- [ ] Tests passing (`make test`)
- [ ] Binary builds (`make build`)
- [ ] Docker image builds (`make docker-build`)
- [ ] Install script works
- [ ] Documentation updated
- [ ] CHANGELOG updated
- [ ] Version tagged in git
