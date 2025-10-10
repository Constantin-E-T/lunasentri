#!/bin/bash
set -e

# LunaSentri Agent Installation Script
# This script installs the LunaSentri monitoring agent on Linux systems

VERSION="1.0.0"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/lunasentri"
SERVICE_FILE="/etc/systemd/system/lunasentri-agent.service"
AGENT_USER="lunasentri"
AGENT_BINARY="lunasentri-agent"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print functions
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then 
        print_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Detect Linux distribution
detect_distro() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        DISTRO=$ID
        print_info "Detected Linux distribution: $NAME"
    else
        print_warn "Unable to detect distribution, proceeding with generic install"
        DISTRO="unknown"
    fi
}

# Create lunasentri user if it doesn't exist
create_user() {
    if id "$AGENT_USER" &>/dev/null; then
        print_info "User '$AGENT_USER' already exists"
    else
        print_info "Creating system user '$AGENT_USER'..."
        useradd --system --no-create-home --shell /bin/false "$AGENT_USER"
        print_info "User '$AGENT_USER' created"
    fi
}

# Download or copy agent binary
install_binary() {
    print_info "Installing agent binary..."
    
    # For MVP, we expect the binary to be in the current directory or dist/
    if [ -f "./lunasentri-agent" ]; then
        cp "./lunasentri-agent" "$INSTALL_DIR/$AGENT_BINARY"
    elif [ -f "./dist/lunasentri-agent" ]; then
        cp "./dist/lunasentri-agent" "$INSTALL_DIR/$AGENT_BINARY"
    else
        print_error "Agent binary not found. Please build the binary first:"
        print_error "  cd apps/agent && go build -o lunasentri-agent"
        exit 1
    fi
    
    chmod +x "$INSTALL_DIR/$AGENT_BINARY"
    print_info "Binary installed to $INSTALL_DIR/$AGENT_BINARY"
}

# Create configuration directory and file
setup_config() {
    print_info "Setting up configuration..."
    
    # Create config directory
    mkdir -p "$CONFIG_DIR"
    
    # Prompt for configuration values
    read -p "Enter LunaSentri server URL [https://api.lunasentri.com]: " SERVER_URL
    SERVER_URL=${SERVER_URL:-https://api.lunasentri.com}
    
    while [ -z "$API_KEY" ]; do
        read -p "Enter your machine API key (required): " API_KEY
        if [ -z "$API_KEY" ]; then
            print_warn "API key is required"
        fi
    done
    
    read -p "Enter metrics collection interval [10s]: " INTERVAL
    INTERVAL=${INTERVAL:-10s}
    
    # Create config file
    cat > "$CONFIG_DIR/agent.yaml" <<EOF
# LunaSentri Agent Configuration
# Generated on $(date)

# Server URL for the LunaSentri API
server_url: "$SERVER_URL"

# Machine API key (keep this secret!)
api_key: "$API_KEY"

# Metrics collection interval
interval: "$INTERVAL"

# System info update period
system_info_period: "1h"

# Maximum retry attempts for failed API calls
max_retries: 3

# Retry backoff duration
retry_backoff: "5s"
EOF
    
    chmod 600 "$CONFIG_DIR/agent.yaml"
    chown root:root "$CONFIG_DIR/agent.yaml"
    
    print_info "Configuration file created at $CONFIG_DIR/agent.yaml"
}

# Create systemd service
setup_systemd() {
    print_info "Setting up systemd service..."
    
    cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=LunaSentri Monitoring Agent
Documentation=https://github.com/Constantin-E-T/lunasentri
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=$AGENT_USER
Group=$AGENT_USER
ExecStart=$INSTALL_DIR/$AGENT_BINARY --config $CONFIG_DIR/agent.yaml
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
ReadWritePaths=$CONFIG_DIR

[Install]
WantedBy=multi-user.target
EOF
    
    # Reload systemd
    systemctl daemon-reload
    print_info "Systemd service created"
}

# Start and enable service
start_service() {
    print_info "Starting LunaSentri agent service..."
    
    systemctl enable lunasentri-agent.service
    systemctl start lunasentri-agent.service
    
    sleep 2
    
    # Check status
    if systemctl is-active --quiet lunasentri-agent.service; then
        print_info "✓ LunaSentri agent is running!"
    else
        print_error "Failed to start agent. Check logs with: journalctl -u lunasentri-agent.service"
        exit 1
    fi
}

# Print post-installation instructions
print_instructions() {
    echo ""
    echo "========================================"
    echo "  LunaSentri Agent Installation Complete"
    echo "========================================"
    echo ""
    print_info "The agent is now running and sending metrics to: $SERVER_URL"
    echo ""
    echo "Useful commands:"
    echo "  - Check status:  sudo systemctl status lunasentri-agent"
    echo "  - View logs:     sudo journalctl -u lunasentri-agent -f"
    echo "  - Restart:       sudo systemctl restart lunasentri-agent"
    echo "  - Stop:          sudo systemctl stop lunasentri-agent"
    echo ""
    echo "Configuration file: $CONFIG_DIR/agent.yaml"
    echo ""
    echo "To uninstall, run:"
    echo "  sudo systemctl stop lunasentri-agent"
    echo "  sudo systemctl disable lunasentri-agent"
    echo "  sudo rm $SERVICE_FILE"
    echo "  sudo rm $INSTALL_DIR/$AGENT_BINARY"
    echo "  sudo rm -rf $CONFIG_DIR"
    echo "  sudo userdel $AGENT_USER"
    echo ""
}

# Main installation flow
main() {
    echo "LunaSentri Agent Installer v${VERSION}"
    echo ""
    
    check_root
    detect_distro
    create_user
    install_binary
    setup_config
    setup_systemd
    start_service
    print_instructions
}

main
