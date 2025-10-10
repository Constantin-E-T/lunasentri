# LunaSentri Agent

A lightweight monitoring agent for collecting and streaming system metrics to the LunaSentri platform.

## Features

- **Lightweight** - Minimal resource footprint
- **Secure** - Runs as non-root user, API key authentication
- **Reliable** - Automatic retry with exponential backoff
- **Cross-platform** - Written in Go, works on Linux (primary), macOS, and Windows
- **Easy to install** - One-command installation script
- **Docker support** - Run in containers
- **Structured logging** - JSON output for easy parsing

## Quick Start

### Installation

```bash
# Download and run installer (requires sudo)
curl -fsSL https://raw.githubusercontent.com/Constantin-E-T/lunasentri/main/apps/agent/scripts/install.sh | sudo bash
```

See [INSTALLATION.md](../../docs/agent/INSTALLATION.md) for detailed instructions.

### Configuration

Edit `/etc/lunasentri/agent.yaml`:

```yaml
server_url: "https://api.lunasentri.com"
api_key: "your-machine-api-key"
interval: "10s"
```

### Usage

```bash
# Check status
sudo systemctl status lunasentri-agent

# View logs
sudo journalctl -u lunasentri-agent -f

# Restart
sudo systemctl restart lunasentri-agent
```

## Development

### Prerequisites

- Go 1.24 or later
- Make (optional)

### Building

```bash
# Build for current platform
make build

# Build for Linux (from macOS)
make build-linux

# Run tests
make test

# Build Docker image
make docker-build
```

### Running Locally

```bash
# Set environment variables
export LUNASENTRI_SERVER_URL=http://localhost:8080
export LUNASENTRI_API_KEY=your-test-api-key

# Run the agent
go run main.go
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestMetricsPayloadShape
```

## Project Structure

```
apps/agent/
├── main.go                      # Entry point
├── go.mod                       # Go module definition
├── Makefile                     # Build automation
├── Dockerfile                   # Docker image definition
├── internal/
│   ├── config/                  # Configuration loading
│   │   ├── config.go
│   │   └── config_test.go
│   ├── collector/               # Metrics collection
│   │   └── collector.go
│   └── transport/               # API communication
│       └── client.go
└── scripts/
    └── install.sh               # Linux installation script
```

## Architecture

The agent consists of three main components:

### 1. Configuration (`internal/config`)

Handles loading configuration from multiple sources with precedence:

1. Command-line flags (highest)
2. Environment variables
3. Configuration file
4. Default values (lowest)

### 2. Collector (`internal/collector`)

Collects system metrics using the `gopsutil` library:

- CPU usage percentage
- Memory usage percentage
- Disk usage percentage
- Network I/O counters
- System uptime
- System information (hostname, platform, hardware details)

### 3. Transport (`internal/transport`)

Handles communication with the LunaSentri API:

- HTTP client with configurable timeout
- Automatic retry with exponential backoff
- Structured JSON logging
- API key authentication

## Metrics Collected

| Metric | Type | Description |
|--------|------|-------------|
| `cpu_pct` | float64 | CPU usage percentage (0-100) |
| `mem_used_pct` | float64 | Memory usage percentage (0-100) |
| `disk_used_pct` | float64 | Root filesystem usage percentage (0-100) |
| `net_rx_bytes` | int64 | Cumulative network bytes received |
| `net_tx_bytes` | int64 | Cumulative network bytes sent |
| `uptime_s` | float64 | System uptime in seconds |

## Configuration Options

### Command-Line Flags

- `--server-url` - LunaSentri server URL
- `--api-key` - Machine API key (required)
- `--interval` - Metrics collection interval (default: 10s)
- `--system-info-period` - System info update period (default: 1h)
- `--max-retries` - Maximum retry attempts (default: 3)
- `--retry-backoff` - Retry backoff duration (default: 5s)
- `--config` - Path to configuration file

### Environment Variables

- `LUNASENTRI_SERVER_URL`
- `LUNASENTRI_API_KEY`
- `LUNASENTRI_INTERVAL`
- `LUNASENTRI_SYSTEM_INFO_PERIOD`
- `LUNASENTRI_MAX_RETRIES`
- `LUNASENTRI_RETRY_BACKOFF`

## Docker Usage

### Build Image

```bash
docker build -t lunasentri/agent:latest .
```

### Run Container

```bash
docker run -d \
  --name lunasentri-agent \
  --restart unless-stopped \
  -e LUNASENTRI_SERVER_URL=https://api.lunasentri.com \
  -e LUNASENTRI_API_KEY=your-api-key \
  lunasentri/agent:latest
```

### Docker Compose

```yaml
version: '3.8'
services:
  lunasentri-agent:
    image: lunasentri/agent:latest
    restart: unless-stopped
    environment:
      LUNASENTRI_SERVER_URL: https://api.lunasentri.com
      LUNASENTRI_API_KEY: ${LUNASENTRI_API_KEY}
      LUNASENTRI_INTERVAL: 10s
```

## Security

- Runs as dedicated non-root user (`lunasentri`)
- API key stored with restricted permissions (600)
- HTTPS-only communication
- No privileged operations during metrics collection
- Systemd service includes security hardening:
  - `NoNewPrivileges=true`
  - `PrivateTmp=true`
  - `ProtectSystem=strict`
  - `ProtectHome=true`

## License

This project is part of the LunaSentri monitoring platform.

## Documentation

- [Installation Guide](../../docs/agent/INSTALLATION.md)
- [Main Documentation](../../docs/README.md)
- [API Documentation](../../docs/deployment/DEPLOYMENT.md)
