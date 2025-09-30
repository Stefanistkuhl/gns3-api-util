# gns3util CLI Command Structure Reference

> This file serves as a reference for documentation.
## Root Command
```
gns3util [flags]
gns3util [command]

```

### Global Flags
- `-s, --server string`: GNS3v3 Server URL (required)
- `-k, --key-file string`: Set a location for a keyfile to use
- `-i, --insecure`: Ignore unsigned SSL-Certificates
- `--no-color`: Disable ANSI color output
- `--raw`: Output all data in raw json
- `-h, --help`: Help for gns3util

## Main Command Groups

### Authentication Commands
```
gns3util auth [flags] [command]
├── login       # Log in as user
└── status      # Check the current status of your authentication
```

### Project Operations
```
gns3util project [command]
├── close         # Close a project
├── delete        # Delete a project
├── duplicate     # Duplicate a project
├── export        # Export a project from GNS3
├── file          # Get a file from a project
├── import        # Import a project from a portable archive
├── info          # Get a project by id or name
├── load          # Load a project from a given path
├── lock          # Lock all drawings and nodes in a project
├── locked        # Get if a project is locked by id or name
├── ls            # Get the projects of the GNS3 Server
├── new           # Create a project
├── open          # Open a project
├── start-capture # Start a packet capture in a project on a given link
├── stats         # Get project-stats by id or name
├── unlock        # Unlock all drawings and nodes in a project
├── update        # Update a project
└── write-file    # Write a file to a project
```

### Node Operations
```
gns3util node [command]
├── auto-idle-pc           # Get the auto-idle-pc of a node in a project by id or name
├── auto-idle-pc-proposals # Get the auto-idle-pc-proposals of a node in a project by id or name
├── console-reset          # Reset a console for a given node
├── create                 # Create a node in a project
├── delete                 # Delete a node from a project
├── duplicate              # Duplicate a node in a project
├── from-template          # Create a node from a template
├── info                   # Get a node in a project by name or id
├── links                  # Get links of a given node in a project by id or name
├── ls                     # Get the nodes within a project by name or id
├── node-file              # Get a file from a node
├── node-isolate           # Isolate a node (suspend all attached links)
├── node-unisolate         # Un-isolate a node (resume all attached suspended links)
├── reload-all             # Reload all nodes belonging to a project
├── start-all              # Start all nodes belonging to a project
├── stop-all               # Stop all nodes belonging to a project
├── suspend-all            # Suspend all nodes belonging to a project
└── update                 # Update a node in a project
```

### Exercise Operations
```
gns3util exercise [command]
├── create      # Create an exercise (project) for every group in a class with ACLs
└── delete      # Delete an exercise
```

### Class Operations
```
gns3util class [command]
├── create      # Create a class
└── delete      # Delete a class
```

### Link Operations
```
gns3util link [command]
├── create      # Create a link
├── delete      # Delete a link
└── update      # Update a link
```

### Template Operations
```
gns3util template [command]
├── create      # Create a template
├── delete      # Delete a template
└── update      # Update a template
```

### User Operations
```
gns3util user [command]
├── create      # Create a user
├── delete      # Delete a user
└── update      # Update a user
```

### Group Operations
```
gns3util group [command]
├── create      # Create a group
├── delete      # Delete a group
└── update      # Update a group
```

### Role Operations
```
gns3util role [command]
├── create      # Create a role
├── delete      # Delete a role
└── update      # Update a role
```

### ACL Operations
```
gns3util acl [command]
├── create      # Create ACL entry
├── delete      # Delete ACL entry
└── update      # Update ACL entry
```

### Image Operations
```
gns3util image [command]
├── create      # Create an image
├── delete      # Delete an image
└── update      # Update an image
```

### Snapshot Operations
```
gns3util snapshot [command]
├── create      # Create a snapshot
├── delete      # Delete a snapshot
└── update      # Update a snapshot
```

### System Operations
```
gns3util system [command]
├── version     # Get GNS3 Server version
├── statistics  # Get server statistics
└── me          # Get current user info
```

### Remote Operations
```
gns3util remote [command]
├── install     # Install remote components
└── uninstall   # Uninstall remote components
```

### Cluster Operations
```
gns3util cluster [command]
├── create      # Create a cluster
├── add-node    # Add a single server to a cluster
├── add-nodes   # Add multiple servers in one run
├── config      # Manage the cluster config file
└── ls          # List clusters (optionally filtered by name)
```

#### Key Flags
- `cluster create`: `--name`, `--description`
- `cluster add-node` / `cluster add-nodes`: `--server`, `--user`, `--password`, `--weight` (default 5), `--max-groups` (default 3)
- `cluster config`: subcommands `edit`, `sync`, `apply`

### Sharing Operations
```
gns3util share [command]
├── send        # Transfer artifacts to an administrator over QUIC
└── receive     # Listen for incoming transfers
```

#### Key Flags
- `share send`: `--all`, `--send-config`, `--send-db`, `--send-key`, `--src-dir`, `--to`, `--discover-timeout`, `--yes`
- `share receive`: no additional flags

### Other Operations
```
gns3util appliance [command]    # Appliance operations
gns3util compute [command]      # Compute operations
gns3util drawing [command]      # Drawing operations
gns3util pool [command]         # Resource pool operations
gns3util symbol [command]       # Symbol operations
```

## Global Flags Usage

### Server Connection
```bash
# Required server flag
gns3util -s https://gns3-server.com:3080 project ls

# With insecure SSL
gns3util -s https://gns3-server.com:3080 -i project ls

# With keyfile
gns3util -s https://gns3-server.com:3080 -k ~/.gns3/gns3key project ls
```

### JSON Output
```bash
# Raw JSON output for scripting
gns3util -s https://gns3-server.com project ls --raw --no-color

# Process JSON with jq (disable color codes)
gns3util -s https://gns3-server.com project ls --raw --no-color | jq '.[] | .name'
```

> **Note:** Pipe-friendly tooling like `jq` requires disabling ANSI color codes. Use `--no-color` together with `--raw` whenever you plan to parse output programmatically.

### Help System
```bash
# General help
gns3util --help

# Command-specific help
gns3util project --help
gns3util node create --help
```

## Configuration Examples

### Environment Variables
```bash
export GNS3_SERVER="https://gns3-server.com:3080"
export GNS3_KEYFILE="~/.gns3/gns3key"
```

### Class JSON Example
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
    }
  ]
}
```

## Troubleshooting

### Common Issues
- **Authentication failed**: Check server URL and credentials
- **Project not found**: Verify project name and permissions
- **Node creation failed**: Check node type and compute resources
- **Template not found**: Ensure template exists on server

### Debug Commands
```bash
gns3util -s https://gns3-server.com system version
gns3util -s https://gns3-server.com system statistics
gns3util -s https://gns3-server.com auth status
```