# CLI Examples

This section provides practical examples of using gns3util commands for common tasks.

## Authentication Examples

### Basic Login
```bash
# Interactive login
gns3util -s https://gns3-server.com:3080 auth login

# Check authentication status
gns3util -s https://gns3-server.com:3080 auth status
```

### Using Keyfile
```bash
# Login with keyfile
gns3util -s https://gns3-server.com:3080 -k ~/.gns3/gns3key project ls

# Save credentials to keyfile
gns3util -s https://gns3-server.com:3080 auth login --save-keyfile ~/.gns3/gns3key
```

## Project Management Examples

### Creating Projects
```bash
# Create a new project
gns3util -s https://gns3-server.com:3080 project new --name "MyLab"

# Create project with auto-close
gns3util -s https://gns3-server.com:3080 project new --name "TempLab" --auto-close true

# Duplicate existing project
gns3util -s https://gns3-server.com:3080 project duplicate "MyLab" --name "MyLabCopy"
```

### Listing Projects
```bash
# List all projects
gns3util -s https://gns3-server.com:3080 project ls

# Get project details
gns3util -s https://gns3-server.com:3080 project info "MyLab"

# List projects with JSON output
gns3util -s https://gns3-server.com:3080 project ls --raw
```

### Project Control
```bash
# Open project
gns3util -s https://gns3-server.com:3080 project open "MyLab"

# Close project
gns3util -s https://gns3-server.com:3080 project close "MyLab"

# Get project statistics
gns3util -s https://gns3-server.com:3080 project stats "MyLab"
```

## Node Management Examples

### Creating Nodes
```bash
# Create a single node
gns3util -s https://gns3-server.com:3080 node create \
  --project "MyLab" \
  --name "Router1" \
  --node-type "qemu" \
  --compute-id "local"

# Create multiple nodes
gns3util -s https://gns3-server.com:3080 node create \
  --project "MyLab" \
  --name "R1,R2,R3" \
  --node-type "qemu" \
  --compute-id "local"

# Create node from template
gns3util -s https://gns3-server.com:3080 node from-template \
  --project "MyLab" \
  --name "Router1" \
  --template "c7200"
```

### Node Operations
```bash
# List nodes in project
gns3util -s https://gns3-server.com:3080 node ls --project "MyLab"

# Get node details
gns3util -s https://gns3-server.com:3080 node info --project "MyLab" --name "Router1"

# Start all nodes
gns3util -s https://gns3-server.com:3080 node start-all --project "MyLab"

# Stop all nodes
gns3util -s https://gns3-server.com:3080 node stop-all --project "MyLab"
```

## Exercise Management Examples

### Creating Exercises
```bash
# Create exercise with template selection
gns3util -s https://gns3-server.com:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --select-template

# Create exercise with specific template
gns3util -s https://gns3-server.com:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "NetworkTemplate"

# Create exercise from file template
gns3util -s https://gns3-server.com:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "template.gns3project"
```

### Managing Classes
```bash
# Create class from JSON file
gns3util -s https://gns3-server.com:3080 class create --file class.json

# Delete class
gns3util -s https://gns3-server.com:3080 class delete --name "CS101" --confirm=false
```

## Template Management Examples

### Working with Templates
```bash
# List available templates
gns3util -s https://gns3-server.com:3080 template ls

# Get template details
gns3util -s https://gns3-server.com:3080 template info "c7200"

# Create new template
gns3util -s https://gns3-server.com:3080 template create \
  --name "MyTemplate" \
  --node-type "qemu"
```

### Project Import/Export
```bash
# Export project
gns3util -s https://gns3-server.com:3080 project export "MyLab" \
  --output "my-lab.gns3project"

# Import project
gns3util -s https://gns3-server.com:3080 project import \
  --file "my-lab.gns3project" \
  --name "ImportedLab"
```

## JSON Output Examples

### Processing JSON Data
```bash
# List projects as JSON
gns3util -s https://gns3-server.com:3080 project ls --raw

# Filter projects by name
gns3util -s https://gns3-server.com:3080 project ls --raw | \
  jq '.[] | select(.name | contains("Lab"))'

# Get project names only
gns3util -s https://gns3-server.com:3080 project ls --raw | \
  jq -r '.[] | .name'

# Count active projects
gns3util -s https://gns3-server.com:3080 project ls --raw | \
  jq '[.[] | select(.status == "opened")] | length'
```

### Node Information
```bash
# List nodes as JSON
gns3util -s https://gns3-server.com:3080 node ls --project "MyLab" --raw

# Get node details
gns3util -s https://gns3-server.com:3080 node info --project "MyLab" --name "Router1" --raw

# Filter by node type
gns3util -s https://gns3-server.com:3080 node ls --project "MyLab" --raw | \
  jq '.[] | select(.node_type == "qemu")'
```

## Bulk Operations Examples

### Creating Multiple Projects
```bash
#!/bin/bash
# Create projects for multiple students

for i in {1..10}; do
  gns3util -s https://gns3-server.com:3080 project new \
    --name "Student-$i-Lab" \
    --auto-close true
done
```

### Bulk Node Creation
```bash
#!/bin/bash
# Add standard nodes to all projects

PROJECTS=$(gns3util -s https://gns3-server.com:3080 project ls --raw | \
  jq -r '.[] | select(.name | contains("Student-")) | .name')

for project in $PROJECTS; do
  # Add routers
  gns3util -s https://gns3-server.com:3080 node create \
    --project "$project" \
    --name "R1,R2" \
    --node-type "qemu" \
    --compute-id "local"
  
  # Add switches
  gns3util -s https://gns3-server.com:3080 node create \
    --project "$project" \
    --name "SW1" \
    --node-type "ethernet_switch" \
    --compute-id "local"
done
```

### Cleanup Operations
```bash
#!/bin/bash
# Clean up all student projects

gns3util -s https://gns3-server.com:3080 project ls --raw | \
  jq -r '.[] | select(.name | contains("Student-")) | .name' | \
  while read project; do
    gns3util -s https://gns3-server.com:3080 project delete "$project" --confirm=false
  done
```

## Error Handling Examples

### Checking Command Success
```bash
#!/bin/bash
# Create project with error checking

if gns3util -s https://gns3-server.com:3080 project new --name "TestLab"; then
  echo "Project created successfully"
else
  echo "Failed to create project"
  exit 1
fi
```

### Validating Resources
```bash
#!/bin/bash
# Check if project exists before operations

PROJECT_NAME="MyLab"

if gns3util -s https://gns3-server.com:3080 project info "$PROJECT_NAME" >/dev/null 2>&1; then
  echo "Project $PROJECT_NAME exists"
else
  echo "Project $PROJECT_NAME not found"
  exit 1
fi
```

## Advanced Examples

### Server Health Check
```bash
#!/bin/bash
# Comprehensive server health check

echo "=== GNS3 Server Health Check ==="
echo "Server Version:"
gns3util -s https://gns3-server.com:3080 system version

echo -e "\nServer Statistics:"
gns3util -s https://gns3-server.com:3080 system statistics

echo -e "\nActive Projects:"
gns3util -s https://gns3-server.com:3080 project ls --raw | \
  jq '.[] | {name: .name, status: .status, nodes: .nodes | length}'

echo -e "\nAuthentication Status:"
gns3util -s https://gns3-server.com:3080 auth status
```

### Automated Lab Setup
```bash
#!/bin/bash
# Complete lab setup automation

CLASS_NAME="CS101"
EXERCISE_NAME="Lab1"
TEMPLATE_NAME="NetworkTemplate"

# Create class
echo "Creating class: $CLASS_NAME"
gns3util -s https://gns3-server.com:3080 class create --file class.json

# Deploy exercise
echo "Deploying exercise: $EXERCISE_NAME"
gns3util -s https://gns3-server.com:3080 exercise create \
  --class "$CLASS_NAME" \
  --exercise "$EXERCISE_NAME" \
  --template "$TEMPLATE_NAME" \
  --confirm=false

# Start all projects
echo "Starting all projects..."
gns3util -s https://gns3-server.com:3080 project ls --raw | \
  jq -r ".[] | select(.name | contains(\"$CLASS_NAME-$EXERCISE_NAME\")) | .name" | \
  while read project; do
    gns3util -s https://gns3-server.com:3080 project open "$project"
  done

echo "Lab setup complete!"
```
