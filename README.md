# GNS3UTIL

> **UNDER DEVELOPMENT** - This project is in active development. APIs and CLI commands are subject to change.

<p align="center">
  <img width=256 src="https://github.com/0xveya/goobering/blob/master/images/gns3util_real.webp?raw=true" alt="a cat with nerd emoji glasses and an explosion in the backround" />
</p>

A toolkit for GNS3 lab management with two main components:

1. **Cluster Orchestration** (New) - Distributed system for managing labs at scale with a master control plane and distributed file storage
2. **GNS3 API Wrapper** (Maintained) - Command-line interface for direct GNS3v3 server management

## Architecture

GNS3UTIL is built as a distributed cluster system with two main components:

### **Cluster Master** (`cmd/master/`)
Central control plane managing the cluster:
- Authentication and authorization (RBAC)
- Cluster node management and health monitoring
- Distributed state via ETCD
- Job orchestration and scheduling
- mDNS service discovery

### **File Store Nodes** (`cmd/file-store/`)
Distributed storage nodes for lab files:
- Multi-protocol support (HTTP/TLS, HTTP/3)
- File and bucket management with permissions
- Metadata via Turso (SQLite edge database)
- Automatic storage optimization and cleanup
- Public token-based file sharing

## Core Features

### **Cluster Orchestration**
- **Master-based coordination**: Central control of distributed cluster
- **Node management**: Register, monitor, and manage file store nodes
- **Health checks**: Automatic node status monitoring
- **Job scheduling**: Distributed job execution across nodes

### **Distributed File Storage**
- **Bucket organization**: Organize files into logical buckets
- **Access control**: Fine-grained permissions for files and buckets
- **Multi-protocol**: HTTP/TLS and HTTP/3 (QUIC) support
- **Public sharing**: Generate public tokens for temporary file access
- **Storage optimization**: Automatic cleanup of expired and orphaned files

### **Identity & Security**
- **JWT-based authentication**: Token generation and validation
- **Role-based access control**: Manage users and roles
- **Certificate management**: Automatic TLS certificate generation and management
- **Cluster authentication**: Secure inter-node communication

### **Observability**
- **OpenTelemetry integration**: Distributed tracing and metrics
- **Prometheus metrics**: Storage, upload, and API performance metrics
- **Health endpoints**: Built-in health check endpoints

## Getting Started

### Build from Source
```bash
# Build the CLI
cd cmd/master && go build -o gns3util-master
cd cmd/file-store && go build -o gns3util-filestore

# Or build all components
go build ./cmd/...
```

### Running Components

**Master Node**:
```bash
./gns3util-master
```

**File Store Node**:
```bash
./gns3util-filestore
```

### Environment Configuration

Both components support configuration via environment variables. See:
- [`cmd/master/README.md`](cmd/master/README.md) for master configuration
- [`cmd/file-store/README.md`](cmd/file-store/README.md) for file store configuration

## CLI Tools

GNS3UTIL includes both a legacy GNS3 API wrapper and the new cluster orchestration CLI.

### Cluster Orchestration CLI (New)

The `cluster-control` (alias: `ctl`) tool manages cluster orchestration. Current commands:

**Cluster Management**:
```bash
# Create and manage clusters
gns3util ctl create <command>
gns3util ctl auth <command>
gns3util ctl add-cluster <command>
gns3util ctl add-discover <command>
gns3util ctl remove <command>
```

**Storage**:
```bash
# Manage object storage buckets and files
gns3util ctl obj <command>
```

**Operations**:
```bash
# Manage distributed jobs
gns3util ctl jobs <command>
```

**RBAC & Identity**:
```bash
# Manage users and roles
gns3util ctl users <command>
gns3util ctl roles <command>
```

For complete reference:
```bash
gns3util ctl --help
```

### Legacy GNS3 API Wrapper (Maintained)

The original GNS3 API wrapper commands are still available:
```bash
# Project management
gns3util project <command>

# Node management  
gns3util node <command>

# Class and exercise management
gns3util class <command>
gns3util exercise <command>

# User and authentication
gns3util user <command>
gns3util auth <command>

# Cluster configuration (legacy)
gns3util cluster <command>

# Server operations
gns3util remote <command>
```

Run `gns3util --help` for the complete command tree.

## Development

### Prerequisites
This project uses [mise](https://mise.jdx.dev/) for task management and tool versioning.

### Building

**Build CLI**:
```bash
mise run build-cli
```

**Build Master Node**:
```bash
go build ./cmd/master
```

**Build File Store Node**:
```bash
go build ./cmd/file-store
```

**Build all platforms**:
```bash
mise run build-all
```

### Development Workflow

**Run formatter (required before PR)**:
```bash
mise run format
```
This runs:
- `goimports` and `gofumpt` for Go code
- `sleek` for SQL formatting
- `buf format` for Protobuf files
- `swag fmt` for Swagger docs

**Run linter (required before PR)**:
```bash
mise run lint
```
This runs:
- `golangci-lint` for Go code
- `buf lint` for Protobuf files

**Run tests**:
```bash
mise run test
```

**Generate code** (after modifying proto, SQL, or handlers):
```bash
mise run generate
```
This runs:
- `sqlc` - Generate SQL query code
- `buf generate` - Generate Protobuf files
- `swag init` - Generate Swagger documentation

**Development mode** (with hot reload):
```bash
# Run all cluster components
mise run dev-cluster

# Or run individually
mise run dev-master
mise run dev-filestore
```

### Contributing

1. Clone the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes in the appropriate directory:
   - `cmd/master/` - Master control plane
   - `cmd/file-store/` - File storage node
   - `cmd/gns3util/` - CLI tool
   - `internal/` - Core implementation
   - `pkg/` - Shared packages
4. Run `mise run format` to format your code
5. Run `mise run lint` to check for issues
6. Run `mise run test` to ensure tests pass
7. If you modified proto files or SQL queries, run `mise run generate`
8. Commit your changes: `git commit -m "feat: description of changes"`
9. Push to your fork and submit a pull request

**PR Requirements**:
- ✅ `mise run format` must pass
- ✅ `mise run lint` must pass
- ✅ `mise run test` must pass
- ✅ All generated code is up-to-date (run `mise run generate` if needed)

## Roadmap

### **In Progress**
- **Cluster Master**: Control plane with ETCD state management
- **File Store Nodes**: Distributed storage with Turso/SQLite
- **Authentication**: JWT-based token system with RBAC
- **CLI Orchestration**: Commands for cluster and storage management

### **Future Features**
- **Replication**: Cross-node data replication
- **Metrics & Monitoring**: Enhanced observability
- **GNS3 Lab Integration**: Deploy GNS3 labs across cluster
- **Advanced Scheduling**: Job scheduling and orchestration
- **Configuration Management**: Cluster-wide configuration sync

## Documentation

### Cluster Orchestration (New)
- **[API Reference](https://gns3util.saygex.xyz/api-reference/)** - Complete cluster API documentation
- **[Master Node](cmd/master/README.md)** - Cluster control plane configuration and features
- **[File Store Node](cmd/file-store/README.md)** - Distributed storage node setup and API

### GNS3 API Wrapper (Legacy)
- **[GNS3UTIL Documentation](https://stefanistkuhl.github.io/gns3-api-util/)** - Full GNS3 API wrapper guide
- **[CLI Reference](https://stefanistkuhl.github.io/gns3-api-util/cli-reference/)** - Complete command reference

### Generated Documentation
- **[Auto-generated API Docs](./docs/)** - OpenAPI/Swagger documentation

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Support

For issues, questions, or contributions:
- Open an issue on GitHub
- Check component-specific READMEs in `cmd/`
- Review API documentation in generated `docs/` files
- Run `./gns3util --help` for CLI usage
