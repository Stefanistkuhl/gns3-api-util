# Remote Operations

The `gns3util remote` command provides direct server management capabilities via SSH, allowing you to perform administrative tasks that go beyond the standard GNS3 API.

## Overview

Remote operations enable you to:
- Install and configure HTTPS reverse proxies
- Manage firewall rules for enhanced security
- Perform server maintenance tasks
- Set up SSL certificates automatically

## HTTPS Reverse Proxy Installation [GIF]

### Basic Installation

Install a Caddy-based HTTPS reverse proxy for your GNS3 server:

```bash
# Basic HTTPS setup
gns3util -s https://your-gns3-server:3080 remote install https username

# Interactive setup with prompts
gns3util -s https://your-gns3-server:3080 remote install https username --interactive
```

### Advanced Configuration

```bash
# Custom domain and firewall rules
gns3util -s https://your-gns3-server:3080 remote install https username \
  --domain gns3.yourdomain.com \
  --firewall-allow 10.0.0.0/24 \
  --gns3-port 3080 \
  --reverse-proxy-port 443 \
  --subject "/CN=gns3.yourdomain.com" \
  --verbose
```

### Firewall Security Options

The HTTPS installation includes powerful firewall management:

```bash
# Block all external access to GNS3 port (only allow localhost)
gns3util -s https://your-gns3-server:3080 remote install https username --firewall-block

# Allow only specific subnet access to GNS3 port
gns3util -s https://your-gns3-server:3080 remote install https username --firewall-allow 192.168.1.0/24

# No firewall changes (default)
gns3util -s https://your-gns3-server:3080 remote install https username
```

### SSH Configuration

```bash
# Custom SSH key and port
gns3util -s https://your-gns3-server:3080 remote install https username \
  --key ~/.ssh/custom_key \
  --port 2222
```

## HTTPS Reverse Proxy Removal [GIF]

### Basic Removal

```bash
# Remove HTTPS configuration
gns3util -s https://your-gns3-server:3080 remote uninstall https username
```

### Complete Cleanup

```bash
# Remove with all original settings
gns3util -s https://your-gns3-server:3080 remote uninstall https username \
  --domain gns3.yourdomain.com \
  --firewall-allow 10.0.0.0/24 \
  --gns3-port 3080 \
  --reverse-proxy-port 443
```

### State File Support

The uninstall command automatically detects and uses state files from previous installations:

```bash
# If a state file exists, no additional arguments needed!
gns3util -s https://your-gns3-server:3080 remote uninstall https username

# State file contains:
# - Domain used during installation
# - Firewall configuration
# - Port settings
# - SSL certificate details
```

### Interactive Removal

```bash
# Interactive cleanup with prompts
gns3util -s https://your-gns3-server:3080 remote uninstall https username --interactive
```

## Complete Workflow Examples [GIF]

### Secure HTTPS Setup with Firewall

```bash
# 1. Install HTTPS with subnet restriction
gns3util -s https://gns3-server.com:3080 remote install https admin \
  --domain gns3.company.com \
  --firewall-allow 10.0.0.0/16 \
  --verbose

# 2. Verify HTTPS access
gns3util -s https://gns3.company.com:443 --help

# 3. Update gns3util to use HTTPS
gns3util -s https://gns3.company.com:443 auth login
```

### Complete Cleanup

```bash
# Remove HTTPS configuration
gns3util -s https://gns3.company.com:443 remote uninstall https admin \
  --domain gns3.company.com \
  --firewall-allow 10.0.0.0/16

# Verify removal
gns3util -s http://gns3-server.com:3080 --help
```

## Command Reference

### Global Flags

All remote commands support these global flags:

- `-s, --server string`: GNS3v3 Server URL (required)
- `-k, --key-file string`: Set a location for a keyfile to use
- `-i, --insecure`: Ignore unsigned SSL-Certificates
- `--raw`: Output all data in raw json

### Install HTTPS Flags

- `--domain string`: Domain to use for the reverse proxy
- `--firewall-allow string`: Block all connections to the GNS3 server port and only allow a given subnet (e.g., 10.0.0.0/24)
- `--firewall-block`: Block all connections to the port of the GNS3 server
- `--gns3-port int`: Port of the GNS3 Server (default: 3080)
- `--interactive`: Set the options for this command interactively
- `--key string`: Path to a custom SSH private key file
- `--port int`: SSH port (default: 22)
- `--reverse-proxy-port int`: Port for the reverse proxy to use (default: 443)
- `--subject string`: Set the subject alternative name for the SSL certificate (default: "/CN=localhost")
- `--verbose`: Run this command with extra logging

### Uninstall HTTPS Flags

- `-d, --domain string`: Domain that was used for the reverse proxy
- `-a, --firewall-allow string`: Firewall allow subnet that was used (e.g., 10.0.0.0/24)
- `-b, --firewall-block`: Whether firewall rules were configured
- `-g, --gns3-port int`: Port of the GNS3 Server (default: 3080)
- `-t, --interactive`: Set the options for this command interactively
- `--key string`: Path to a custom SSH private key file
- `-p, --port int`: SSH port (default: 22)
- `-r, --reverse-proxy-port int`: Port for the reverse proxy that was used (default: 443)
- `--subject string`: Subject that was used for the SSL certificate (default: "/CN=localhost")
- `-v, --verbose`: Enable verbose output

## Troubleshooting

### Common Issues

**SSH Connection Failed**
- Verify SSH key permissions: `chmod 600 ~/.ssh/your_key`
- Check SSH port: `--port 2222` if using non-standard port
- Verify SSH connection manually: `ssh username@server`

**Firewall Rules Not Applied**
- Ensure you have sudo privileges on the target server
- Check if ufw/iptables is available
- Use `--verbose` flag for detailed logging

**SSL Certificate Issues**
- Verify domain DNS points to your server
- Check if port 80/443 are accessible
- Use `--subject` flag for custom certificate details

**Permission Denied**
- Ensure the user has sudo privileges
- Check if Caddy can bind to port 443
- Verify firewall management permissions

**State File Issues**
- State file is automatically created during installation
- Contains all configuration settings for easy uninstall
- If state file is missing, provide all original settings manually
- State file location: `/tmp/gns3util-https-state.json` (on remote server)

### Logs and Debugging

```bash
# Enable verbose output for troubleshooting
gns3util -s https://your-server:3080 remote install https username --verbose

# Check Caddy logs on the server
sudo journalctl -u caddy -f

# Verify firewall rules
sudo ufw status
```

## Security Considerations

- **Firewall Rules**: Use `--firewall-allow` to restrict access to specific subnets
- **SSH Keys**: Use dedicated SSH keys for remote operations
- **Domain Validation**: Ensure your domain properly resolves to the server
- **Certificate Management**: Caddy automatically handles Let's Encrypt certificates
- **Port Security**: Consider using non-standard ports for additional security