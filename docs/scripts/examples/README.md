# GNS3util Educational Scripts

This directory contains educational bash scripts for automating GNS3 lab management and exercise deployment with advanced template-based functionality.

## 🎯 Template-Based Exercise System

The scripts demonstrate a powerful template system that allows instructors to:
- **Use existing projects as templates** for consistent lab environments
- **Import template files** for sharing lab configurations
- **Automatically duplicate templates** for each student group
- **Manage access control** so students only see their assigned projects

## Scripts Overview

### 🚀 Template-Based Exercise Scripts

#### `deploy-template-exercise.sh`
Deploy an exercise using an existing project as a template. This is the **recommended approach** for most use cases.

**Usage:**
```bash
./deploy-template-exercise.sh <server_url> <class_name> <exercise_name> <template_project>
```

**Example:**
```bash
./deploy-template-exercise.sh http://gns3-server:3080 CS101 Lab1 NetworkTemplate
```

**What it does:**
1. Uses an existing project on the server as a template
2. Creates exercise projects for each student group
3. Duplicates the template for each group with proper naming
4. Sets up access control for students

#### `create-exercise-interactive.sh`
Create an exercise with interactive template selection using fuzzy finder. Perfect for when you want to choose from available templates.

**Usage:**
```bash
./create-exercise-interactive.sh <server_url> <class_name> <exercise_name>
```

**Example:**
```bash
./create-exercise-interactive.sh http://gns3-server:3080 CS101 Lab1
```

**What it does:**
1. Shows a fuzzy-finder interface to select from available projects
2. Uses the selected project as a template
3. Creates exercise projects for each student group
4. Provides interactive feedback during selection

#### `import-template-and-create-exercise.sh`
Create an exercise using a template file (.gns3project). Useful for sharing templates or when templates aren't on the server.

**Usage:**
```bash
./import-template-and-create-exercise.sh <server_url> <class_name> <exercise_name> <template_file>
```

**Example:**
```bash
./import-template-and-create-exercise.sh http://gns3-server:3080 CS101 Lab1 /path/to/template.gns3project
```

**What it does:**
1. Imports a .gns3project file to the server
2. Uses the imported project as a template
3. Creates exercise projects for each student group
4. Cleans up the temporary imported template

### Lab Management Scripts

#### `setup-class-lab.sh`
Create individual lab projects for students with basic network topology.

**Usage:**
```bash
./setup-class-lab.sh <server_url> [student_count]
```

**Example:**
```bash
./setup-class-lab.sh http://gns3-server:3080 10
```

#### `cleanup-class.sh`
Clean up projects with a specific prefix.

**Usage:**
```bash
./cleanup-class.sh <server_url> [project_prefix]
```

**Example:**
```bash
./cleanup-class.sh http://gns3-server:3080 Student-
```

### Testing Scripts

#### `test-all-scripts.sh`
Test all example scripts with comprehensive reporting.

**Usage:**
```bash
./test-all-scripts.sh <server_url>
```

**Example:**
```bash
./test-all-scripts.sh http://gns3-server:3080
```

## Prerequisites

1. **GNS3util CLI**: Ensure `gns3util` is built and available in the project root
2. **GNS3 Server**: Access to a running GNS3v3 server
3. **Authentication**: Valid credentials for the GNS3 server
4. **Dependencies**: 
   - `jq` for JSON processing (for cleanup script)
   - `bash` shell

## Quick Start

1. **Make scripts executable:**
   ```bash
   chmod +x scripts/examples/*.sh
   ```

2. **Test all scripts:**
   ```bash
   ./scripts/examples/test-all-scripts.sh http://your-gns3-server:3080
   ```

3. **Create a class and exercise:**
   ```bash
   ./scripts/examples/deploy-template-exercise.sh http://your-gns3-server:3080 CS101 Lab1 MyTemplate
   ```

## Features

### Template Support
- **Server-based templates**: Use existing projects as templates
- **File-based templates**: Import .gns3project files
- **Interactive selection**: Fuzzy finder for template selection
- **Automatic cleanup**: Template projects are cleaned up after duplication

### Class Management
- **JSON-based class creation**: Structured class definition
- **Multiple groups**: Support for multiple student groups
- **User management**: Automatic student account creation
- **Permission isolation**: Students can only access their assigned projects

### Lab Automation
- **Bulk operations**: Create multiple projects/nodes at once
- **Template duplication**: Identical lab setups for all students
- **Resource management**: Proper cleanup and resource allocation
- **Error handling**: Robust error checking and user feedback

## Educational Use Cases

### For Instructors
- **Quick lab setup**: Deploy identical labs for entire class
- **Template management**: Create reusable lab configurations
- **Class organization**: Manage multiple classes and exercises
- **Resource cleanup**: Automated cleanup after classes

### For Students
- **Isolated environments**: Each student gets their own lab instance
- **Consistent setup**: All students work with identical configurations
- **Self-service**: Students can start/stop their own labs
- **Progress tracking**: Individual project management

## Troubleshooting

### Common Issues

1. **Command not found**: Ensure `gns3util` is in the project root
2. **Authentication failed**: Check server URL and credentials
3. **Template not found**: Verify template project exists on server
4. **Permission denied**: Ensure proper ACLs are set up

### Debug Commands

```bash
# Check authentication
gns3util -s http://server:3080 auth status

# List available projects
gns3util -s http://server:3080 project ls

# Check server version
gns3util -s http://server:3080 system version
```

## Contributing

When adding new scripts:

1. Use `#!/usr/bin/env bash` shebang for compatibility
2. Include proper usage instructions and examples
3. Add error checking and user feedback
4. Follow the existing naming convention
5. Test scripts thoroughly before committing

## License

These scripts are part of the gns3util project and follow the same license terms.
