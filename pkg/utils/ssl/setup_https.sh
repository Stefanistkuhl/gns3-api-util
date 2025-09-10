#!/usr/bin/env bash
set -e
set -o pipefail

# Require root or passwordless sudo
if [ "$EUID" -ne 0 ] && ! sudo -n true &>/dev/null; then
  echo "Error: must be root or have passwordless sudo" >&2
  exit 1
fi
SUDO=""
if [ "$EUID" -ne 0 ]; then
  SUDO="sudo"
fi

UFW=""
RP_PORT=""
GNS3_PORT=""
DOMAIN=""

# Check if Caddy is already installed and running
echo "Checking Caddy installation status..."
if command -v caddy &>/dev/null; then
  echo "Caddy already installed: $(caddy version)"
  SKIP_INSTALL=true
else
  echo "Caddy not found in PATH, checking package status..."
  # Check if Caddy package is installed but binary is missing (Ubuntu issue)
  if dpkg -l | grep "^ii.*caddy" >/dev/null 2>&1; then
    echo "Caddy package is installed but binary is missing, fixing..."
    $SUDO apt-get install --reinstall -y caddy >/dev/null 2>&1
    if command -v caddy &>/dev/null; then
      echo "Caddy binary fixed: $(caddy version)"
      SKIP_INSTALL=true
    else
      echo "Failed to fix Caddy binary, will reinstall"
      SKIP_INSTALL=false
    fi
  else
    echo "Caddy package not found, will install"
    SKIP_INSTALL=false
  fi
fi
echo "SKIP_INSTALL=$SKIP_INSTALL"

# Load OS info
. /etc/os-release

install_debian() {
  echo "Installing Caddy on Debian/Ubuntu/Raspbian..."
  export DEBIAN_FRONTEND=noninteractive
  $SUDO apt-get update -qq > /dev/null 2>&1
  $SUDO apt-get install -y -qq \
    debian-keyring debian-archive-keyring \
    apt-transport-https curl gnupg $UFW \
    > /dev/null 2>&1
  
  CADDY_GPG_KEYRING="/usr/share/keyrings/caddy-stable-archive-keyring.gpg"

  if [ -f "$CADDY_GPG_KEYRING" ]; then
    echo "Removing existing Caddy GPG key: $CADDY_GPG_KEYRING"
    $SUDO rm -f "$CADDY_GPG_KEYRING"
  fi

  curl -1sLf \
    'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' \
    | $SUDO gpg --dearmor --batch \
      -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg

  curl -1sLf \
    'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' \
    | $SUDO tee /etc/apt/sources.list.d/caddy-stable.list \
      >/dev/null

  $SUDO apt-get update -qq >/dev/null 2>&1
  $SUDO apt-get install -y caddy >/dev/null 2>&1
  
  # Verify installation
  if ! command -v caddy &>/dev/null; then
    echo "Error: Caddy installation failed" >&2
    exit 1
  fi
}

install_fedora_rhel() {
  if command -v dnf &>/dev/null; then
    echo "Installing Caddy via COPR (dnf)..."
    $SUDO dnf install -y 'dnf-command(copr)'
    $SUDO dnf copr enable -y @caddy/caddy
    $SUDO dnf install -y caddy $UFW
  elif command -v yum &>/dev/null; then
    echo "Installing Caddy via COPR (yum)..."
    $SUDO yum install -y yum-plugin-copr
    $SUDO yum copr enable @caddy/caddy
    $SUDO yum install -y caddy $UFW
  else
    echo "Error: no dnf or yum found" >&2
    exit 1
  fi
}

install_arch() {
  echo "Installing Caddy on Arch/Manjaro..."
  $SUDO pacman -Sy --noconfirm caddy $UFW
}

install_caddy() {
  case "$ID" in
    debian|ubuntu|raspbian)
      install_debian
      ;;
    fedora|centos|rhel)
      install_fedora_rhel
      ;;
    arch|manjaro)
      install_arch
      ;;
    *)
      echo "Unsupported distro: $ID. Install Caddy manually." >&2
      exit 1
      ;;
  esac
}

if [ "$SKIP_INSTALL" = "false" ]; then
  install_caddy
  
  # Verify installation
  if ! command -v caddy &>/dev/null; then
    echo "Error: Caddy installation failed" >&2
    exit 1
  fi
fi

# Create caddy user and group if they don't exist
if ! id "caddy" &>/dev/null; then
  $SUDO useradd --system --home /var/lib/caddy --shell /usr/sbin/nologin --no-create-home caddy
fi

# Create caddy directories
$SUDO mkdir -p /var/lib/caddy
$SUDO mkdir -p /etc/caddy
$SUDO chown -R caddy:caddy /var/lib/caddy
$SUDO chown -R caddy:caddy /etc/caddy

# Certificate subject
SUBJ=""

# Generate a self-signed cert
CERT_DIR=/etc/caddy/certs
$SUDO mkdir -p "$CERT_DIR"
$SUDO openssl req -new -x509 -days 365 -nodes \
  -out "$CERT_DIR/gns3.cert" \
  -keyout "$CERT_DIR/gns3.key" \
  -subj "$SUBJ" \
  >/dev/null 2>&1
$SUDO chown -R caddy:caddy "$CERT_DIR"

# Write Caddyfile
$SUDO tee /etc/caddy/Caddyfile >/dev/null <<CADDYFILE_EOF
:${RP_PORT} {
    reverse_proxy 127.0.0.1:${GNS3_PORT}
    tls ${CERT_DIR}/gns3.cert ${CERT_DIR}/gns3.key
}
CADDYFILE_EOF

# Configure firewall rules
UFW_ENABLE

# Enable UFW if not already enabled
if ! ufw status | grep -q "Status: active"; then
    echo "Enabling UFW firewall..."
    $SUDO ufw --force enable
    echo "UFW firewall enabled"
else
    echo "UFW firewall already active"
fi

# Configure systemd service for better reliability
# Create the service directory if it doesn't exist
$SUDO mkdir -p /etc/systemd/system/caddy.service.d

$SUDO tee /etc/systemd/system/caddy.service.d/override.conf >/dev/null <<'OVERRIDE_EOF'
[Service]
# Override only restart and reliability settings
Restart=always
RestartSec=5
StartLimitInterval=60s
StartLimitBurst=3
TimeoutStopSec=5s
LimitNOFILE=1048576
LimitNPROC=1048576
OVERRIDE_EOF

# Reload systemd to pick up the new configuration
$SUDO systemctl daemon-reload

# Enable and start the service with safety restarts
$SUDO systemctl enable caddy

# Function to safely start Caddy with retries
start_caddy_safely() {
  local max_attempts=3
  local attempt=1
  
  echo "Starting Caddy service with safety restarts..."
  
  while [ $attempt -le $max_attempts ]; do
    echo "Attempt $attempt/$max_attempts: Starting Caddy..."
    
    # Stop any existing Caddy processes
    $SUDO systemctl stop caddy 2>/dev/null || true
    sleep 1
    
    # Start Caddy
    if $SUDO systemctl start caddy; then
      echo "Caddy start command succeeded"
      
      # Wait for service to fully start
      sleep 3
      
      # Check if service is actually running
      if systemctl is-active --quiet caddy; then
        echo "Caddy service is running successfully"
        return 0
      else
        echo "Caddy service failed to start properly (attempt $attempt)"
        $SUDO systemctl status caddy --no-pager -l
      fi
    else
      echo "Caddy start command failed (attempt $attempt)"
      $SUDO systemctl status caddy --no-pager -l
    fi
    
    # If this wasn't the last attempt, wait before retrying
    if [ $attempt -lt $max_attempts ]; then
      echo "Waiting 5 seconds before retry..."
      sleep 5
    fi
    
    attempt=$((attempt + 1))
  done
  
  echo "Caddy failed to start after $max_attempts attempts"
  return 1
}

# Start Caddy with safety restarts
if ! start_caddy_safely; then
  echo "Error: Caddy service failed to start after multiple attempts"
  echo "Checking system logs for more details..."
  $SUDO journalctl -u caddy --no-pager -l --since "5 minutes ago"
  exit 1
fi

# Final verification and health check
if ! systemctl is-active --quiet caddy; then
  echo "Error: Caddy service is not running after successful start"
  $SUDO systemctl status caddy --no-pager -l
  exit 1
fi

# Health check function
check_caddy_health() {
  local max_attempts=5
  local attempt=1
  
  echo "Performing Caddy health check..."
  
  while [ $attempt -le $max_attempts ]; do
    echo "Health check attempt $attempt/$max_attempts..."
    
    # Check if Caddy is listening on the expected port
    if ss -tulpn | grep -q ":$RP_PORT "; then
      echo "Caddy is listening on port $RP_PORT"
      
      # Test if Caddy responds to HTTP requests
      if curl -s -k --connect-timeout 5 --max-time 10 "https://localhost:$RP_PORT" >/dev/null 2>&1; then
        echo "Caddy is responding to HTTPS requests"
        return 0
      else
        echo "Caddy is not responding to HTTPS requests (attempt $attempt)"
      fi
    else
      echo "Caddy is not listening on port $RP_PORT (attempt $attempt)"
    fi
    
    if [ $attempt -lt $max_attempts ]; then
      echo "Waiting 3 seconds before retry..."
      sleep 3
    fi
    
    attempt=$((attempt + 1))
  done
  
  echo "Caddy health check failed after $max_attempts attempts"
  return 1
}

# Perform health check
if ! check_caddy_health; then
  echo "Warning: Caddy health check failed, but service is running"
  echo "This might be due to certificate issues or configuration problems"
  echo "Checking Caddy logs..."
  $SUDO journalctl -u caddy --no-pager -l --since "2 minutes ago"
  echo "Continuing with installation..."
fi

echo "Caddy service is running and healthy"

# Create renewal script
RENEW=/usr/local/bin/renew-caddy-gns3-cert.sh
$SUDO tee "$RENEW" >/dev/null <<'RENEW_SCRIPT_EOL'
#!/usr/bin/env bash
set -euo pipefail
CERT_DIR=/etc/caddy/certs
openssl req -new -x509 -days 365 -nodes \
  -out "$CERT_DIR/gns3.cert" \
  -keyout "$CERT_DIR/gns3.key" \
  -subj '$SUBJ'
sudo chown caddy:caddy "$CERT_DIR/gns3.cert" "$CERT_DIR/gns3.key"
sudo systemctl reload caddy
RENEW_SCRIPT_EOL
$SUDO chmod +x "$RENEW"

# Schedule cron renewal
CRON_DAY=$(date +%-d)
CRON_MONTH=$(date -d '+364 days' +%-m)
CRON_JOB="0 0 $CRON_DAY $CRON_MONTH * /bin/bash $RENEW"
( $SUDO crontab -l 2>/dev/null; echo "$CRON_JOB" ) \
  | $SUDO crontab -

echo "Caddy installed, cert generated, cron job added:"
echo "   $CRON_JOB"

# Save installation state on remote server
STATE_DIR="/etc/gns3"
$SUDO mkdir -p "$STATE_DIR"
STATE_FILE="$STATE_DIR/ssl_state.json"
$SUDO tee "$STATE_FILE" >/dev/null << STATE_EOF
{
  "server_host": "$(hostname)",
  "install_time": "$(date -Iseconds)",
  "reverse_proxy_port": $RP_PORT,
  "gns3_port": $GNS3_PORT,
  "domain": "$DOMAIN",
  "firewall_block": $FIREWALL_BLOCK,
  "firewall_allow": "$FIREWALL_ALLOW",
  "distro": "$ID",
  "ufw_enabled": true,
  "ufw_rules": [
    "allow ssh",
    "allow 22",
    "deny $GNS3_PORT"
  ]
}
STATE_EOF

$SUDO chown root:root "$STATE_FILE"
$SUDO chmod 644 "$STATE_FILE"
echo "Installation state saved to: $STATE_FILE"

# Also save a copy in /var/lib/gns3 for immediate access
$SUDO mkdir -p /var/lib/gns3
$SUDO cp "$STATE_FILE" "/var/lib/gns3/ssl_state.json"
$SUDO chown root:root "/var/lib/gns3/ssl_state.json"
$SUDO chmod 644 "/var/lib/gns3/ssl_state.json"
echo "State backup saved to: /var/lib/gns3/ssl_state.json"
