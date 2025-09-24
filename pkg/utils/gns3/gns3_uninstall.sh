#!/bin/bash

# GNS3 Uninstall Script
# This script removes GNS3 server installation

# Don't exit on errors, handle them gracefully
set -o pipefail

# Configuration variables (will be replaced by the Go code)
GNS3_USER=""
GNS3_HOME=""
GNS3_PORT=""
VERBOSE=""
PRESERVE_DATA=""

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

log "Starting GNS3 server uninstallation..."

# Check if GNS3 is actually installed
GNS3_INSTALLED=false

# Check for GNS3 service
if systemctl list-unit-files | grep -q "gns3.service" 2>/dev/null; then
    GNS3_INSTALLED=true
fi

# Check for GNS3 packages
if command -v apt-get &> /dev/null; then
    if dpkg -l | grep -q "gns3-server" 2>/dev/null; then
        GNS3_INSTALLED=true
    fi
elif command -v yum &> /dev/null; then
    if yum list installed | grep -q "gns3-server" 2>/dev/null; then
        GNS3_INSTALLED=true
    fi
elif command -v pacman &> /dev/null; then
    if pacman -Q | grep -q "gns3-server" 2>/dev/null; then
        GNS3_INSTALLED=true
    fi
fi

# Check for GNS3 user
if id "$GNS3_USER" &>/dev/null 2>/dev/null; then
    GNS3_INSTALLED=true
fi

# Check for GNS3 home directory
if [ -d "$GNS3_HOME" ]; then
    GNS3_INSTALLED=true
fi

# Check for GNS3 config directory
if [ -d "/etc/gns3" ]; then
    GNS3_INSTALLED=true
fi

if [ "$GNS3_INSTALLED" = false ]; then
    log "GNS3 server does not appear to be installed on this system."
    log "No GNS3 components found to remove."
    log "Uninstallation completed - nothing to do."
    exit 0
fi

log "GNS3 installation detected. Proceeding with uninstallation..."

# Stop and disable GNS3 service
log "Stopping GNS3 service..."
if systemctl is-active --quiet gns3 2>/dev/null; then
    $SUDO systemctl stop gns3
    log "GNS3 service stopped"
else
    warn "GNS3 service was not running"
fi

if systemctl is-enabled --quiet gns3 2>/dev/null; then
    $SUDO systemctl disable gns3
    log "GNS3 service disabled"
else
    warn "GNS3 service was not enabled"
fi

# Remove GNS3 systemd service
SERVICE_FILE="/lib/systemd/system/gns3.service"
if [ -f "$SERVICE_FILE" ]; then
    log "Removing GNS3 systemd service..."
    $SUDO rm -f "$SERVICE_FILE"
    $SUDO systemctl daemon-reload
    log "GNS3 systemd service removed"
else
    warn "GNS3 systemd service file not found"
fi

# Remove GNS3 configuration
GNS3_CONFIG_DIR="/etc/gns3"
if [ -d "$GNS3_CONFIG_DIR" ]; then
    log "Removing GNS3 configuration..."
    $SUDO rm -rf "$GNS3_CONFIG_DIR"
    log "GNS3 configuration removed"
else
    warn "GNS3 configuration directory not found"
fi

# Remove GNS3 user home directory (if specified and exists)
if [ "$PRESERVE_DATA" = "True" ]; then
    log "Preserving GNS3 home directory: $GNS3_HOME (--preserve-data flag used)"
elif [ -n "$GNS3_HOME" ] && [ -d "$GNS3_HOME" ]; then
    log "Removing GNS3 home directory: $GNS3_HOME"
    $SUDO rm -rf "$GNS3_HOME"
    log "GNS3 home directory removed"
else
    warn "GNS3 home directory not found or not specified"
fi

# Note: GNS3 user is preserved to avoid breaking file ownership
# The user can be manually removed later if desired with: sudo userdel gns3

# Remove GNS3 log and run directories
GNS3_DIRS=("/var/log/gns3" "/var/run/gns3")
for dir in "${GNS3_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        log "Removing directory: $dir"
        $SUDO rm -rf "$dir"
    fi
done

# Remove GNS3 packages
log "Removing GNS3 packages..."
if command -v apt-get &> /dev/null; then
    # Ubuntu/Debian
    PACKAGES_TO_REMOVE=(
        "gns3-server"
        "gns3-iou"
    )
    
    for package in "${PACKAGES_TO_REMOVE[@]}"; do
        if dpkg -l | grep -q "^ii.*$package"; then
            log "Removing package: $package"
            $SUDO apt-get remove --purge -y "$package" 2>/dev/null || true
        fi
    done
    
    # Remove GNS3 PPA
    if [ -f "/etc/apt/sources.list.d/gns3-ubuntu-ppa-v3-*.list" ]; then
        log "Removing GNS3 PPA..."
        $SUDO add-apt-repository --remove -y "ppa:gns3/ppa-v3" 2>/dev/null || true
        $SUDO apt-get update -qq 2>/dev/null || true
    fi
elif command -v yum &> /dev/null; then
    # CentOS/RHEL/Fedora
    log "Removing GNS3 packages (yum/dnf)..."
    $SUDO yum remove -y gns3-server 2>/dev/null || true
elif command -v pacman &> /dev/null; then
    # Arch Linux
    log "Removing GNS3 packages (pacman)..."
    $SUDO pacman -R --noconfirm gns3-server 2>/dev/null || true
fi

# Clean up any remaining processes
log "Cleaning up any remaining GNS3 processes..."
$SUDO pkill -f gns3server 2>/dev/null || true
$SUDO pkill -f gns3 2>/dev/null || true

# Remove any remaining log files
LOG_FILES=("/var/log/gns3.log" "/var/log/gns3/gns3.log")
for log_file in "${LOG_FILES[@]}"; do
    if [ -f "$log_file" ]; then
        $SUDO rm -f "$log_file"
    fi
done

# Show summary of what was actually removed
echo ""
log "GNS3 server uninstallation completed successfully!"
log "Summary of components processed:"

# Track what was actually removed
REMOVED_ITEMS=()

# Check what was actually removed
if systemctl list-unit-files | grep -q "gns3.service" 2>/dev/null; then
    warn "  - GNS3 systemd service: Still present (may need manual removal)"
else
    REMOVED_ITEMS+=("GNS3 systemd service")
fi

if [ -d "/etc/gns3" ]; then
    warn "  - GNS3 configuration: Still present"
else
    REMOVED_ITEMS+=("GNS3 configuration files")
fi

if [ -d "$GNS3_HOME" ]; then
    if [ "$PRESERVE_DATA" = "True" ]; then
        log "  ✓ GNS3 home directory ($GNS3_HOME): Preserved as requested"
    else
        warn "  - GNS3 home directory ($GNS3_HOME): Still present"
    fi
else
    REMOVED_ITEMS+=("GNS3 home directory")
fi

# Note: GNS3 user is intentionally preserved

# Show what was successfully removed
if [ ${#REMOVED_ITEMS[@]} -gt 0 ]; then
    log "Successfully removed:"
    for item in "${REMOVED_ITEMS[@]}"; do
        log "  ✓ $item"
    done
else
    log "No components were removed (they may not have been present)"
fi
log ""

warn "Note: The following components were NOT removed and may need manual cleanup:"
warn "  - QEMU/KVM packages (qemu-system-x86, qemu-utils)"
warn "  - Virtualization packages (libvirt-daemon-system, virtinst)"
warn "  - Docker (if installed for GNS3)"
warn "  - VirtualBox (if installed for GNS3)"
warn "  - VMware integration packages"
warn "  - Network tools (bridge-utils, ubridge)"
warn ""
warn "These components may be used by other applications."
warn "Remove them manually if you're sure they're not needed:"
warn "  sudo apt-get remove --purge qemu-system-x86 qemu-utils libvirt-daemon-system"
warn "  sudo apt-get remove --purge bridge-utils ubridge"
warn "  sudo apt-get autoremove"

if [ -n "$GNS3_PORT" ]; then
    log "GNS3 was running on port: $GNS3_PORT"
fi

# Exit successfully
exit 0
