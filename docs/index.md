# GNS3 API Utility

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

### **Developer Tools**
- **Example scripts**: Ready-to-use bash scripts for common workflows
- **Educational examples**: Step-by-step tutorials and use cases

## Quick Start

### Installation
```bash
# Build from source
go build -o gns3util

# Or use the pre-built binary
./gns3util --help
```

### Authentication
```bash
# Login to your GNS3 server
./gns3util -s https://your-gns3-server:3080 auth login

# Or use a keyfile
./gns3util -s https://your-gns3-server:3080 -k ~/.gns3/gns3key
```

### Basic Usage

#### Create a Class
```bash
# Create class from JSON file
./gns3util -s https://server:3080 class create --file class.json

# Interactive class creation
./gns3util -s https://server:3080 class create --interactive
```

#### Create an Exercise with Template
```bash
# Using existing project as template
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "NetworkTemplate" \
  --confirm=false

# Interactive template selection
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --select-template

# Using template file
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "/path/to/template.gns3project"
```

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

### **Validate All Scripts**
```bash
# Run comprehensive validation
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
./gns3util -s https://server:3080 project ls

# Create new project
./gns3util -s https://server:3080 project new --name "MyProject" --auto-close true

# Duplicate project
./gns3util -s https://server:3080 project duplicate "MyProject" --name "MyProjectCopy"
```

### Node Management
```bash
# List nodes in project
./gns3util -s https://server:3080 node ls "MyProject"

# Create nodes
./gns3util -s https://server:3080 node create "MyProject" \
  --name "Router1" \
  --node-type "qemu" \
  --compute-id "local"
```

### Class Management
```bash
# List classes
./gns3util -s https://server:3080 class ls

# Delete class
./gns3util -s https://server:3080 class delete --name "CS101" --confirm=false
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

### Validation
```bash
# Run example scripts validation
./scripts/examples/test-all-scripts.sh http://test-server:3080
```

### Contributing
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Update documentation for new functionality
5. Submit a pull request

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](../LICENSE) file for details.

## Support

For issues, questions, or contributions:
- Open an issue on GitHub
- Check the [online documentation](https://stefanistkuhl.github.io/gns3-api-util/)
- Review the [example scripts](scripts/examples/) for usage patterns
- Review the CLI help: `./gns3util --help`
# Updated Thu Sep 11 01:43:57 AM CEST 2025
