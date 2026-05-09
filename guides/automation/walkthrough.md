# GNS3util Automation Walkthrough

This comprehensive walkthrough demonstrates how to automate GNS3 lab management using the gns3util tool and example scripts.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Template-Based Exercise Creation](#template-based-exercise-creation)
4. [Class Management Automation](#class-management-automation)
5. [Advanced Automation Scenarios](#advanced-automation-scenarios)
6. [Troubleshooting](#troubleshooting)
7. [Best Practices](#best-practices)

## Prerequisites

### System Requirements
- **GNS3v3 Server**: Running GNS3v3 server with API access
- **GNS3util CLI**: Built and available in your PATH
- **Dependencies**: `jq` for JSON processing, `bash` shell
- **Authentication**: Valid credentials for the GNS3 server

### Initial Setup
```bash
# Build gns3util
go build -o gns3util

# Make scripts executable
chmod +x scripts/examples/*.sh

# Test connection
./gns3util -s https://your-gns3-server:3080 system version
```

## Quick Start

### 1. Authentication
```bash
# Interactive login
./gns3util -s https://your-gns3-server:3080 auth login

# Or use keyfile
./gns3util -s https://your-gns3-server:3080 -k ~/.gns3/gns3key
```

### 2. Test All Scripts
```bash
# Run comprehensive test suite
./scripts/examples/test-all-scripts.sh https://your-gns3-server:3080
```

### 3. Create Your First Exercise
```bash
# Create a class
./gns3util -s https://your-gns3-server:3080 class create --interactive

# Create an exercise with template
./scripts/examples/deploy-template-exercise.sh \
  https://your-gns3-server:3080 \
  "CS101" \
  "Lab1" \
  "NetworkTemplate"
```

## Template-Based Exercise Creation

### Understanding Templates

Templates are the foundation of automated lab deployment. They allow you to:
- **Standardize lab environments** across all student groups
- **Reuse configurations** for multiple exercises
- **Ensure consistency** in network topologies and device configurations

### Template Types

#### 1. Server-Based Templates (Recommended)
Use existing projects already on the server as templates.

**Advantages:**
- Fastest deployment
- No file upload required
- Easy to manage and update
- Interactive selection available

**Example:**
```bash
# List available projects
./gns3util -s https://server:3080 project ls

# Use existing project as template
./scripts/examples/deploy-template-exercise.sh \
  https://server:3080 \
  "CS101" \
  "Lab1" \
  "NetworkTemplate"
```

#### 2. File-Based Templates
Import `.gns3project` files as templates.

**Advantages:**
- Easy to share and version control
- Can be stored in repositories
- Useful for backup and restore

**Example:**
```bash
# Create exercise from template file
./scripts/examples/import-template-and-create-exercise.sh \
  https://server:3080 \
  "CS101" \
  "Lab1" \
  "/path/to/template.gns3project"
```

#### 3. Interactive Template Selection
Use fuzzy finder to select from available templates.

**Advantages:**
- User-friendly interface
- No need to remember exact names
- Shows all available options

**Example:**
```bash
# Interactive template selection
./scripts/examples/create-exercise-interactive.sh \
  https://server:3080 \
  "CS101" \
  "Lab1"
```

### Creating Templates

#### Method 1: Manual Template Creation
```bash
# Create a new project
./gns3util -s https://server:3080 project new --name "NetworkTemplate" --auto-close true

# Add nodes to the template
./gns3util -s https://server:3080 node create "NetworkTemplate" \
  --name "Router1" \
  --node-type "qemu" \
  --compute-id "local"

./gns3util -s https://server:3080 node create "NetworkTemplate" \
  --name "Switch1" \
  --node-type "ethernet_switch" \
  --compute-id "local"

# Add more nodes as needed...
```

#### Method 2: Export Existing Project
```bash
# Export an existing project as template
./gns3util -s https://server:3080 project export "MyProject" --file "template.gns3project"
```

## Class Management Automation

### Creating Classes Programmatically

#### Using JSON Files
Create a class definition file:

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

Then create the class:
```bash
./gns3util -s https://server:3080 class create --file class.json
```

#### Using Interactive Mode
```bash
./gns3util -s https://server:3080 class create --interactive
```

### Managing Multiple Classes

#### List All Classes
```bash
./gns3util -s https://server:3080 class ls
```

#### Delete Classes
```bash
./gns3util -s https://server:3080 class delete --name "CS101" --confirm=false
```

#### Bulk Class Creation
Create a script to manage multiple classes:

```bash
#!/usr/bin/env bash
# create-multiple-classes.sh

SERVER="https://your-gns3-server:3080"

# Create classes for different courses
for course in "CS101" "CS102" "CS201"; do
  echo "Creating class: $course"
  
  # Create class JSON
  cat > "/tmp/${course}.json" << EOF
{
  "name": "$course",
  "groups": [
    {
      "name": "Group1",
      "students": [
        {"username": "${course,,}student1", "password": "password123"},
        {"username": "${course,,}student2", "password": "password123"}
      ]
    }
  ]
}
EOF
  
  # Create the class
  ./gns3util -s $SERVER class create --file "/tmp/${course}.json"
  
  # Clean up
  rm "/tmp/${course}.json"
done
```

## Advanced Automation Scenarios

### Scenario 1: Semester Setup

Automate the entire semester setup process:

```bash
#!/usr/bin/env bash
# semester-setup.sh

SERVER="https://your-gns3-server:3080"
SEMESTER="Fall2024"

# Create template projects
echo "Creating template projects..."
./gns3util -s $SERVER project new --name "BasicNetworkTemplate" --auto-close true
./gns3util -s $SERVER project new --name "AdvancedNetworkTemplate" --auto-close true

# Create classes
echo "Creating classes..."
for course in "CS101" "CS102" "CS201"; do
  ./gns3util -s $SERVER class create --interactive
done

# Deploy exercises
echo "Deploying exercises..."
./scripts/examples/deploy-template-exercise.sh $SERVER "CS101" "Lab1" "BasicNetworkTemplate"
./scripts/examples/deploy-template-exercise.sh $SERVER "CS102" "Lab1" "AdvancedNetworkTemplate"
```

### Scenario 2: Weekly Lab Rotation

Automate weekly lab rotations:

```bash
#!/usr/bin/env bash
# weekly-lab-rotation.sh

SERVER="https://your-gns3-server:3080"
CLASS="CS101"
WEEK=$1

if [ -z "$WEEK" ]; then
  echo "Usage: $0 <week_number>"
  exit 1
fi

# Clean up previous week's projects
echo "Cleaning up previous week's projects..."
./scripts/examples/cleanup-class.sh $SERVER "${CLASS}-Week$((WEEK-1))-"

# Deploy new week's lab
echo "Deploying week $WEEK lab..."
./scripts/examples/deploy-template-exercise.sh \
  $SERVER \
  "$CLASS" \
  "Week$WEEK" \
  "Week${WEEK}Template"
```

### Scenario 3: Individual Student Labs

Create individual lab environments for each student:

```bash
#!/usr/bin/env bash
# individual-student-labs.sh

SERVER="https://your-gns3-server:3080"
STUDENT_COUNT=20

echo "Creating individual lab projects for $STUDENT_COUNT students..."
./scripts/examples/setup-class-lab.sh $SERVER $STUDENT_COUNT

echo "Individual labs created successfully!"
echo "Students can access their labs at: https://your-gns3-server:3080"
```

### Scenario 4: Template Management

Automate template creation and management:

```bash
#!/usr/bin/env bash
# template-management.sh

SERVER="https://your-gns3-server:3080"
TEMPLATE_DIR="/path/to/templates"

# Create templates from files
for template_file in "$TEMPLATE_DIR"/*.gns3project; do
  if [ -f "$template_file" ]; then
    template_name=$(basename "$template_file" .gns3project)
    echo "Creating template: $template_name"
    
    # Import template
    ./gns3util -s $SERVER project import "$template_file" --name "$template_name"
  fi
done

# List all templates
echo "Available templates:"
./gns3util -s $SERVER project ls | grep "Template"
```

## Troubleshooting

### Common Issues

#### 1. Authentication Problems
```bash
# Check authentication status
./gns3util -s https://server:3080 auth status

# Re-authenticate if needed
./gns3util -s https://server:3080 auth login
```

#### 2. Template Not Found
```bash
# List available projects
./gns3util -s https://server:3080 project ls

# Check if template exists
./gns3util -s https://server:3080 project info "TemplateName"
```

#### 3. Class Creation Failed
```bash
# Check class JSON format
cat class.json | jq .

# Validate class structure
./gns3util -s https://server:3080 class create --file class.json --dry-run
```

#### 4. Node Creation Issues
```bash
# Check available node types
./gns3util -s https://server:3080 node ls "ProjectName"

# Verify compute IDs
./gns3util -s https://server:3080 compute ls
```

### Debug Commands

```bash
# Check server status
./gns3util -s https://server:3080 system version

# List all resources
./gns3util -s https://server:3080 project ls
./gns3util -s https://server:3080 class ls
./gns3util -s https://server:3080 user ls

# Check project details
./gns3util -s https://server:3080 project info "ProjectName"

# Check node details
./gns3util -s https://server:3080 node ls "ProjectName"
```

## Best Practices

### 1. Template Management
- **Use descriptive names**: `NetworkLabTemplate`, `SecurityLabTemplate`
- **Version control**: Keep template files in git repositories
- **Documentation**: Document what each template contains
- **Regular updates**: Keep templates up-to-date with course requirements

### 2. Class Organization
- **Consistent naming**: Use consistent naming conventions for classes
- **Group management**: Organize students into logical groups
- **Access control**: Ensure proper ACLs are set up
- **Cleanup**: Regular cleanup of old projects and classes

### 3. Automation Scripts
- **Error handling**: Always include error checking
- **Logging**: Log important operations
- **Idempotency**: Scripts should be safe to run multiple times
- **Documentation**: Document script parameters and usage

### 4. Security
- **Authentication**: Use secure authentication methods
- **Access control**: Limit student access to their assigned projects
- **Resource limits**: Set appropriate resource limits
- **Regular audits**: Regularly audit access and permissions

### 5. Performance
- **Batch operations**: Use batch operations when possible
- **Resource monitoring**: Monitor server resources
- **Cleanup**: Regular cleanup of unused resources
- **Optimization**: Optimize scripts for performance

## Example Workflows

### Complete Semester Setup
```bash
#!/usr/bin/env bash
# complete-semester-setup.sh

SERVER="https://your-gns3-server:3080"

# 1. Create template projects
echo "Step 1: Creating template projects..."
./gns3util -s $SERVER project new --name "BasicNetworkTemplate" --auto-close true
./gns3util -s $SERVER project new --name "AdvancedNetworkTemplate" --auto-close true

# 2. Create classes
echo "Step 2: Creating classes..."
for course in "CS101" "CS102" "CS201"; do
  ./gns3util -s $SERVER class create --interactive
done

# 3. Deploy exercises
echo "Step 3: Deploying exercises..."
./scripts/examples/deploy-template-exercise.sh $SERVER "CS101" "Lab1" "BasicNetworkTemplate"
./scripts/examples/deploy-template-exercise.sh $SERVER "CS102" "Lab1" "AdvancedNetworkTemplate"

# 4. Verify deployment
echo "Step 4: Verifying deployment..."
./gns3util -s $SERVER class ls
./gns3util -s $SERVER project ls

echo "Semester setup complete!"
```

### Daily Lab Management
```bash
#!/usr/bin/env bash
# daily-lab-management.sh

SERVER="https://your-gns3-server:3080"
DATE=$(date +%Y%m%d)

# 1. Clean up old projects
echo "Cleaning up old projects..."
./scripts/examples/cleanup-class.sh $SERVER "Old-"

# 2. Deploy today's labs
echo "Deploying today's labs..."
./scripts/examples/deploy-template-exercise.sh $SERVER "CS101" "Lab$DATE" "DailyTemplate"

# 3. Check system status
echo "Checking system status..."
./gns3util -s $SERVER system version

echo "Daily lab management complete!"
```

This walkthrough provides a comprehensive guide to automating GNS3 lab management using the gns3util tool and example scripts. The examples can be adapted to fit your specific educational environment and requirements.

