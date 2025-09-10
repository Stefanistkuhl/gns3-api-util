# Template-Based Exercises

The template system allows you to create identical lab environments for multiple student groups by using a base project as a template.

## How Templates Work

### 1. Template Selection
Choose from existing projects on the server or import a `.gns3project` file.

### 2. Automatic Duplication
The template is duplicated for each student group in the class.

### 3. Project Naming
Projects are named using the format: `{{class}}-{{exercise}}-{{group}}-{{uuid}}`

### 4. Access Control
Students can only see and access their assigned projects.

## Template Types

### Server-Based Templates (Recommended)

Use existing projects already on the server:

```bash
# Use existing project as template
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "NetworkTemplate" \
  --confirm=false
```

**Advantages:**
- Fastest deployment
- No file upload required
- Interactive selection available
- No cleanup needed

### File-Based Templates

Import `.gns3project` files as templates:

```bash
# Import template file
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "/path/to/template.gns3project" \
  --confirm=false
```

**Advantages:**
- Useful for sharing templates
- Version control friendly
- Portable across servers
- Automatic cleanup after import

### Interactive Template Selection

Use the fuzzy picker to select from available projects:

```bash
# Interactive selection
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --select-template
```

**Features:**
- Search through available projects
- Preview project details
- Real-time filtering
- User-friendly interface

## Template Creation

### Creating a Template Project

1. **Design the Lab Environment**
   - Create nodes (routers, switches, PCs)
   - Configure connections
   - Set up initial configurations
   - Test the topology

2. **Save as Template**
   ```bash
   # Create a new project
   ./gns3util -s https://server:3080 project new --name "NetworkTemplate"
   
   # Add nodes and configure
   ./gns3util -s https://server:3080 node create "NetworkTemplate" \
     --name "Router1" \
     --node-type "qemu" \
     --compute-id "local"
   ```

3. **Export Template (Optional)**
   ```bash
   # Export project for sharing
   ./gns3util -s https://server:3080 project export "NetworkTemplate" \
     --output "template.gns3project"
   ```

### Template Best Practices

#### Node Naming
- Use descriptive names
- Avoid special characters
- Keep names consistent

#### Configuration
- Include initial configurations
- Set up basic connectivity
- Add placeholder configurations

#### Documentation
- Include README files
- Document lab objectives
- Provide setup instructions

## Exercise Creation Process

### 1. Class Setup
```bash
# Create class with groups
./gns3util -s https://server:3080 class create --file class.json
```

### 2. Template Selection
```bash
# Server-based template
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "NetworkTemplate"

# File-based template
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "template.gns3project"

# Interactive selection
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --select-template
```

### 3. Project Duplication
For each student group:
- Template is duplicated
- Project is renamed with group identifier
- ACL is configured for group access
- Project is started automatically

### 4. Access Control
- Students can only see their group's projects
- Projects are isolated by group
- No cross-group access allowed

## Template Management

### Listing Available Templates
```bash
# List all projects (potential templates)
./gns3util -s https://server:3080 project ls

# Search for specific templates
./gns3util -s https://server:3080 project ls | grep -i template
```

### Template Validation
```bash
# Check template project
./gns3util -s https://server:3080 project info "NetworkTemplate"

# List nodes in template
./gns3util -s https://server:3080 node ls "NetworkTemplate"
```

### Template Updates
```bash
# Update template project
./gns3util -s https://server:3080 project update "NetworkTemplate" \
  --name "UpdatedNetworkTemplate"

# Recreate exercises with updated template
./gns3util -s https://server:3080 exercise delete --class "CS101" --exercise "Lab1"
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "UpdatedNetworkTemplate"
```

## Advanced Features

### Template Inheritance
- Templates can reference other templates
- Hierarchical template structure
- Modular lab design

### Dynamic Configuration
- Environment-specific settings
- Variable substitution
- Conditional node creation

### Template Versioning
- Version control integration
- Template history tracking
- Rollback capabilities

## Troubleshooting

### Common Issues

#### Template Not Found
```bash
# Check if template exists
./gns3util -s https://server:3080 project ls | grep "TemplateName"

# Use interactive selection
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --select-template
```

#### Duplication Failed
- Check server resources
- Verify template project integrity
- Check server logs for errors

#### Access Control Issues
- Verify group membership
- Check ACL configuration
- Test student login

### Debug Commands

```bash
# Verbose output
./gns3util -s https://server:3080 --verbose exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "NetworkTemplate"

# Check project details
./gns3util -s https://server:3080 project info "CS101-Lab1-Group1-<uuid>"

# Verify ACL settings
./gns3util -s https://server:3080 project acl "CS101-Lab1-Group1-<uuid>"
```

## Examples

### Basic Template Exercise
```bash
# Create class
./gns3util -s https://server:3080 class create --file class.json

# Create exercise with template
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "NetworkTemplate" \
  --confirm=false
```

### Interactive Template Selection
```bash
# Create exercise with interactive template selection
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --select-template
```

### File-Based Template
```bash
# Create exercise from template file
./gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "template.gns3project" \
  --confirm=false
```

### Template Management
```bash
# List available templates
./gns3util -s https://server:3080 project ls | grep -i template

# Check template details
./gns3util -s https://server:3080 project info "NetworkTemplate"

# Export template
./gns3util -s https://server:3080 project export "NetworkTemplate" \
  --output "template.gns3project"
```
