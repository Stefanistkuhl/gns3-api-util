# Authentication

GNS3util supports multiple authentication methods for connecting to GNS3v3 servers.

## Authentication Methods

### 1. Interactive Login (Recommended)

The most user-friendly method for first-time setup:

```bash
gns3util -s https://your-gns3-server:3080 auth login
```

This will prompt you for:
- Username
- Password
- Optionally save credentials to a keyfile

### 2. Keyfile Authentication

For automated scripts and repeated use:

```bash
# Create keyfile
mkdir -p ~/.gns3
echo "your-api-key" > ~/.gns3/gns3key
chmod 600 ~/.gns3/gns3key

# Use keyfile
gns3util -s https://server:3080 -k ~/.gns3/gns3key project ls
```

### 3. Username/Password

Direct authentication without saving credentials:

```bash
gns3util -s https://server:3080 -u username -p password project ls
```

### 4. Environment Variables

Set authentication via environment variables:

```bash
export GNS3_SERVER="https://your-gns3-server:3080"
export GNS3_KEYFILE="~/.gns3/gns3key"
# or
export GNS3_USERNAME="username"
export GNS3_PASSWORD="password"

gns3util project ls
```

## Server Configuration

### Required Flags

- `-s, --server`: GNS3v3 Server URL (required)
- `-k, --key-file`: Path to authentication keyfile
- `-u, --username`: Username for authentication
- `-p, --password`: Password for authentication
- `-i, --insecure`: Ignore SSL certificate errors

### SSL/TLS Configuration

#### Ignore Certificate Errors
```bash
gns3util -s https://server:3080 -i --help
```

#### Custom CA Certificate
```bash
export SSL_CERT_FILE="/path/to/ca-cert.pem"
gns3util -s https://server:3080 project ls
```

## Keyfile Management

### Creating a Keyfile
```bash
# Interactive creation
gns3util -s https://server:3080 auth login --save-keyfile ~/.gns3/gns3key

# Manual creation
echo "your-api-key" > ~/.gns3/gns3key
chmod 600 ~/.gns3/gns3key
```

### Keyfile Format
The keyfile should contain only the API key:
```
your-api-key-here
```

### Keyfile Security
- Store in a secure location (e.g., `~/.gns3/gns3key`)
- Set restrictive permissions (`chmod 600`)
- Consider using a password manager for the API key
- Rotate keys regularly

## Authentication Flow

### 1. Server Connection
```bash
gns3util -s https://server:3080 --help
```

### 2. Credential Validation
```bash
gns3util -s https://server:3080 auth login
```

### 3. Session Management
- Credentials are cached for the session
- Keyfile is read on each command
- No persistent session storage

## Troubleshooting

### Common Issues

#### Authentication Failed
```bash
# Check credentials
gns3util -s https://server:3080 -u username -p password project ls

# Verify keyfile
cat ~/.gns3/gns3key

# Test connection
gns3util -s https://server:3080 --help
```

#### SSL Certificate Errors
```bash
# Ignore certificate errors
gns3util -s https://server:3080 -i project ls

# Use custom CA
export SSL_CERT_FILE="/path/to/ca-cert.pem"
gns3util -s https://server:3080 project ls
```

#### Connection Refused
- Verify server URL and port
- Check network connectivity
- Ensure GNS3 server is running
- Check firewall settings

#### Permission Denied
- Verify user has appropriate permissions
- Check API key validity
- Ensure user is in correct groups

### Debug Mode

Enable verbose logging:
```bash
gns3util -s https://server:3080 --verbose project ls
```

### Testing Authentication

Test your authentication setup:
```bash
# Test basic connection
gns3util -s https://server:3080 --help

# Test authentication
gns3util -s https://server:3080 project ls

# Test specific permissions
gns3util -s https://server:3080 class ls
```

## Security Best Practices

### 1. Use Keyfiles
- Prefer keyfile authentication over username/password
- Store keyfiles securely
- Use appropriate file permissions

### 2. Rotate Credentials
- Change passwords regularly
- Rotate API keys periodically
- Monitor for unauthorized access

### 3. Network Security
- Use HTTPS for all connections
- Validate SSL certificates
- Consider VPN for remote access

### 4. Access Control
- Use least privilege principle
- Create separate accounts for different purposes
- Monitor access logs

## Examples

### Basic Authentication
```bash
# Interactive login
gns3util -s https://gns3.example.com:3080 auth login

# With keyfile
gns3util -s https://gns3.example.com:3080 -k ~/.gns3/gns3key project ls
```

### Script Authentication
```bash
#!/bin/bash
SERVER="https://gns3.example.com:3080"
KEYFILE="~/.gns3/gns3key"

gns3util -s "$SERVER" -k "$KEYFILE" project ls
```

### Environment-based Authentication
```bash
#!/bin/bash
export GNS3_SERVER="https://gns3.example.com:3080"
export GNS3_KEYFILE="~/.gns3/gns3key"

gns3util project ls
```
