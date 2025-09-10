# gns3util CLI Command Structure Reference

> This file serves as a reference for documentation.
> Commands marked with [GIF] have demonstration videos available.
> Commands marked with [EDUCATION] are prioritized for educational use cases.

## Root Command
```
gns3util [flags]
gns3util [command]
```

### Global Flags
- `-s, --server string`: GNS3v3 Server URL (required)
- `-k, --key-file string`: Set a location for a keyfile to use
- `-i, --insecure`: Ignore unsigned SSL-Certificates
- `--raw`: Output all data in raw json
- `-h, --help`: Help for gns3util

## Main Command Groups

### Authentication Commands
```
gns3util auth [flags] [command]
├── login       # Log in as user [GIF] [EDUCATION]
└── status      # Check the current status of your Authentication [GIF] [EDUCATION]
```

### Project Operations
```
gns3util project [command]
├── close         # Close a project [GIF] [EDUCATION]
├── delete        # Delete a project [GIF] [EDUCATION]
├── duplicate     # Duplicate a Project [GIF] [EDUCATION]
├── export        # Export a project from GNS3 [EDUCATION]
├── file          # Get a file from a project
├── import        # Import a project from a portable archive [EDUCATION]
├── info          # Get a project by id or name [GIF] [EDUCATION]
├── load          # Load a project from a given path
├── lock          # Lock all drawings and nodes in a project
├── locked        # Get if a project is locked by id or name
├── ls            # Get the projects of the GNS3 Server [GIF] [EDUCATION]
├── new           # Create a Project [GIF] [EDUCATION]
├── open          # Open a project
├── start-capture # Start a packet capture in a project on a given link
├── stats         # Get project-stats by id or name
├── unlock        # Unlock all drawings and nodes in a project
├── update        # Update a Project
└── write-file    # Write a file to a project
```

### Node Operations
```
gns3util node [command]
├── auto-idle-pc           # Get the auto-idle-pc of a node in a project by id or name
├── auto-idle-pc-proposals # Get the auto-idle-pc-proposals of a node in a project by id or name
├── console-reset          # Reset a console for a given node
├── create                 # Create a node in a project [GIF] [EDUCATION]
├── delete                 # Delete a node from a project [GIF] [EDUCATION]
├── duplicate              # Duplicate a Node in a Project [GIF] [EDUCATION]
├── from-template          # Create a node from a template [GIF] [EDUCATION]
├── info                   # Get a node in a project by name or id [GIF] [EDUCATION]
├── links                  # Get links of a given node in a project by id or name
├── ls                     # Get the nodes within a project by name or id [GIF] [EDUCATION]
├── node-file              # Get a file from a node
├── node-isolate           # Isolate a node (suspend all attached links)
├── node-unisolate         # Un-isolate a node (resume all attached suspended links)
├── reload-all             # Reload all nodes belonging to a project
├── start-all              # Start all nodes belonging to a project [GIF] [EDUCATION]
├── stop-all               # Stop all nodes belonging to a project [GIF] [EDUCATION]
├── suspend-all            # Suspend all nodes belonging to a project
└── update                 # Update a Node in a Project
```

### Exercise Operations
```
gns3util exercise [command]
├── create      # Create an exercise (project) for every group in a class with ACLs [EDUCATION]
└── delete      # Delete an exercise [EDUCATION]
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

### Other Operations
```
gns3util appliance [command]    # Appliance operations
gns3util compute [command]      # Compute operations
gns3util drawing [command]      # Drawing operations
gns3util pool [command]         # Resource pool operations
gns3util symbol [command]       # Symbol operations
```

## GIF Demonstration Placeholders

### Authentication Flow
```
[GIF: auth-login.gif]
- Shows interactive login process
- Demonstrates keyfile creation
- Shows authentication status check
```

### Project Management
```
[GIF: project-basics.gif]
- Creating a new project
- Listing projects with --raw output
- Project details and status
```

### Node Management
```
[GIF: node-creation.gif]
- Creating multiple nodes at once
- Using different node types
- Node configuration and status
```

### Template System
```
[GIF: template-system.gif]
- Creating exercises with templates
- Interactive template selection
- File-based template import
```

### Exercise Management
```
[GIF: exercise-creation.gif]
- Creating classes and exercises
- Template-based exercise deployment
- Student project access control
```

### Bulk Operations
```
[GIF: bulk-operations.gif]
- Creating multiple projects
- Bulk node creation
- Mass cleanup operations
```

### JSON Integration
```
[GIF: json-output.gif]
- Using --raw flag for JSON output
- Processing JSON with jq
- Scripting with JSON data
```

## Essential Commands for Education

### Authentication
```bash
gns3util -s https://gns3-server.com auth login
gns3util -s https://gns3-server.com auth status
```

### Project Management
```bash
gns3util -s https://gns3-server.com project new --name "Lab1"
gns3util -s https://gns3-server.com project ls --raw
gns3util -s https://gns3-server.com project info "Lab1"
```

### Node Management
```bash
gns3util -s https://gns3-server.com node create --project "Lab1" --name "R1" --node-type "qemu"
gns3util -s https://gns3-server.com node ls --project "Lab1"
gns3util -s https://gns3-server.com node start-all --project "Lab1"
```

### Exercise Management
```bash
gns3util -s https://gns3-server.com exercise create --class "CS101" --exercise "Lab1" --select-template
gns3util -s https://gns3-server.com exercise create --class "CS101" --exercise "Lab1" --template "NetworkTemplate"
```

### Class Management
```bash
gns3util -s https://gns3-server.com class create --file class.json
gns3util -s https://gns3-server.com class delete --name "CS101"
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
gns3util -s https://gns3-server.com project ls --raw

# Process JSON with jq
gns3util -s https://gns3-server.com project ls --raw | jq '.[] | .name'
```

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