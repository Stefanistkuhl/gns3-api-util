#!/bin/bash

# GNS3 SSL Uninstall Script
# This script removes the SSL reverse proxy setup for GNS3

set -e

# Configuration variables (will be replaced by the Go code)
VERBOSE=""
REVERSE_PROXY_PORT=""
GNS3_PORT=""
DOMAIN=""
FIREWALL_ALLOW=""
FIREWALL_BLOCK=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root or with sudo
if [ "$EUID" -eq 0 ]; then
    SUDO=""
else
    SUDO="sudo"
    if ! command -v sudo &> /dev/null; then
        error "sudo is required but not installed"
        exit 1
    fi
fi

log "Starting GNS3 SSL uninstallation..."

# Stop and disable Caddy service
log "Stopping Caddy service..."
if systemctl is-active --quiet caddy; then
    $SUDO systemctl stop caddy
    log "Caddy service stopped"
else
    warn "Caddy service was not running"
fi

if systemctl is-enabled --quiet caddy; then
    $SUDO systemctl disable caddy
    log "Caddy service disabled"
else
    warn "Caddy service was not enabled"
fi

# Remove Caddy configuration
CADDY_CONFIG_DIR="/etc/caddy"
CADDY_CONFIG_FILE="$CADDY_CONFIG_DIR/Caddyfile"

if [ -f "$CADDY_CONFIG_FILE" ]; then
    log "Removing Caddy configuration..."
    $SUDO rm -f "$CADDY_CONFIG_FILE"
    log "Caddy configuration removed"
else
    warn "Caddy configuration file not found"
fi

# Remove Caddy from system
log "Removing Caddy..."
if command -v caddy &> /dev/null; then
    # Check if caddy package is installed
    if dpkg -l | grep -q '^ii.*caddy'; then
        log "Caddy package is installed, removing package..."
        $SUDO apt-get remove --purge -y caddy
        log "Caddy package removed"
    else
        # Package not installed, just remove binary
        $SUDO rm -f /usr/bin/caddy
        log "Caddy binary removed"
    fi
else
    warn "Caddy binary not found"
fi

# Remove Caddy systemd service
SERVICE_FILE="/etc/systemd/system/caddy.service"
if [ -f "$SERVICE_FILE" ]; then
    log "Removing Caddy systemd service..."
    $SUDO rm -f "$SERVICE_FILE"
    $SUDO systemctl daemon-reload
    log "Caddy systemd service removed"
else
    warn "Caddy systemd service file not found"
fi

# Remove Caddy user and group
if id "caddy" &>/dev/null; then
    log "Removing Caddy user and group..."
    $SUDO userdel caddy 2>/dev/null || true
    $SUDO groupdel caddy 2>/dev/null || true
    log "Caddy user and group removed"
else
    warn "Caddy user not found"
fi

# Remove Caddy directories
CADDY_DIRS=("/etc/caddy" "/var/lib/caddy" "/var/log/caddy")
for dir in "${CADDY_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        log "Removing directory: $dir"
        $SUDO rm -rf "$dir"
    fi
done

# Remove SSL certificates
CERT_DIR="/etc/ssl/certs/gns3"
if [ -d "$CERT_DIR" ]; then
    log "Removing SSL certificates..."
    $SUDO rm -rf "$CERT_DIR"
    log "SSL certificates removed"
else
    warn "SSL certificate directory not found"
fi

# Remove private key
KEY_DIR="/etc/ssl/private/gns3"
if [ -d "$KEY_DIR" ]; then
    log "Removing SSL private keys..."
    $SUDO rm -rf "$KEY_DIR"
    log "SSL private keys removed"
else
    warn "SSL private key directory not found"
fi

# Load installation state if available (check multiple locations)
STATE_FILE=""
for location in "/etc/gns3/ssl_state.json" "/var/lib/gns3/ssl_state.json"; do
    if [ -f "$location" ]; then
        STATE_FILE="$location"
        break
    fi
done

if [ -n "$STATE_FILE" ]; then
    log "Loading installation state from $STATE_FILE"
    FIREWALL_BLOCK=$(grep -o '"firewall_block": [^,]*' "$STATE_FILE" | cut -d' ' -f2)
    FIREWALL_ALLOW=$(grep -o '"firewall_allow": "[^"]*"' "$STATE_FILE" | cut -d'"' -f4)
    GNS3_PORT=$(grep -o '"gns3_port": [^,]*' "$STATE_FILE" | cut -d' ' -f2)
    log "Loaded state: FIREWALL_BLOCK=$FIREWALL_BLOCK, FIREWALL_ALLOW=$FIREWALL_ALLOW, GNS3_PORT=$GNS3_PORT"
else
    warn "No installation state found, using default values"
fi

# Remove firewall rules
if [ "$FIREWALL_BLOCK" = "true" ] || [ "$FIREWALL_ALLOW" != "" ]; then
    log "Removing firewall rules..."
    
    # Remove UFW rules for GNS3 port
    if command -v ufw &> /dev/null; then
        # Remove deny rule for GNS3 port
        $SUDO ufw --force delete deny $GNS3_PORT 2>/dev/null || true
        
        # Remove allow rule for specific subnet
        if [ "$FIREWALL_ALLOW" != "" ]; then
            $SUDO ufw --force delete allow from $FIREWALL_ALLOW to any port $GNS3_PORT 2>/dev/null || true
        fi
        
        log "Firewall rules removed"
    else
        warn "UFW not found, skipping firewall rule removal"
    fi
fi

# Remove reverse proxy port from GNS3 configuration (if it exists)
GNS3_CONFIG="/etc/gns3/gns3_server.conf"
if [ -f "$GNS3_CONFIG" ]; then
    log "Checking GNS3 configuration..."
    if grep -q "port.*$REVERSE_PROXY_PORT" "$GNS3_CONFIG"; then
        warn "Found reverse proxy port in GNS3 config. You may need to manually remove it."
        warn "File: $GNS3_CONFIG"
    fi
fi

# Clean up any remaining processes
log "Cleaning up any remaining Caddy processes..."
$SUDO pkill -f caddy 2>/dev/null || true

# Remove any remaining log files
LOG_FILES=("/var/log/caddy.log" "/var/log/caddy/access.log" "/var/log/caddy/error.log")
for log_file in "${LOG_FILES[@]}"; do
    if [ -f "$log_file" ]; then
        $SUDO rm -f "$log_file"
    fi
done

# Clean up state files
log "Cleaning up state files..."
for state_file in "/etc/gns3/ssl_state.json" "/var/lib/gns3/ssl_state.json"; do
    if [ -f "$state_file" ]; then
        $SUDO rm -f "$state_file"
        log "Removed state file: $state_file"
    fi
done

# Clean up /var/lib/gns3 directory if empty
if [ -d "/var/lib/gns3" ] && [ -z "$(ls -A /var/lib/gns3 2>/dev/null)" ]; then
    $SUDO rmdir /var/lib/gns3
    log "Removed empty directory: /var/lib/gns3"
fi

log "GNS3 SSL uninstallation completed successfully!"
log "The following services have been removed:"
log "  - Caddy reverse proxy"
log "  - SSL certificates"
log "  - Firewall rules (if any)"
log "  - Systemd service"
log "  - State files"
log ""
log "Note: GNS3 server itself was not removed, only the SSL reverse proxy setup."
log "Your GNS3 server should now be accessible on its original port: $GNS3_PORT"
