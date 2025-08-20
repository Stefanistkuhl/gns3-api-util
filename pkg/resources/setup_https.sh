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

# Exit if Caddy already installed
if command -v caddy &>/dev/null; then
  echo "Caddy already installed: $(caddy version)"
  exit 0
fi

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

install_caddy

# Enable & start the service
$SUDO systemctl enable --now caddy

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
$DOMAIN:$RP_PORT {
    reverse_proxy 127.0.0.1:$GNS3_PORT
    tls $CERT_DIR/gns3.cert $CERT_DIR/gns3.key
}
CADDYFILE_EOF

# Reload Caddy to pick up new config
$SUDO systemctl reload caddy

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
chown caddy:caddy "$CERT_DIR/gns3.cert" "$CERT_DIR/gns3.key"
systemctl reload caddy
RENEW_SCRIPT_EOL
$SUDO chmod +x "$RENEW"

# Schedule cron renewal
CRON_DAY=$(date +%-d)
CRON_MONTH=$(date -d '+364 days' +%-m)
CRON_JOB="0 0 $CRON_DAY $CRON_MONTH * /bin/bash $RENEW"
( $SUDO crontab -l 2>/dev/null; echo "$CRON_JOB" ) \
  | $SUDO crontab -

UFW_ENABLE

echo "Caddy installed, cert generated, cron job added:"
echo "   $CRON_JOB"
