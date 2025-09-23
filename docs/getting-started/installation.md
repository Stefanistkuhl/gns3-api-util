# Installation

## Prerequisites

- Go 1.19 or later
- Access to a GNS3v3 server
- Network connectivity to the GNS3 server

## Package Managers

### AUR (Arch Linux)
```bash
# Using paru (recommended)
paru -S gns3util

# Or using yay
yay -S gns3util
```

### Homebrew (macOS)
```bash
# Add the tap first
brew tap Stefanistkuhl/tap

# Then install gns3util
brew install gns3util
```

### Windows - Coming Soon
```bash
scoop bucket add gns3util https://github.com/stefanistkuhl/bucket.git
scoop install gns3util
```
```

## Building from Source

### Clone the Repository
```bash
git clone https://github.com/stefanistkuhl/gns3-api-util.git
cd gns3-api-util
```

### Build the Binary
```bash
go build -o gns3util
```

### Install Dependencies
```bash
go mod download
```

## Pre-built Binaries

Pre-built binaries are available for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

Download from the [releases page](https://github.com/stefanistkuhl/gns3-api-util/releases).


## Verification

Test your installation:
```bash
gns3util --help
```

You should see the help output with available commands.

## Configuration

### Environment Variables (Optional)
```bash
# Optional: Set default server and keyfile
export GNS3_SERVER="https://your-gns3-server:3080"
export GNS3_KEYFILE="~/.gns3/gns3key"
```

### Authentication Keyfile
The program will automatically create a keyfile when you use interactive login:
```bash
# The program creates ~/.gns3/gns3key automatically
gns3util -s https://server:3080 auth login
```

## Troubleshooting

### Common Issues

#### Permission Denied
```bash
chmod +x gns3util
```

#### SSL Certificate Errors
```bash
gns3util -s https://server:3080 -i --help
```

#### Connection Refused
- Verify the GNS3 server is running
- Check the server URL and port
- Ensure network connectivity

#### Authentication Failed
- Verify your API key is correct
- Check the keyfile permissions
- Try interactive login: `gns3util auth login`
