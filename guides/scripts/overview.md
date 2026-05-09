# Scripts Overview

The `scripts/examples/` directory contains ready-to-use bash scripts for automating common GNS3util workflows.

## Available Scripts

### **deploy-template-exercise.sh**
Deploys an exercise using an existing project as a template.

**Usage:**
```bash
./scripts/examples/deploy-template-exercise.sh \
  <server_url> \
  <class_name> \
  <exercise_name> \
  <template_name>
```

**Example:**
```bash
./scripts/examples/deploy-template-exercise.sh \
  http://gns3-server:3080 \
  "CS101" \
  "Lab1" \
  "NetworkTemplate"
```

### **create-exercise-interactive.sh**
Creates an exercise with interactive template selection.

**Usage:**
```bash
./scripts/examples/create-exercise-interactive.sh \
  <server_url> \
  <class_name> \
  <exercise_name>
```

**Example:**
```bash
./scripts/examples/create-exercise-interactive.sh \
  http://gns3-server:3080 \
  "CS101" \
  "Lab1"
```

### **import-template-and-create-exercise.sh**
Creates an exercise using a template file.

**Usage:**
```bash
./scripts/examples/import-template-and-create-exercise.sh \
  <server_url> \
  <class_name> \
  <exercise_name> \
  <template_file>
```

**Example:**
```bash
./scripts/examples/import-template-and-create-exercise.sh \
  http://gns3-server:3080 \
  "CS101" \
  "Lab1" \
  "template.gns3project"
```

### **setup-class-lab.sh**
Creates individual lab projects for students with basic topology.

**Usage:**
```bash
./scripts/examples/setup-class-lab.sh \
  <server_url> \
  <num_students>
```

**Example:**
```bash
./scripts/examples/setup-class-lab.sh \
  http://gns3-server:3080 \
  5
```

### **cleanup-class.sh**
Cleans up projects with a specified prefix.

**Usage:**
```bash
./scripts/examples/cleanup-class.sh \
  <server_url> \
  <project_prefix>
```

**Example:**
```bash
./scripts/examples/cleanup-class.sh \
  http://gns3-server:3080 \
  "Student-"
```

### **test-all-scripts.sh**
Comprehensive validation script that runs all example scripts.

**Usage:**
```bash
./scripts/examples/test-all-scripts.sh <server_url>
```

**Example:**
```bash
./scripts/examples/test-all-scripts.sh http://gns3-server:3080
```

## Script Features

### **Common Features**
- All scripts use `#!/usr/bin/env bash` for better compatibility
- Server URL is passed as a command-line argument
- Authentication is expected to be pre-configured
- Error handling and validation included
- Colored output for better readability

### **Error Handling**
- Input validation
- Server connectivity checks
- Command execution verification
- Graceful error messages

### **Output Formatting**
- Colored success/error messages
- Progress indicators
- Detailed logging
- JSON output support

## Prerequisites

### **Authentication**
Before running scripts, ensure you're authenticated:

```bash
# Interactive login
gns3util -s https://server:3080 auth login

# Or use keyfile
gns3util -s https://server:3080 -k ~/.gns3/gns3key project ls
```

### **Dependencies**
- `gns3util` binary in PATH or current directory
- `jq` for JSON processing
- `grep`, `awk`, `tr` for text processing
- `curl` for HTTP requests (if needed)

### **Server Requirements**
- GNS3v3 server running
- Network connectivity
- Appropriate permissions for operations

## Usage Patterns

### **Educational Workflow**
```bash
# 1. Create class
gns3util -s https://server:3080 class create --file class.json

# 2. Deploy exercise with template
./scripts/examples/deploy-template-exercise.sh \
  https://server:3080 \
  "CS101" \
  "Lab1" \
  "NetworkTemplate"

# 3. Verify deployment
gns3util -s https://server:3080 project ls

# 4. Cleanup when done
./scripts/examples/cleanup-class.sh \
  https://server:3080 \
  "CS101-Lab1-"
```

### **Template Management**
```bash
# 1. Create template project
gns3util -s https://server:3080 project new --name "NetworkTemplate"

# 2. Configure template
gns3util -s https://server:3080 node create "NetworkTemplate" \
  --name "Router1" \
  --node-type "qemu" \
  --compute-id "local"

# 3. Use template in exercises
./scripts/examples/deploy-template-exercise.sh \
  https://server:3080 \
  "CS101" \
  "Lab1" \
  "NetworkTemplate"
```

### **Validation and Verification**
```bash
# Run comprehensive validation
./scripts/examples/test-all-scripts.sh https://server:3080

# Verify specific functionality
./scripts/examples/setup-class-lab.sh https://server:3080 3
```

## Customization

### **Modifying Scripts**
- Edit script parameters
- Add custom validation
- Extend functionality
- Add logging

### **Creating New Scripts**
- Follow existing patterns
- Include error handling
- Add help text
- Validate thoroughly

### **Integration**
- Use in CI/CD pipelines
- Integrate with monitoring
- Add to automation workflows
- Extend with additional tools

## Troubleshooting

### **Common Issues**

#### Script Not Found
```bash
# Ensure script is executable
chmod +x scripts/examples/*.sh

# Check script path
ls -la scripts/examples/
```

#### Permission Denied
```bash
# Fix permissions
chmod +x gns3util
chmod +x scripts/examples/*.sh
```

#### Authentication Failed
```bash
# Verify authentication
gns3util -s https://server:3080 project ls

# Check keyfile
cat ~/.gns3/gns3key
```

#### Server Connection Failed
```bash
# Verify connectivity
gns3util -s https://server:3080 --help

# Check server status
gns3util -s https://server:3080 project ls
```

### **Debug Mode**
```bash
# Enable verbose output
set -x
./scripts/examples/deploy-template-exercise.sh \
  https://server:3080 \
  "CS101" \
  "Lab1" \
  "NetworkTemplate"
set +x
```

### **Logging**
```bash
# Redirect output to log file
./scripts/examples/test-all-scripts.sh https://server:3080 2>&1 | tee test.log
```

## Best Practices

### **Script Development**
- Use descriptive variable names
- Include comprehensive error handling
- Add input validation
- Document script purpose and usage

### **Validation**
- Validate with different server configurations
- Check error conditions
- Verify edge cases
- Confirm output accuracy

### **Maintenance**
- Keep scripts up to date
- Monitor for deprecations
- Update documentation
- Version control changes

## Examples

### **Complete Workflow**
```bash
#!/bin/bash
# Complete educational workflow example

SERVER="https://gns3-server:3080"
CLASS="CS101"
EXERCISE="Lab1"
TEMPLATE="NetworkTemplate"

# 1. Create class
echo "Creating class..."
gns3util -s "$SERVER" class create --file class.json

# 2. Deploy exercise
echo "Deploying exercise..."
./scripts/examples/deploy-template-exercise.sh \
  "$SERVER" \
  "$CLASS" \
  "$EXERCISE" \
  "$TEMPLATE"

# 3. Verify deployment
echo "Verifying deployment..."
gns3util -s "$SERVER" project ls | grep "$CLASS-$EXERCISE"

# 4. Cleanup
echo "Cleaning up..."
./scripts/examples/cleanup-class.sh \
  "$SERVER" \
  "$CLASS-$EXERCISE-"
```

### **Template Management**
```bash
#!/bin/bash
# Template management example

SERVER="https://gns3-server:3080"
TEMPLATE="NetworkTemplate"

# Create template
echo "Creating template..."
gns3util -s "$SERVER" project new --name "$TEMPLATE"

# Configure template
echo "Configuring template..."
gns3util -s "$SERVER" node create "$TEMPLATE" \
  --name "Router1" \
  --node-type "qemu" \
  --compute-id "local"

# Export template
echo "Exporting template..."
gns3util -s "$SERVER" project export "$TEMPLATE" \
  --output "template.gns3project"
```
