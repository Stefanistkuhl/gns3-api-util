# Cluster Management

The `gns3util cluster` command group makes it easy to coordinate multiple GNS3 servers and keep their configuration in sync. This page walks through the common workflows for building and maintaining a cluster.

## Prerequisites

- Access to each GNS3 server you plan to add to the cluster
- Valid credentials for the servers (set via flags or environment variables)
- A configured `gns3util` binary (see `getting-started/installation.md`)

## Create a Cluster

Give your cluster a name and optional description. Creation only needs to run once.

```bash
# Create a new cluster with a descriptive label
gns3util cluster create \
  --name lab-cluster \
  --description "Primary training cluster"
```

If you omit `--name`, the command prompts interactively. Re-run the command with the same name to update the description.

## Add Cluster Nodes

Use `cluster add-node` for one server or `cluster add-nodes` for multiple servers at once. Supply the server URLs and (optionally) scheduling weights.

```bash
# Add a single server with a lower scheduling weight
gns3util cluster add-node lab-cluster \
  --server https://gns3-host-a.example.com:3080 \
  --weight 3 \
  --user admin \
  --password "$GNS3_PASSWORD"

# Add several servers in one command (comma-separated list)
gns3util cluster add-nodes lab-cluster \
  --server https://gns3-host-b.example.com:3080,https://gns3-host-c.example.com:3080 \
  --weight 7 \
  --user admin \
  --password "$GNS3_PASSWORD"
```

**Tip:** Set `GNS3_USER` and `GNS3_PASSWORD` environment variables so you do not have to pass `--user` and `--password` every time.

### Control Group Capacity

Each node limits how many user groups it can host. Adjust the limit when adding nodes:

```bash
# Allow a node to host up to 5 student groups at once
gns3util cluster add-node lab-cluster \
  --server https://gns3-host-a.example.com:3080 \
  --max-groups 5
```

## Manage Cluster Configuration Files

`gns3util` keeps a local cluster configuration file that you can edit and apply. The config workflow uses three subcommands:

```bash
# Open the configuration in your $EDITOR (create it if necessary)
gns3util cluster config edit

# Align the config file with the current database state (writes to disk)
gns3util cluster config sync

# Apply the config on disk back into the local database
gns3util cluster config apply
```

This round-trip lets you version-control the cluster definition while ensuring the local database matches the file.

## Inspect Clusters

List all clusters or scope the listing to a single name:

```bash
# Show every cluster you can access
gns3util cluster ls

# Show detailed membership for one cluster
gns3util cluster ls lab-cluster --raw --no-color | jq '.'
```

`--raw` outputs JSON so you can pipe to `jq` or other tooling.

## Recommended Workflow

1. Create the cluster with `cluster create`.
2. Add servers using `cluster add-node` or `cluster add-nodes`, assigning weights and group limits.
3. Run `cluster config sync` to capture the state in a file.
4. Commit the configuration file to version control.
5. When changes are needed, edit the file, apply it with `cluster config apply`, and re-sync to confirm the database state.

By following this loop, you maintain a reproducible record of your cluster topology and can scale your GNS3 environment confidently.

> **Important limitation:** GNS3 clusters do not provide a single shared UI session. Administrators and students must log directly into each server URL they use. As a result, `gns3util class ls` or `exercise ls` only report items on the active server. Treat the cluster feature as a scheduling workaround: targeting `--cluster <name>` or `-s/--server` hits a single node at a time, and a one-node cluster created via `--server` is still stored in the local cluster config. Plan class scheduling so each group knows which server (or cluster node) they must connect to.
