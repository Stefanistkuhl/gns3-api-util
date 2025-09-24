# GNS3 API Util

<p align="center">
  <img width=256 src="https://i.imgur.com/t1PNyl4.gif" alt="surely a temporary logo" />
</p>

A powerful command-line utility for managing GNS3v3 servers, with advanced template-based exercise creation for educational environments.

## Features

### **Template-Based Exercise Creation**
- **Server-based templates**: Use existing projects on the server as templates
- **File-based templates**: Import `.gns3project` files as templates
- **Interactive selection**: Fuzzy picker for choosing templates
- **Automatic duplication**: Templates are duplicated for each student group
- **Smart fallback**: Prioritizes server templates over file imports

### **Educational Workflow**
- **Class management**: Create classes with multiple student groups
- **Exercise deployment**: Deploy identical lab environments for all groups
- **Access control**: Automatic ACL setup for student access
- **Resource management**: Efficient project and node management

### **Remote Server Management**
- **HTTPS setup**: Install Caddy reverse proxy with SSL certificates
- **GNS3 server installation**: Remote installation and configuration of GNS3 servers
- **Firewall management**: Configure security rules and access restrictions
- **SSH operations**: Direct server administration via SSH
- **State file support**: Automatic configuration tracking for easy cleanup

### **Developer Tools**
- **Example scripts**: Ready-to-use bash scripts for common workflows
- **Educational examples**: Step-by-step tutorials and use cases

## Quick Start

### Installation

#### Package Managers (Recommended)
```bash
# Arch Linux (AUR)
paru -S gns3util

# macOS (Homebrew)
brew tap stefanistkuhl/tap
brew install gns3util

# Windows (Scoop)
scoop bucket add stefanistkuhl https://github.com/stefanistkuhl/bucket
scoop install gns3util
```

#### Pre-built Binaries
Download pre-built binaries from the [Releases page](https://github.com/stefanistkuhl/gns3-api-util/releases) for your platform.

#### Build from Source
```bash
# Build from source
go build -o gns3util
```

### Authentication
```bash
# Login to your GNS3 server
gns3util -s https://your-gns3-server:3080 auth login

# Or use a keyfile
gns3util -s https://your-gns3-server:3080 -k ~/.gns3/gns3key
```

### Basic Usage

#### Create a Class
```bash
# Create class from JSON file
gns3util -s https://server:3080 class create --file class.json

# Interactive class creation
gns3util -s https://server:3080 class create --interactive
```

#### Create an Exercise with Template
```bash
# Interactive template selection (recommended)
gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --select-template

# Using existing project as template
gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "NetworkTemplate" \
  --confirm=false

# Using template file
gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "/path/to/template.gns3project"
```

#### Exercise Management with Fuzzy Selection
```bash
# Interactive class selection for exercise deletion
gns3util -s https://server:3080 exercise delete --select-class

# Interactive class and group selection
gns3util -s https://server:3080 exercise delete --select-class --select-group

# Multi-select exercises for deletion
gns3util -s https://server:3080 exercise delete --select-exercise --multi

# Delete exercises from specific cluster (no server flag needed)
gns3util exercise delete --cluster production-cluster --select-exercise
```

#### Remote Server Management

**GNS3 Server Installation**:
```bash
# Install GNS3 server with default options
gns3util -s https://server:3080 remote install gns3 admin

# Install with Docker and VirtualBox support
gns3util -s https://server:3080 remote install gns3 admin \
  --install-docker \
  --install-virtualbox \
  --gns3-port 3080 \
  --home-dir /opt/gns3

# Install with IOU support (requires valid license)
gns3util -s https://server:3080 remote install gns3 admin \
  --use-iou \
  --enable-i386 \
  --username gns3-server

# Interactive installation (recommended for first-time setup)
gns3util -s https://server:3080 remote install gns3 admin --interactive
```

**GNS3 Server Uninstallation**:
```bash
# Uninstall GNS3 server (preserves user data)
gns3util -s https://server:3080 remote uninstall gns3 admin \
  --preserve-data

# Complete uninstall (removes everything)
gns3util -s https://server:3080 remote uninstall gns3 admin

# Interactive uninstallation
gns3util -s https://server:3080 remote uninstall gns3 admin --interactive
```

**HTTPS Reverse Proxy Setup**:
```bash
# Install HTTPS reverse proxy with firewall rules
gns3util -s https://server:3080 remote install https admin \
  --domain gns3.yourdomain.com \
  --firewall-allow 10.0.0.0/24

# Install with custom SSL certificate subject
gns3util -s https://server:3080 remote install https admin \
  --domain gns3.yourdomain.com \
  --subject "/CN=gns3.yourdomain.com" \
  --firewall-block

# Interactive HTTPS setup
gns3util -s https://server:3080 remote install https admin --interactive
```

**HTTPS Reverse Proxy Removal**:
```bash
# Remove HTTPS configuration (uses state file automatically)
gns3util -s https://server:3080 remote uninstall https admin

# Interactive HTTPS removal
gns3util -s https://server:3080 remote uninstall https admin --interactive
```

#### Remote Server Management Features

**State Management**: The remote installation system automatically saves installation state for easy cleanup and configuration tracking.

**GNS3 Server Installation Options**:
- **Docker Support**: Install Docker for containerized appliances
- **VirtualBox Support**: Enable VirtualBox integration
- **VMware Support**: Install VMware integration packages
- **IOU Support**: Configure IOU (IOS on Unix) support (requires valid license)
- **KVM Acceleration**: Hardware acceleration for QEMU (enabled by default)
- **Custom Configuration**: Specify custom ports, directories, and usernames

**HTTPS Reverse Proxy Options**:
- **SSL Certificates**: Automatic certificate generation and management
- **Firewall Rules**: Configure security rules and access restrictions
- **Custom Domains**: Support for custom domain names and subjects
- **Port Configuration**: Configurable reverse proxy and GNS3 server ports

**Uninstallation Options**:
- **Data Preservation**: Keep GNS3 home directory and user projects
- **Complete Removal**: Remove all GNS3 components and configurations
- **State Cleanup**: Automatic removal of installation state files
- **Selective Cleanup**: Remove only specific components (HTTPS, GNS3, etc.)

## Example Scripts

The `scripts/examples/` directory contains ready-to-use bash scripts for common workflows:

### **Template-Based Exercise Deployment**
```bash
# Deploy exercise using existing template
./scripts/examples/deploy-template-exercise.sh \
  http://gns3-server:3080 \
  "CS101" \
  "Lab1" \
  "NetworkTemplate"
```

### **Interactive Template Selection**
```bash
# Create exercise with interactive template selection
./scripts/examples/create-exercise-interactive.sh \
  http://gns3-server:3080 \
  "CS101" \
  "Lab1"
```

### **File-Based Template Import**
```bash
# Create exercise from template file
./scripts/examples/import-template-and-create-exercise.sh \
  http://gns3-server:3080 \
  "CS101" \
  "Lab1" \
  "template.gns3project"
```

### **Individual Lab Setup**
```bash
# Create individual lab projects for students
./scripts/examples/setup-class-lab.sh \
  http://gns3-server:3080 \
  5  # Number of students
```

### **Cleanup**
```bash
# Clean up projects with specific prefix
./scripts/examples/cleanup-class.sh \
  http://gns3-server:3080 \
  "Student-"  # Project name prefix
```

### **Test All Scripts**
```bash
# Run comprehensive test suite
./scripts/examples/test-all-scripts.sh http://gns3-server:3080
```

## Template System

### How Templates Work

1. **Template Selection**: Choose from existing projects or import files
2. **Automatic Duplication**: Template is duplicated for each student group
3. **Project Naming**: Uses format `{{class}}-{{exercise}}-{{group}}-{{uuid}}`
4. **Access Control**: Students only see their assigned projects

### Template Types

#### Server-Based Templates (Recommended)
- Use existing projects already on the server
- Fastest deployment
- No file upload required
- Interactive selection available

#### File-Based Templates
- Import `.gns3project` files
- Useful for sharing templates
- Automatic cleanup after import
- Fallback when server templates unavailable

### Example Class JSON
```json
{
  "name": "CS101",
  "groups": [
    {
      "name": "Group1",
      "students": [
        {"username": "student1", "password": "password123"},
        {"username": "student2", "password": "password123"}
      ]
    },
    {
      "name": "Group2", 
      "students": [
        {"username": "student3", "password": "password123"},
        {"username": "student4", "password": "password123"}
      ]
    }
  ]
}
```

## Advanced Features

### Project Management
```bash
# List all projects
gns3util -s https://server:3080 project ls

# Create new project
gns3util -s https://server:3080 project new --name "MyProject" --auto-close true

# Duplicate project
gns3util -s https://server:3080 project duplicate "MyProject" --name "MyProjectCopy"
```

### Node Management
```bash
# List nodes in project
gns3util -s https://server:3080 node ls "MyProject"

# Create nodes
gns3util -s https://server:3080 node create "MyProject" \
  --name "Router1" \
  --node-type "qemu" \
  --compute-id "local"
```

### Class Management
```bash
# List classes
gns3util -s https://server:3080 class ls

# Delete class
gns3util -s https://server:3080 class delete --name "CS101" --confirm=false
```

## Configuration

### Global Flags
- `-s, --server`: GNS3v3 Server URL (required)
- `-k, --key-file`: Path to authentication keyfile
- `-i, --insecure`: Ignore SSL certificate errors
- `--raw`: Output raw JSON instead of formatted text

### Authentication
The tool supports multiple authentication methods:
- Interactive login: `auth login`
- Keyfile: `-k ~/.gns3/gns3key`
- Environment variables: `GNS3_SERVER`, `GNS3_KEYFILE`

## Development

### Building
```bash
go build -o gns3util
```

### Contributing
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## Roadmap

### **Future Features**
- **Multi-Server Management**
  - Copy projects between servers
  - Centralized project management
  - ~~Remote install/uninstall~~ ✅ **Implemented**

- **Backup & Migration**
  - Automated project backups
  - Easy server migrations
  - Project versioning

- **Custom YAML Scripting**
  - Similar to GitHub Actions
  - Define workflows in YAML
  - Automated task execution

## Documentation

**Comprehensive Documentation Available**

- **[Online Documentation](https://stefanistkuhl.github.io/gns3-api-util/)** - Complete guide with examples
- **[CLI Reference](https://stefanistkuhl.github.io/gns3-api-util/cli-reference/commands/)** - Full command reference

### Quick Links
- [Scripts Walkthrough](https://stefanistkuhl.github.io/gns3-api-util/scripts/overview/) - Detailed script usage guide
- [Automation Guide](https://stefanistkuhl.github.io/gns3-api-util/automation/walkthrough/) - Comprehensive automation walkthrough
- [Complete Documentation](https://stefanistkuhl.github.io/gns3-api-util/) - Full documentation structure

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Support

For issues, questions, or contributions:
- Open an issue on GitHub
- Check the [online documentation](https://stefanistkuhl.github.io/gns3-api-util/)
- Review the [example scripts](scripts/examples/) for usage patterns
- Review the CLI help: `./gns3util --help`
