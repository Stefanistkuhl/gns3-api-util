#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# CONFIG (edit as needed)
# ============================================================
GNS3_USER="gns3"               # Service user
GNS3_HOME="/opt/gns3"          # Home/data root
GNS3_PORT=3080                 # gns3server API port
GNS3_LISTEN_HOST="0.0.0.0"     # Bind address
DISABLE_KVM=0                  # 1 to disable KVM in GNS3 config

# Optional integrations (set to 1 to enable)
INSTALL_DOCKER=0               # Docker used by many GNS3 appliances
INSTALL_VIRTUALBOX=0           # VirtualBox support for nodes
INSTALL_VMWARE=0               # VMware Workstation/Player support for nodes

# Optional IOU support (requires valid license; no hostname changes are made here)
USE_IOU=0                      # Install gns3-iou
ENABLE_I386_FOR_IOU=0          # Add i386 arch for legacy 32-bit IOU
# ============================================================

# Require root or passwordless sudo
if [ "$EUID" -ne 0 ] && ! sudo -n true &>/dev/null; then
  echo "Error: must be root or have passwordless sudo" >&2
  exit 1
fi
SUDO=""
if [ "$EUID" -ne 0 ]; then
  SUDO="sudo"
fi

log() { echo "=> $*" >&2; }

# Ubuntu-only guard and LTS check
if ! command -v lsb_release >/dev/null 2>&1; then
  $SUDO apt-get update -qq >/dev/null 2>&1 || true
  $SUDO apt-get install -y -qq lsb-release >/dev/null 2>&1 || true
fi
if [ "$(lsb_release -is 2>/dev/null || echo "Unknown")" != "Ubuntu" ]; then
  echo "Error: This script only supports Ubuntu." >&2
  exit 1
fi
if ! lsb_release -d | grep -q "LTS"; then
  echo "Error: This script requires an Ubuntu LTS release." >&2
  exit 1
fi

# Python 3.9+ required for ppa-v3 (GNS3 v3+)
if ! python3 - <<'PY' >/dev/null 2>&1
import sys; exit(0 if sys.version_info >= (3,9) else 1)
PY
then
  echo "Error: GNS3 v3+ requires Python >= 3.9" >&2
  exit 1
fi

export DEBIAN_FRONTEND=noninteractive

log "Updating apt index and installing prerequisites"
$SUDO apt-get update -qq >/dev/null
$SUDO apt-get install -y -qq curl ca-certificates software-properties-common apt-transport-https gnupg lsb-release >/dev/null

log "Upgrading system packages (safe defaults)"
$SUDO apt-get -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" -y upgrade >/dev/null

log "Adding GNS3 repository ppa:gns3/ppa-v3"
$SUDO add-apt-repository -y "ppa:gns3/ppa-v3" >/dev/null
$SUDO apt-get update -qq >/dev/null

# Base packages for GNS3 server
BASE_PKGS=(
  gns3-server
  ubridge
  qemu-system-x86
  qemu-utils
  libvirt-daemon-system
  virtinst
  bridge-utils
  # helpful tools
  net-tools
  iproute2
  iptables
  socat
  unzip
  jq
)

log "Installing base GNS3 packages"
$SUDO apt-get install -y "${BASE_PKGS[@]}"

# Create gns3 user and home
if ! id -u "$GNS3_USER" >/dev/null 2>&1; then
  log "Creating user $GNS3_USER with home $GNS3_HOME"
  $SUDO useradd -m -d "$GNS3_HOME" "$GNS3_USER"
else
  [ -d "$GNS3_HOME" ] || $SUDO mkdir -p "$GNS3_HOME"
fi

# Group memberships: ubridge, kvm, libvirt
log "Adding $GNS3_USER to groups: ubridge, kvm, libvirt"
$SUDO usermod -aG ubridge "$GNS3_USER" || true
$SUDO usermod -aG kvm "$GNS3_USER" || true
$SUDO usermod -aG libvirt "$GNS3_USER" || true

# Optional: Docker
if [ "$INSTALL_DOCKER" -eq 1 ]; then
  if ! command -v docker >/dev/null 2>&1; then
    log "Installing Docker using official convenience script"
    curl -fsSL https://get.docker.com | $SUDO bash
  else
    log "Docker already installed: $(docker --version 2>/dev/null || echo present)"
  fi
  $SUDO usermod -aG docker "$GNS3_USER" || true
fi

# Optional: VirtualBox
if [ "$INSTALL_VIRTUALBOX" -eq 1 ]; then
  log "Installing VirtualBox"
  # Use Ubuntu repo VirtualBox for compatibility with headers/driver
  $SUDO apt-get install -y virtualbox
  # Some systems may require kernel headers for vboxdrv to build
  if command -v uname >/dev/null 2>&1; then
    KVER=$(uname -r)
    $SUDO apt-get install -y "linux-headers-${KVER}" || true
  fi
  # Add user to vboxusers group
  $SUDO usermod -aG vboxusers "$GNS3_USER" || true
fi

# Optional: VMware Workstation/Player integration
# Note: This installs the GNS3 integration packages. VMware itself must be installed separately by the admin.
if [ "$INSTALL_VMWARE" -eq 1 ]; then
  log "Installing VMware integration packages for GNS3"
  # gns3 has vmware integration helpers in the PPA; also install open-vm-tools for convenience
  $SUDO apt-get install -y open-vm-tools open-vm-tools-desktop || true
  # The GNS3 VMware integration is typically inside gns3-gui for desktop, but server uses remote VMs via vmrun.
  # Install vmrun if VMware Workstation/Player is installed; we do not fetch proprietary VMware here.
  # You can later place vmrun in PATH and GNS3 will detect it.
fi

# Optional IOU support
if [ "$USE_IOU" -eq 1 ]; then
  log "Enabling IOU support"
  if [ "$ENABLE_I386_FOR_IOU" -eq 1 ]; then
    log "Enabling i386 architecture for legacy IOU"
    $SUDO dpkg --add-architecture i386
    $SUDO apt-get update -qq >/dev/null
  fi
  $SUDO apt-get install -y gns3-iou
  # No hostname or hostid changes are performed here.
fi

# Prepare GNS3 server paths
GNS3_LOG_DIR="/var/log/gns3"
GNS3_RUN_DIR="/var/run/gns3"
$SUDO mkdir -p "$GNS3_LOG_DIR" "$GNS3_RUN_DIR" \
  "${GNS3_HOME}/images" "${GNS3_HOME}/projects" "${GNS3_HOME}/appliances" "${GNS3_HOME}/configs"
$SUDO chown -R "$GNS3_USER:$GNS3_USER" "$GNS3_LOG_DIR" "$GNS3_RUN_DIR" "$GNS3_HOME"

# Write GNS3 server config
log "Writing /etc/gns3/gns3_server.conf"
$SUDO mkdir -p /etc/gns3
$SUDO tee /etc/gns3/gns3_server.conf >/dev/null <<EOF
[Server]
host = ${GNS3_LISTEN_HOST}
port = ${GNS3_PORT}
images_path = ${GNS3_HOME}/images
projects_path = ${GNS3_HOME}/projects
appliances_path = ${GNS3_HOME}/appliances
configs_path = ${GNS3_HOME}/configs
report_errors = True

[Qemu]
enable_hardware_acceleration = True
require_hardware_acceleration = True
EOF

if [ "$DISABLE_KVM" -eq 1 ]; then
  log "Disabling KVM in GNS3 config as requested"
  $SUDO sed -i 's/enable_hardware_acceleration = True/enable_hardware_acceleration = False/' /etc/gns3/gns3_server.conf
  $SUDO sed -i 's/require_hardware_acceleration = True/require_hardware_acceleration = False/' /etc/gns3/gns3_server.conf
fi

$SUDO chown -R "$GNS3_USER:$GNS3_USER" /etc/gns3
$SUDO chmod -R 700 /etc/gns3

# Systemd service
log "Installing systemd unit for gns3server"
$SUDO tee /lib/systemd/system/gns3.service >/dev/null <<EOF
[Unit]
Description=GNS3 server
After=network-online.target
Wants=network-online.target

[Service]
User=${GNS3_USER}
Group=${GNS3_USER}
PermissionsStartOnly=true
EnvironmentFile=-/etc/environment
ExecStartPre=/bin/mkdir -p ${GNS3_LOG_DIR} ${GNS3_RUN_DIR}
ExecStartPre=/bin/chown -R ${GNS3_USER}:${GNS3_USER} ${GNS3_LOG_DIR} ${GNS3_RUN_DIR}
ExecStart=/usr/bin/gns3server --host ${GNS3_LISTEN_HOST} --port ${GNS3_PORT} --log ${GNS3_LOG_DIR}/gns3.log
ExecReload=/bin/kill -s HUP \$MAINPID
Restart=on-failure
RestartSec=5
LimitNOFILE=16384

[Install]
WantedBy=multi-user.target
EOF

$SUDO chmod 644 /lib/systemd/system/gns3.service
$SUDO systemctl daemon-reload
$SUDO systemctl enable --now gns3

log "GNS3 installation completed successfully."
log "Service status: systemctl status gns3"
log "Listening: ${GNS3_LISTEN_HOST}:${GNS3_PORT}"

# Hints for optional integrations
if [ "$INSTALL_VMWARE" -eq 1 ]; then
  echo "Note: VMware Workstation/Player must be installed separately for vmrun support." >&2
fi
if [ "$INSTALL_VIRTUALBOX" -eq 1 ]; then
  echo "If VirtualBox kernel modules failed to build, install matching linux headers and reboot." >&2
fi
if [ "$INSTALL_DOCKER" -eq 1 ]; then
  echo "Docker installed. You might need to re-login for docker group to take effect." >&2
fi
