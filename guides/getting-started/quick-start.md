# Quick Start

This guide walks through the minimum steps to connect `gns3util` to a GNS3v3 server, validate connectivity, and discover existing resources. Before continuing, complete the installation steps in `getting-started/installation.md` so the binary and prerequisites are ready.

## Step 1 · Authenticate

```bash
# Interactive login (recommended)
gns3util -s https://your-gns3-server:3080 auth login

# Or reuse an existing key file
gns3util -s https://your-gns3-server:3080 -k ~/.gns3/gns3key
```

If you plan to script commands with `jq` or other tooling, remember to add `--no-color --raw` so the output is easy to parse. See the `cli-reference/commands.md` page for all global flags.

## Step 2 · Discover Projects and Topology

Spend a minute exploring what already exists before you modify anything.

```bash
# List projects you can access
gns3util -s https://your-gns3-server:3080 project ls

# Inspect a project in detail (replace the name as needed)
gns3util -s https://your-gns3-server:3080 project info "Example"

# Enumerate project members such as nodes, links, and captures
gns3util -s https://your-gns3-server:3080 node ls "Example"
gns3util -s https://your-gns3-server:3080 snapshot ls "Example"
```

## Step 3 · Review Templates and Images

Templates define reusable topologies and appliances. Confirm what is available and gather metadata.

```bash
# Overview of template names
gns3util -s https://your-gns3-server:3080 template ls

# Detailed information for a specific template
gns3util -s https://your-gns3-server:3080 template info "NetworkTemplate"

# List registered images (disk or appliance files)
gns3util -s https://your-gns3-server:3080 image ls
```

If the server is empty or you need teaching templates, jump to `features/educational-workflows.md#template-management-workflow` for creation workflows.

## Step 4 · Check System Health and Connected Users

Confirm that the server is healthy before making changes.

```bash
# Validate server version and uptime details
gns3util -s https://your-gns3-server:3080 system version
gns3util -s https://your-gns3-server:3080 system statistics

# Audit current users and groups
gns3util -s https://your-gns3-server:3080 user ls
gns3util -s https://your-gns3-server:3080 group ls
gns3util -s https://your-gns3-server:3080 role ls
```

Once you understand the environment, you can proceed to provisioning tasks or automation.

## Step 5 · Explore Optional Workflows

Depending on your role, choose the next area to explore:

- **Education-focused labs**: See `features/educational-workflows.md` for class, exercise, and student management automation.
- **Remote operations**: `features/remote-operations.md` covers HTTPS setup, firewalls, and maintenance commands.
- **Sharing artifacts**: Exchange projects and configs with peers using the workflow in `features/sharing-and-collaboration.md`.
- **Automation scripts**: Review reusable shell scripts in `scripts/overview.md` and the full catalog in `scripts/examples.md`.

## Troubleshooting & Help

- Use `--help` on any command (for example `gns3util project --help`) to view available subcommands and flags.
- Consult the full CLI reference at `cli-reference/commands.md` for a tree of every command group.
- When piping output to other tools, run commands with `--no-color --raw` to avoid ANSI escape sequences.
- If you encounter authentication or connectivity issues, re-run Step 1 and confirm the server URL, credentials, and TLS requirements.
