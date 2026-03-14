# Educational Workflows

This section walks through repeatable workflows for planning, deploying, and maintaining GNS3 lab environments with `gns3util` in educational settings. 

## Class Management Workflow

### Step 1 · Create a Class
```bash
# Create class from JSON file
gns3util -s https://server:3080 class create --file class.json

# Interactive class creation
gns3util -s https://server:3080 class create --interactive
```

> **Cluster scope note:** Classes are scoped to the endpoint you target. Use `-s/--server` for an individual server or `--cluster <name>` for a configured cluster entry—both ultimately hit a single cluster node. Clusters provide scheduling convenience but not aggregation: `gns3util class ls` only reports data for the nodes you reach, and a "single-node" cluster created via `--server` is still tracked in your local cluster config files.

### Step 2 · Deploy Exercises
```bash
# Recommended: reuse an existing template project
gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "NetworkTemplate" \
  --confirm=false

# Interactive template selection via fuzzy picker
gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --select-template

# Create empty exercises (omit template flag entirely)
gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab0"
```

> **Exercise listings:** `gns3util exercise ls` behaves just like class listing—results come from whichever server or cluster node you target. Plan class schedules so each group knows the exact server (or cluster node) they must log into.


## Template Management Workflow

### How Templates Work
- **Selection**: Choose an existing project on the server or import a `.gns3project` archive to act as the base topology.
- **Duplication**: Deploying an exercise clones the template for every group in the class.
- **Naming**: Projects follow `{Class}-{Exercise}-{Group}-{UUID}` to keep names unique.
- **Access control**: ACLs are applied so each student group sees only their assigned lab.

### Step 1 · Prepare a Template in the GNS3 Web UI (Recommended)
1. **Open the GNS3 web interface** (typically `http://your-gns3-server:3080`).
2. **Create a new project** with a clear name (for example, `NetworkTemplate`).
3. **Design and validate the topology** until it meets your teaching objectives.
4. **Save the project**—it is now ready to use as a template.

### Template Types
- **Server-based templates (recommended)**: Reuse a project already stored on the server. Fastest option and keeps artifacts local.
- **File-based templates**: Import a `.gns3project` archive. Useful for moving labs between environments, but confirm images exist at the destination first.
- **Cluster awareness**: Clusters schedule deployments, but they do **not** share UI sessions. Instructors and students still authenticate to each server URL individually.

### Step 2 · Select a Template
```bash
# Interactive fuzzy picker (shows all candidate projects)
gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --select-template

# Use template by exact project name
gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "NetworkTemplate"

# Import a template archive from disk
gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "template.gns3project"
```

### Step 3 · Validate Template Quality
```bash
# Inspect metadata and notes
gns3util -s https://server:3080 project info "NetworkTemplate"

# List nodes included in the template project
gns3util -s https://server:3080 node ls "NetworkTemplate"
```

### Step 4 · Update Templates Safely
```bash
# Rename or tweak a template project
gns3util -s https://server:3080 project update "NetworkTemplate" \
  --name "UpdatedNetworkTemplate"

# Recreate exercises so students receive the updated lab
gns3util -s https://server:3080 exercise delete --class "CS101" --exercise "Lab1"
gns3util -s https://server:3080 exercise create \
  --class "CS101" \
  --exercise "Lab1" \
  --template "UpdatedNetworkTemplate"
```

### Step 5 · Export Templates (Optional)
```bash
# Export for sharing or offline backup
gns3util -s https://server:3080 project export "NetworkTemplate" \
  --output "network-template.gns3project"
```

**Recommendation:** Within a single environment or cluster it is usually faster to duplicate the template project than to export/import it. Export is best for moving labs between isolated infrastructures. Destination hosts must already have the required appliance images—automatic image syncing across the cluster is not yet available.

> **Cluster limitation reminder:** Template deployment runs on the server you target with `-s/--server`. When working across a cluster, repeat the deployment on every node that hosts student groups or script a loop over server URLs.

### Troubleshooting Template Deployments
- **Template not found**: Run `gns3util project ls | grep -i template` or rely on `--select-template` to discover names.
- **Duplication failed**: Check server resources, confirm the template opens cleanly, and inspect server logs.
- **Access control issues**: Verify class membership and execute `gns3util -s https://server:3080 project acl <project>` to review ACLs.
