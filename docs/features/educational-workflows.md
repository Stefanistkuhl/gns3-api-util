# Educational Workflows

This section covers common educational workflows using gns3util for managing GNS3 lab environments in educational settings.

## Class Management Workflow

### 1. Create a Class
```bash
# Create class from JSON file
gns3util -s https://server:3080 class create --file class.json

# Interactive class creation
gns3util -s https://server:3080 class create --interactive
```

### 2. Deploy Exercises
```bash
# Using existing template
gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "NetworkTemplate" \
  --confirm=false

# Interactive template selection
gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --select-template
```

### 3. Monitor Student Progress
```bash
# List all projects for the class
gns3util -s https://server:3080 project ls --raw | jq '.[] | select(.name | contains("CS101-Lab1"))'

# Check specific student project
gns3util -s https://server:3080 project info "CS101-Lab1-Group1-<uuid>"
```

## Template Management Workflow

### 1. Create Template in GNS3 Web UI (Recommended)
1. **Open the normal GNS3 web interface** in your browser (usually at `http://your-gns3-server:3080`)
2. **Create a new project** with your desired name (e.g., "NetworkTemplate")
3. **Design your lab topology** using the normal GNS3 interface:
   - Drag and drop devices from the device panel
   - Connect devices with cables
   - Configure device settings
   - Test the topology by starting devices
4. **Save the project** - it's now ready to use as a template

### 2. Use Template with Fuzzy Picker
```bash
# Use interactive template selection (recommended)
gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --select-template

# The fuzzy picker will show all available projects
# Select your template from the list
```

### 3. Alternative: Use Template by Name
```bash
# Use template by exact name
gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "NetworkTemplate"
```

### 4. Export Template (Optional)
```bash
# Export for sharing or backup
gns3util -s https://server:3080 project export "NetworkTemplate" \
  --output "network-template.gns3project"
```

## Lab Management Workflow

### 1. Start All Labs
```bash
# Start all projects for a class
gns3util -s https://server:3080 project ls --raw | \
  jq -r '.[] | select(.name | contains("CS101")) | .name' | \
  while read project; do
    gns3util -s https://server:3080 project open "$project"
  done
```

### 2. Monitor Lab Status
```bash
# Check project status
gns3util -s https://server:3080 project ls --raw | \
  jq '.[] | {name: .name, status: .status}'
```

### 3. Create Snapshots
```bash
# Create snapshot for each student project
gns3util -s https://server:3080 project ls --raw | \
  jq -r '.[] | select(.name | contains("CS101")) | .name' | \
  while read project; do
    gns3util -s https://server:3080 snapshot create \
      --project "$project" \
      --name "Initial-Setup"
  done
```

## Cleanup Workflow

### 1. Clean Up After Class
```bash
# Delete all projects for a class
gns3util -s https://server:3080 project ls --raw | \
  jq -r '.[] | select(.name | contains("CS101")) | .name' | \
  while read project; do
    gns3util -s https://server:3080 project delete "$project" --confirm=false
  done
```

### 2. Clean Up Classes
```bash
# Delete class
gns3util -s https://server:3080 class delete --name "CS101" --confirm=false
```

## Advanced Workflows

### Bulk Project Creation
```bash
#!/bin/bash
# create-student-labs.sh

CLASS_NAME="CS101"
EXERCISE_NAME="Lab1"
STUDENT_COUNT=10

for i in $(seq 1 $STUDENT_COUNT); do
  PROJECT_NAME="Student-$i-$EXERCISE_NAME"
  
  # Create project
  gns3util -s https://server:3080 project new --name "$PROJECT_NAME"
  
  # Add basic nodes
  gns3util -s https://server:3080 node create "$PROJECT_NAME" \
    --name "Router1" \
    --node-type "qemu" \
    --compute-id "local"
  
  gns3util -s https://server:3080 node create "$PROJECT_NAME" \
    --name "PC1" \
    --node-type "vpcs" \
    --compute-id "local"
done
```

### Automated Exercise Deployment
```bash
#!/bin/bash
# deploy-exercise.sh

CLASS_NAME=$1
EXERCISE_NAME=$2
TEMPLATE_NAME=$3

# Create class if it doesn't exist
gns3util -s https://server:3080 class create --file class.json

# Deploy exercise
gns3util -s https://server:3080 exercise create \
  --class "$CLASS_NAME" \
  --exercise "$EXERCISE_NAME" \
  --template "$TEMPLATE_NAME" \
  --confirm=false

echo "Exercise '$EXERCISE_NAME' deployed for class '$CLASS_NAME'"
```

### Health Check Script
```bash
#!/bin/bash
# health-check.sh

echo "=== GNS3 Server Health Check ==="
echo "Server Version:"
gns3util -s https://server:3080 system version

echo -e "\nServer Statistics:"
gns3util -s https://server:3080 system statistics

echo -e "\nActive Projects:"
gns3util -s https://server:3080 project ls --raw | \
  jq '.[] | {name: .name, status: .status, nodes: .nodes | length}'

echo -e "\nAuthentication Status:"
gns3util -s https://server:3080 auth status
```

## Best Practices

### 1. Naming Conventions
- Use consistent naming: `{Class}-{Exercise}-{Group}-{UUID}`
- Include semester/year in class names: `CS101-Fall2024`
- Use descriptive exercise names: `Basic-Routing`, `VLAN-Configuration`

### 2. Resource Management
- Monitor server resources regularly
- Use resource pools for student access control
- Set appropriate project limits

### 3. Backup Strategy
- Create snapshots before major changes
- Export important templates
- Regular project backups

### 4. Security
- Use ACLs to restrict student access
- Regular password rotation
- Monitor user activities

## Troubleshooting

### Common Issues
- **Project not starting**: Check compute resources
- **Node creation failed**: Verify node types and images
- **Authentication issues**: Check server connectivity and credentials
- **Template not found**: Verify template exists and is accessible

### Debug Commands
```bash
# Check server status
gns3util -s https://server:3080 system version
gns3util -s https://server:3080 system statistics

# Check authentication
gns3util -s https://server:3080 auth status

# List all resources
gns3util -s https://server:3080 project ls --raw
gns3util -s https://server:3080 template ls --raw
```
