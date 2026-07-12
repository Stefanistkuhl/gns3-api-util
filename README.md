# GNS3 API Util

## Dead Project

This project is basically dead.

GNS3 was not flexible enough for what I wanted. It is hard to scale, hard to integrate with other systems, and not really built to run in a distributed way. Because of that I started working on [**tethux**](https://github.com/0xveya/tethux), which is meant to replace it for my use cases.

tethux is still in development and not usable yet. The only reason this repo is not archived is in case there is a serious bug in the CLI that needs fixing.

It probably will not be ready in 2026. Depending on when you are reading this, it might exist or it might still be unfinished.

tethux is basically what I originally tried to do with `gns3util`, but instead of gluing things on top of GNS3, it is built in from the start.

## Why tethux Exists

This is mostly about the problems I ran into with GNS3.

* **Compute and clustering are awkward.** GNS3 has a “compute” concept, but it feels bolted on. You have to manually assign nodes to servers, and it does not feel like a real cluster. With `gns3util` I ended up using SQLite as glue just to track where things were and what API calls to make, which is not a good solution.
* **No real abstraction over where things run.** You still have to care which server a node is on. There is no proper scheduling or automatic placement.
* **No replication or HA.** If something dies, it dies. There is no built-in way to replicate nodes or services.
* **Integration is painful.** Things like object storage, backups, or external systems are not easy to plug in.
* **Auth and RBAC are weak.** Roles are mostly cosmetic. You cannot properly assign roles to users or groups in a meaningful way. There is no solid OAuth/OIDC story either.
* **Single server mindset.** You often end up connecting to specific servers instead of interacting with a cluster as a whole.

On top of that, the direction of the project did not help.

With GNS3 3.1 they did a big frontend rewrite, added an AI assistant, and changed the UI in ways that feel like a regression. Meanwhile, a lot of backend issues are still there. The project feels messy and not very focused.

Because of all that I lost interest in trying to work around it.

tethux is me trying to build something that actually treats compute, clustering, and integration as first class parts of the system. I also want things like importing Packet Tracer topologies, proper OAuth for users, and real RBAC that actually works.

It is not done, and it might take a long time, but it is closer to what I wanted than trying to keep patching GNS3.

> ~~Complete overhaul of the clustering system and the CLI in progress~~ <br>
> ~~No changes will happen on `master` until v2 is done.~~ <br>
> ~~Check the `features/orchestration` branch for progress.~~

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

### **Cluster Management**
- **Cluster creation**: Provision logical clusters that coordinate multiple GNS3 servers
- **Node enrollment**: Add single or multiple nodes to scaling clusters on demand
- **Configuration tooling**: Manage cluster configuration through `cluster config`
- **Topology visibility**: List clusters and view class/exercise distribution across nodes

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

# Launch interactive class builder on a custom address
gns3util -s https://server:3080 class create --interactive  --port 9090

# Create a class and register it with a cluster
gns3util class create --cluster production-cluster --file class.json
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

#### Class Operations
```bash
# List classes and show node distribution
gns3util class ls --cluster production-cluster

# Delete a single class non-interactively without confirmation
gns3util -s https://server:3080 class delete --name "CS101" --no-confirm

# Delete classes via fuzzy finder (multi-select)
gns3util -s https://server:3080 class delete --multi

# Remove a class and its exercises from a cluster definition
gns3util class delete --cluster production-cluster --name "CS101" --delete-exercises --no-confirm
```

#### Exercise Operations
```bash
# List exercises across the cluster
gns3util exercise ls --cluster production-cluster

# Filter exercise list by class 
gns3util -s https://server:3080 exercise ls --class "CS101"

# Delete all exercises for a class from the controller
gns3util -s https://server:3080 exercise delete --class "CS101" --no-confirm

# Delete multiple exercises interactively with multi-select
gns3util -s https://server:3080 exercise delete --select-exercise --multi
```

#### Cluster Operations
```bash
# Create a new cluster definition
gns3util cluster create --name production-cluster

# Add nodes to a cluster (repeat --server for each node)
gns3util cluster add-nodes production-cluster \
  --server https://cluster-node-01:3080 \
  --server https://cluster-node-02:3080 \
  --user admin --password "$GNS3_PASSWORD"

# Review cluster configuration and membership
gns3util cluster ls

# Edit cluster defaults (opens file in $EDITOR)
gns3util cluster config edit

# Apply an updated cluster configuration file
gns3util cluster config apply cluster.yaml
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

### Cluster Management Commands
```bash
# Add a single node to an existing cluster
gns3util cluster add-node production-cluster \
  --server https://edge-01:3080 \
  --user admin --password "$GNS3_PASSWORD" \
  --weight 6

# Add multiple nodes from a configuration file
gns3util cluster add-nodes production-cluster \
  --server https://edge-02:3080 \
  --server https://edge-03:3080

# Edit stored cluster configuration values
gns3util cluster config edit

# Synchronize the edited config back to the database
gns3util cluster config sync
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
