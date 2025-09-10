# Quick Start

This guide will get you up and running with GNS3util in minutes.

## Step 1: Authentication

First, authenticate with your GNS3 server:

```bash
# Interactive login
./gns3util -s https://your-gns3-server:3080 auth login

# Or use a keyfile
./gns3util -s https://your-gns3-server:3080 -k ~/.gns3/gns3key
```

## Step 2: Create Your First Class

Create a class with student groups:

```bash
# Create class from JSON file
./gns3util -s https://server:3080 class create --file class.json

# Or create interactively
./gns3util -s https://server:3080 class create --interactive
```

Example class.json:
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

## Step 3: Create an Exercise with Template

### Using an Existing Project as Template
```bash
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "NetworkTemplate" \
  --confirm=false
```

### Interactive Template Selection
```bash
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --select-template
```

### Using a Template File
```bash
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "/path/to/template.gns3project"
```

## Step 4: Verify the Setup

Check that everything was created correctly:

```bash
# List classes
./gns3util -s https://server:3080 class ls

# List projects
./gns3util -s https://server:3080 project ls

# List nodes in a project
./gns3util -s https://server:3080 node ls "CS101-Lab1-Group1-<uuid>"
```

## Step 5: Test Student Access

Login as a student to verify they can only see their assigned projects:

```bash
# Login as student
./gns3util -s https://server:3080 -u student1 -p password123 auth login

# List projects (should only show student's projects)
./gns3util -s https://server:3080 project ls
```

## Next Steps

### Explore Example Scripts
```bash
# Run the test suite
./scripts/examples/test-all-scripts.sh https://server:3080

# Deploy a template-based exercise
./scripts/examples/deploy-template-exercise.sh \
  https://server:3080 \
  "CS101" \
  "Lab2" \
  "NetworkTemplate"
```

### Advanced Features
- [Template System](features/template-based-exercises.md)
- [Educational Workflows](features/educational-workflows.md)
- [Scripts and Automation](scripts/overview.md)

## Common Commands

### Project Management
```bash
# List projects
./gns3util -s https://server:3080 project ls

# Create project
./gns3util -s https://server:3080 project new --name "MyProject"

# Duplicate project
./gns3util -s https://server:3080 project duplicate "MyProject" --name "MyProjectCopy"
```

### Node Management
```bash
# List nodes
./gns3util -s https://server:3080 node ls "MyProject"

# Create node
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

## Troubleshooting

### Common Issues

#### Template Not Found
- Ensure the template project exists on the server
- Check the project name spelling
- Use `--select-template` for interactive selection

#### Permission Denied
- Verify student credentials
- Check ACL settings
- Ensure students are in the correct groups

#### Project Creation Failed
- Check server resources
- Verify template project is not corrupted
- Check server logs for detailed error messages

### Getting Help

- Use `--help` flag for command-specific help
- Check the [CLI Reference](cli-reference/commands.md)
- Review [example scripts](scripts/examples.md)
- Open an issue on GitHub
