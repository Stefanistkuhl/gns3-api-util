# GNS3util Scripts Walkthrough

This walkthrough provides detailed instructions for using the example scripts in the `scripts/examples/` directory to automate GNS3 lab management.

## Table of Contents

1. [Script Overview](#script-overview)
2. [Getting Started](#getting-started)
3. [Template-Based Exercise Scripts](#template-based-exercise-scripts)
4. [Lab Management Scripts](#lab-management-scripts)
5. [Testing and Validation](#testing-and-validation)
6. [Real-World Examples](#real-world-examples)
7. [Troubleshooting Scripts](#troubleshooting-scripts)

## Script Overview

The example scripts are designed to demonstrate the full capabilities of the gns3util tool in educational environments. Each script focuses on a specific aspect of lab management:

### Core Scripts
- **`deploy-template-exercise.sh`** - Deploy exercises using existing templates
- **`create-exercise-interactive.sh`** - Interactive template selection
- **`import-template-and-create-exercise.sh`** - File-based template import
- **`setup-class-lab.sh`** - Individual student lab setup
- **`cleanup-class.sh`** - Project cleanup and management
- **`test-all-scripts.sh`** - Comprehensive testing suite

## Getting Started

### Prerequisites
```bash
# Ensure gns3util is built and available
go build -o gns3util

# Make scripts executable
chmod +x scripts/examples/*.sh

# Verify dependencies
which jq  # Required for JSON processing
which bash  # Required shell
```

### Initial Setup
```bash
# Test connection to GNS3 server
./gns3util -s https://your-gns3-server:3080 system version

# Authenticate (if not using keyfile)
./gns3util -s https://your-gns3-server:3080 auth login
```

## Template-Based Exercise Scripts

### 1. Deploy Template Exercise (`deploy-template-exercise.sh`)

**Purpose**: Deploy an exercise using an existing project as a template.

**Usage**:
```bash
./deploy-template-exercise.sh <server_url> <class_name> <exercise_name> <template_project>
```

**Step-by-Step Example**:
```bash
# 1. First, create a template project
./gns3util -s https://server:3080 project new --name "NetworkTemplate" --auto-close true

# 2. Add nodes to the template
./gns3util -s https://server:3080 node create "NetworkTemplate" \
  --name "Router1" \
  --node-type "qemu" \
  --compute-id "local"

./gns3util -s https://server:3080 node create "NetworkTemplate" \
  --name "Switch1" \
  --node-type "ethernet_switch" \
  --compute-id "local"

# 3. Create a class
./gns3util -s https://server:3080 class create --interactive

# 4. Deploy the exercise
./deploy-template-exercise.sh \
  https://server:3080 \
  "CS101" \
  "Lab1" \
  "NetworkTemplate"
```

**What happens**:
1. Script uses the existing "NetworkTemplate" project
2. Creates exercise projects for each group in the class
3. Duplicates the template for each group
4. Sets up proper naming: `CS101-Lab1-Group1-<uuid>`
5. Configures access control for students

### 2. Interactive Template Selection (`create-exercise-interactive.sh`)

**Purpose**: Create an exercise with interactive template selection using fuzzy finder.

**Usage**:
```bash
./create-exercise-interactive.sh <server_url> <class_name> <exercise_name>
```

**Step-by-Step Example**:
```bash
# 1. Ensure you have multiple projects to choose from
./gns3util -s https://server:3080 project ls

# 2. Run the interactive script
./create-exercise-interactive.sh \
  https://server:3080 \
  "CS101" \
  "Lab2"

# 3. Select template from the fuzzy finder interface
# The script will show available projects and let you choose
```

**What happens**:
1. Script lists all available projects
2. Presents a fuzzy finder interface for selection
3. Uses the selected project as template
4. Creates exercise projects for each group
5. Provides feedback during the process

### 3. File-Based Template Import (`import-template-and-create-exercise.sh`)

**Purpose**: Create an exercise using a template file (.gns3project).

**Usage**:
```bash
./import-template-and-create-exercise.sh <server_url> <class_name> <exercise_name> <template_file>
```

**Step-by-Step Example**:
```bash
# 1. Export an existing project as template file
./gns3util -s https://server:3080 project export "MyProject" --file "template.gns3project"

# 2. Use the template file to create exercise
./import-template-and-create-exercise.sh \
  https://server:3080 \
  "CS101" \
  "Lab3" \
  "template.gns3project"
```

**What happens**:
1. Script imports the .gns3project file to the server
2. Uses the imported project as template
3. Creates exercise projects for each group
4. Cleans up the temporary imported template
5. Maintains the original template file

## Lab Management Scripts

### 4. Setup Class Lab (`setup-class-lab.sh`)

**Purpose**: Create individual lab projects for students with basic network topology.

**Usage**:
```bash
./setup-class-lab.sh <server_url> [student_count]
```

**Step-by-Step Example**:
```bash
# Create individual labs for 10 students
./setup-class-lab.sh https://server:3080 10
```

**What happens**:
1. Creates individual projects: `Student-1-Lab`, `Student-2-Lab`, etc.
2. Adds basic network topology to each project:
   - 2 QEMU routers (R1, R2)
   - 1 Ethernet switch (SW1)
   - 2 VPCs (PC1, PC2)
3. Configures auto-close for resource management
4. Provides project UUIDs for each student

**Customization**:
You can modify the script to add different node types or configurations:

```bash
# Edit the script to add more nodes
./gns3util -s $GNS3_SERVER node create "$PROJECT_UUID" \
  --name "Server1" \
  --node-type "qemu" \
  --compute-id "local"
```

### 5. Cleanup Class (`cleanup-class.sh`)

**Purpose**: Clean up projects with a specific prefix.

**Usage**:
```bash
./cleanup-class.sh <server_url> [project_prefix]
```

**Step-by-Step Example**:
```bash
# Clean up all student lab projects
./cleanup-class.sh https://server:3080 Student-

# Clean up specific exercise projects
./cleanup-class.sh https://server:3080 CS101-Lab1-
```

**What happens**:
1. Lists all projects on the server
2. Filters projects matching the prefix
3. Deletes each matching project
4. Provides feedback on deletion status

## Testing and Validation

### 6. Test All Scripts (`test-all-scripts.sh`)

**Purpose**: Comprehensive testing suite for all example scripts.

**Usage**:
```bash
./test-all-scripts.sh <server_url>
```

**What it tests**:
1. **Pre-test cleanup** - Removes any existing test data
2. **Template creation** - Creates test template projects
3. **Class creation** - Creates test classes with groups
4. **Exercise deployment** - Tests all template-based scripts
5. **Lab setup** - Tests individual lab creation
6. **Cleanup** - Tests cleanup functionality
7. **Post-test cleanup** - Removes all test data

**Example Output**:
```
=== Testing All GNS3util Example Scripts ===
Server: https://server:3080
Test Prefix: TestScript1234567890

--- Pre-test cleanup ---
--- Creating test template project ---
Template project created: TestTemplate1234567890

--- Creating test classes ---
Test classes created successfully

=== Test 1: Template-based Exercise Deployment ===
✅ deploy-template-exercise.sh: PASSED

=== Test 2: Interactive Template Selection ===
✅ create-exercise-interactive.sh: PASSED

=== Test 3: File-based Template Import ===
✅ import-template-and-create-exercise.sh: PASSED

=== Test 4: Setup Class Lab ===
✅ setup-class-lab.sh: PASSED

=== Test 5: Cleanup Script ===
✅ cleanup-class.sh: PASSED

=== Test Summary ===
All script tests completed!
```

## Real-World Examples

### Example 1: Complete Course Setup

**Scenario**: Setting up a complete networking course with multiple labs.

```bash
#!/usr/bin/env bash
# course-setup.sh

SERVER="https://your-gns3-server:3080"
COURSE="CS101"

echo "Setting up $COURSE networking course..."

# 1. Create template projects
echo "Creating template projects..."
./gns3util -s $SERVER project new --name "BasicNetworkTemplate" --auto-close true
./gns3util -s $SERVER project new --name "AdvancedNetworkTemplate" --auto-close true

# 2. Add nodes to templates
echo "Configuring basic network template..."
./gns3util -s $SERVER node create "BasicNetworkTemplate" \
  --name "Router1" --node-type "qemu" --compute-id "local"
./gns3util -s $SERVER node create "BasicNetworkTemplate" \
  --name "Switch1" --node-type "ethernet_switch" --compute-id "local"

echo "Configuring advanced network template..."
./gns3util -s $SERVER node create "AdvancedNetworkTemplate" \
  --name "Router1" --node-type "qemu" --compute-id "local"
./gns3util -s $SERVER node create "AdvancedNetworkTemplate" \
  --name "Router2" --node-type "qemu" --compute-id "local"
./gns3util -s $SERVER node create "AdvancedNetworkTemplate" \
  --name "Switch1" --node-type "ethernet_switch" --compute-id "local"

# 3. Create class
echo "Creating class..."
./gns3util -s $SERVER class create --interactive

# 4. Deploy labs
echo "Deploying labs..."
./deploy-template-exercise.sh $SERVER $COURSE "Lab1" "BasicNetworkTemplate"
./deploy-template-exercise.sh $SERVER $COURSE "Lab2" "AdvancedNetworkTemplate"

echo "Course setup complete!"
```

### Example 2: Weekly Lab Rotation

**Scenario**: Automating weekly lab rotations for a semester.

```bash
#!/usr/bin/env bash
# weekly-rotation.sh

SERVER="https://your-gns3-server:3080"
CLASS="CS101"
WEEK=$1

if [ -z "$WEEK" ]; then
  echo "Usage: $0 <week_number>"
  exit 1
fi

echo "Setting up week $WEEK for $CLASS..."

# 1. Clean up previous week
echo "Cleaning up previous week..."
./cleanup-class.sh $SERVER "${CLASS}-Week$((WEEK-1))-"

# 2. Deploy current week
echo "Deploying week $WEEK lab..."
./deploy-template-exercise.sh \
  $SERVER \
  "$CLASS" \
  "Week$WEEK" \
  "Week${WEEK}Template"

# 3. Verify deployment
echo "Verifying deployment..."
./gns3util -s $SERVER project ls | grep "$CLASS-Week$WEEK"

echo "Week $WEEK setup complete!"
```

### Example 3: Individual Student Projects

**Scenario**: Creating individual lab environments for each student.

```bash
#!/usr/bin/env bash
# individual-labs.sh

SERVER="https://your-gns3-server:3080"
STUDENT_COUNT=25

echo "Creating individual labs for $STUDENT_COUNT students..."

# Create individual lab projects
./setup-class-lab.sh $SERVER $STUDENT_COUNT

# Verify creation
echo "Verifying lab creation..."
./gns3util -s $SERVER project ls | grep "Student-" | wc -l

echo "Individual labs created successfully!"
echo "Students can access their labs at: https://your-gns3-server:3080"
```

## Troubleshooting Scripts

### Common Issues and Solutions

#### 1. Script Permission Errors
```bash
# Make scripts executable
chmod +x scripts/examples/*.sh

# Check script permissions
ls -la scripts/examples/
```

#### 2. Authentication Issues
```bash
# Check authentication status
./gns3util -s https://server:3080 auth status

# Re-authenticate if needed
./gns3util -s https://server:3080 auth login
```

#### 3. Template Not Found
```bash
# List available projects
./gns3util -s https://server:3080 project ls

# Check specific project
./gns3util -s https://server:3080 project info "TemplateName"
```

#### 4. Class Creation Failed
```bash
# Check class JSON format
cat class.json | jq .

# Validate with dry run
./gns3util -s https://server:3080 class create --file class.json --dry-run
```

#### 5. Node Creation Issues
```bash
# Check available node types
./gns3util -s https://server:3080 node ls "ProjectName"

# Verify compute IDs
./gns3util -s https://server:3080 compute ls
```

### Debug Mode

Enable debug mode for detailed output:

```bash
# Run scripts with debug output
bash -x ./scripts/examples/deploy-template-exercise.sh \
  https://server:3080 \
  "CS101" \
  "Lab1" \
  "Template"
```

### Logging

Add logging to scripts for better debugging:

```bash
#!/usr/bin/env bash
# Add to any script

# Enable logging
set -e  # Exit on error
set -u  # Exit on undefined variable

# Log function
log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a script.log
}

# Use logging
log "Starting script execution"
log "Creating project: $PROJECT_NAME"
```

## Best Practices

### 1. Script Organization
- **Use descriptive names**: `deploy-weekly-labs.sh`, `cleanup-old-projects.sh`
- **Include usage information**: Always include usage examples
- **Error handling**: Check for errors and provide meaningful messages
- **Logging**: Log important operations for debugging

### 2. Template Management
- **Consistent naming**: Use consistent naming conventions
- **Version control**: Keep template files in version control
- **Documentation**: Document what each template contains
- **Regular updates**: Keep templates up-to-date

### 3. Class Management
- **Structured data**: Use JSON files for class definitions
- **Validation**: Validate class data before creation
- **Cleanup**: Regular cleanup of old classes and projects
- **Access control**: Ensure proper ACLs are set up

### 4. Testing
- **Test regularly**: Run test scripts regularly
- **Validate results**: Check that scripts produce expected results
- **Document issues**: Document any issues and solutions
- **Update tests**: Keep test scripts up-to-date

This walkthrough provides comprehensive guidance for using the example scripts effectively in educational environments. The scripts are designed to be robust, well-documented, and easy to customize for specific needs.

