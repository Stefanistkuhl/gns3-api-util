# GNS3 API Util

<p align="center">
  <img width=256 src="https://i.imgur.com/t1PNyl4.gif" alt="surely a temporary logo" />
</p>

[![PyPI Version](https://img.shields.io/pypi/v/gns3util)](https://pypi.org/project/gns3util/)
![Python Version](https://img.shields.io/python/required-version-toml?tomlFilePath=https%3A%2F%2Fraw.githubusercontent.com%2FStefanistkuhl%2Fgns3-api-util%2Frefs%2Fheads%2Fmaster%2Fpyproject.toml)
![GitHub Issues or Pull Requests by label](https://img.shields.io/github/issues/stefanistkuhl/gns3-api-util)
![language count](https://img.shields.io/github/languages/count/stefanistkuhl/gns3-api-util)
![repo size](https://img.shields.io/github/repo-size/stefanistkuhl/gns3-api-util)
![GitHub License](https://img.shields.io/github/license/stefanistkuhl/gns3-api-util)

A command-line utility for interacting with the GNS3 API, designed especially for setting up educational environments. This tool streamlines common API operations—such as authentication, GET, POST, and DELETE requests—against a GNS3 server, making it easier to automate and manage classes, exercises, users, groups, and more in your network emulation labs.

## Table of Contents

- [Summary](#summary)
- [Features](#features)
- [Installation](#installation)
  - [Using uv (Recommended)](#using-uv-recommended)
  - [From PyPI](#from-pypi)
  - [From Source](#from-source)
- [Usage](#usage)
  - [Authentication](#authentication)
  - [Running the CLI](#running-the-cli)
  - [Main Commands](#main-commands)
  - [Educational Workflow](#educational-workflow)
    - [Create a Class (with Students)](#1-create-a-class-with-students)
    - [Exercise Management](#2-exercise-management)
    - [Delete a Class or Exercise](#3-delete-a-class-or-exercise)
    - [Interactive Fuzzy Search](#4-interactive-fuzzy-search)
  - [Help](#help)
- [Command Reference](#command-reference)

## Features

- **Educational Environment Setup**: Easily create and manage classes and exercises for students.
- **Interactive CLI**: Use fuzzy search and interactive prompts for selecting users, groups, and more.
- **Automation**: Scriptable for bulk operations and integration into CI/CD or lab provisioning.
- **Web-based Class Generator**: Launch a local web UI to generate class JSON for bulk student creation.
- **Comprehensive API Coverage**: Supports most GNS3 v3 API endpoints for GET, POST, DELETE.
- **Remote SSL Setup**: Install and configure HTTPS (via Caddy reverse proxy) on your GNS3 server remotely over SSH with a single command.

## Installation

### Using uv (Recommended)

[uv](https://github.com/astral-sh/uv) is a fast Python package manager and virtual environment tool.

1. **Install uv** (if not already installed):

   ```bash
   pip install uv
   ```

2. **Create a virtual environment and install gns3util**:

   ```bash
   uv venv venv
   source venv/bin/activate
   uv pip install gns3util
   ```

   Or, for user installs (no root required):

   ```bash
   uv pip install --user gns3util
   ```

### From PyPI

Install the latest release directly from [PyPI](https://pypi.org/project/gns3util/):

```bash
pip install gns3util
```

Or, for user installs (no root required):

```bash
pip install --user gns3util
```

### From Source

1. **Clone the Repository**

   ```bash
   git clone https://github.com/Stefanistkuhl/gns3-api-util.git
   cd gns3-api-util
   ```

2. **(Optional) Create a Virtual Environment**

   Using uv:

   ```bash
   uv venv venv
   source venv/bin/activate
   ```

   Or with standard venv:

   ```bash
   python -m venv venv
   source venv/bin/activate
   ```

3. **Install the Project**

   ```bash
   pip install -e .
   ```

### Dependencies

The following Python packages are required and will be installed automatically:

- [click](https://click.palletsprojects.com/)
- [requests](https://docs.python-requests.org/)
- [rich](https://github.com/Textualize/rich)
- [bottle](https://github.com/bottlepy/bottle)
- [InquirerPy](https://github.com/kazhala/InquirerPy)

For enhanced interactive selection, you can also install [fzf](https://github.com/junegunn/fzf) (optional):

```bash
# On Ubuntu/Debian
sudo apt install fzf

# Or via Homebrew (macOS/Linux)
brew install fzf

# Or via winget (Windows)
winget install --id=junegunn.fzf
```

## Usage

After installing, the utility can be executed directly from the command line using the entry point `gns3util`.

### Authentication

You can authenticate to the GNS3 server using several methods:

- **Flags**:  
  Use `--username` and `--password` (or `-u` and `-p`) to provide credentials directly.
  ```bash
  gns3util --server http://localhost:3080 --username admin --password secret ...
  ```
- **Environment Variables**:  
  Set `GNS3_USERNAME` and `GNS3_PASSWORD` in your environment.
  ```bash
  export GNS3_USERNAME=admin
  export GNS3_PASSWORD=secret
  gns3util --server http://localhost:3080 ...
  ```
- **Standard Input**:  
  If you omit the password, you will be prompted securely for it via stdin.

### Running the CLI

At a minimum, provide the `--server` (or `-s`) option with the URL of your GNS3 server:

```bash
gns3util --server http://<GNS3_SERVER_ADDRESS>
```

### Install HTTPS/SSL on GNS3 Server (Remote)

You can set up HTTPS for your GNS3 server using the built-in `install ssl` command, which connects to your server via SSH and installs/configures [Caddy](https://caddyserver.com/) as a reverse proxy. This enables secure access to your GNS3 server with SSL.

#### Usage

```bash
gns3util --server http://<GNS3_SERVER_ADDRESS> install ssl <user> [OPTIONS]
```

- `<user>`: The SSH username for the remote server.

#### Options

- `-p, --port INTEGER`  
  SSH port (default: 22).

- `-k, --key PATH`  
  Path to a custom SSH private key file.

- `-rp, --reverse-proxy-port INTEGER`  
  Port for the reverse proxy to use (default: 443).

- `-gp, --gns3-port INTEGER`  
  Port of the GNS3 Server (default: 3080).

- `-d, --domain TEXT`  
  Domain to use for the reverse proxy (default: none).

- `-s, --subject TEXT`  
  Subject for the SSL certificate (default: `/CN=localhost`).  
  Format: `/C=COUNTRY/ST=STATE/L=CITY/O=ORG/OU=UNIT/CN=NAME/emailAddress=EMAIL`

- `-fa, --firewall-allow TEXT`  
  Allow only a given subnet to access the GNS3 server port (e.g., `10.0.0.0/24`).  
  Requires `--firewall-block`.

- `-fb, --firewall-block`  
  Block all connections to the GNS3 server port except those allowed by `--firewall-allow`.

- `-i, --interactive`  
  Edit all options interactively in your editor before running.

- `-v, --verbose`  
  Enable extra logging.

#### Example

```bash
gns3util --server http://10.10.10.10 install ssl ubuntu -k ~/.ssh/id_ed25519 -d mylab.example.com -fa 10.0.0.0/24 -fb
```

This will:
- Connect to `10.10.10.10` as user `ubuntu` using the specified SSH key.
- Set up Caddy as a reverse proxy on port 443 for the GNS3 server on port 3080.
- Use `mylab.example.com` as the domain.
- Restrict GNS3 server port access to the `10.0.0.0/24` subnet.

#### Interactive Mode

To edit all options in your editor before running:

```bash
gns3util --server http://10.10.10.10 install ssl ubuntu --interactive
```

#### Script Location

The script that is pushed and executed on the remote server can be found in the source tree at:

```
src/gns3util/resources/setup_https.sh
```

You can review or modify this script as needed before running the command.

#### Requirements & Notes

- The remote user must be `root` or have passwordless `sudo` privileges.
- Only works on Linux servers.
- The command will attempt to use SSH keys from `~/.ssh` or a custom path, and will fall back to password authentication if needed.

### Main Commands

The CLI supports several subcommands to interact with the GNS3 API:

- **auth**: Manage authentication (login, status).
- **get**: Perform GET requests (list users, groups, projects, etc).
- **post**: Perform POST requests (create resources, run actions).
- **delete**: Perform DELETE requests (remove users, groups, projects, etc).
- **create**: High-level creation commands for classes, exercises, and more.
- **add**: Add members/resources to groups or pools.
- **update**: Update resources with new data.
- **fuzzy**: Interactive fuzzy-finder commands for selecting users/groups.
- **stream**: Stream notifications from the GNS3 server.

### Educational Workflow

#### 1. Create a Class (with Students)

You can launch a local web UI to generate a class JSON and bulk-create students:

```bash
gns3util --server http://localhost:3080 create class --create
```

Or use an existing JSON file:

```bash
gns3util --server http://localhost:3080 create class path/to/class.json
```

#### 2. Exercise Management

Create an exercise for a class:

```bash
gns3util --server http://localhost:3080 create exercise <class_name> <exercise_name>
```

Delete an exercise interactively:

```bash
gns3util --server http://localhost:3080 delete exercise
```

**Non-interactive deletion for exercises:**  
Use the correct flags as defined in the CLI:

- `-n, --non_interactive <value>`: Run the command non-interactively.
- `-c, --set_class <class>`: Set the class from which to delete the exercise.
- `-g, --set_group <group>`: Set the group from which to delete the exercise.
- `-a, --delete_all`: Delete all exercises.

Example:

```bash
gns3util --server http://localhost:3080 delete exercise --non_interactive yes --set_class MyClass --set_group MyGroup
```

Delete all exercises in a class:

```bash
gns3util --server http://localhost:3080 delete exercise --set_class MyClass --set_group MyGroup --delete_all
```

#### 3. Delete a Class or Exercise

Delete a class and all its students interactively:

```bash
gns3util --server http://localhost:3080 delete class
```

**Non-interactive deletion:**  
You can delete a class directly by specifying its name with the `--non_interactive` flag:

```bash
gns3util --server http://localhost:3080 delete class --non_interactive MyClass
```

**Delete all classes:**  
To delete all classes non-interactively, use the `--delete_all` flag:

```bash
gns3util --server http://localhost:3080 delete class --delete_all
```

**Delete all exercises of a class:**  
To delete all exercises of a class when deleting the class, use the `--delete_exercises` flag:

```bash
gns3util --server http://localhost:3080 delete class --non_interactive MyClass --delete_exercises
```

#### 4. Interactive Fuzzy Search

Find users or groups with fuzzy search:

```bash
gns3util --server http://localhost:3080 fuzzy user-info
gns3util --server http://localhost:3080 fuzzy group-info
```

### Help

You can view the help text by using the `--help` option:

```bash
gns3util --help
```

Or for any subcommand:

```bash
gns3util <subcommand> --help
```

This will display usage information and options for each command.

---

## Command Reference

Below is a list of all subcommands grouped by their command groups:

<details>
<summary><strong>auth</strong> (Authentication)</summary>

- `login`
- `status`
</details>

<details>
<summary><strong>get</strong> (GET API endpoints)</summary>

- `version`
- `iou-license`
- `statistics`
- `me`
- `users`
- `projects`
- `groups`
- `roles`
- `privileges`
- `acl-endpoints`
- `acl`
- `templates`
- `symbols`
- `images`
- `default-symbols`
- `computes`
- `appliances`
- `pools`
- `user`
- `user-groups`
- `project`
- `project-stats`
- `project-locked`
- `group`
- `group-members`
- `role`
- `role-privileges`
- `template`
- `compute`
- `docker-images`
- `virtualbox-vms`
- `vmware-vms`
- `images_by_path`
- `snapshots`
- `appliance`
- `pool`
- `pool-resources`
- `drawings`
- `symbol`
- `acl-rule`
- `links`
- `nodes`
- `node`
- `node-links`
- `link`
- `link-filters`
- `drawing`
- `usernames-and-ids`
- `uai`
</details>

<details>
<summary><strong>post</strong> (POST API endpoints)</summary>

- `controller reload`
- `controller shutdown`
- `image install_img`
- `project duplicate_template`
- `project project_close`
- `project project_open`
- `project project_lock`
- `project project_unlock`
- `node start_nodes`
- `node stop_nodes`
- `node suspend_nodes`
- `node reload_nodes`
- `node nodes_console_reset`
- `compute connect_compute`
- `image upload_img`
- `project project_import`
- `project project_write_file`
- `node node_isolate`
- `node node_unisolate`
- `node node_console_reset`
- `link reset_link`
- `link stop_link_capture`
- `snapshot snapshot_restore`
- `image add_applience_version`
- `node duplicate_node`
- `link start_link_capture`
- `class`
- `exercise`
- `check_version`
- `user_authenticate`
</details>

<details>
<summary><strong>delete</strong> (DELETE API endpoints)</summary>

- `prune_images`
- `user`
- `compute`
- `project`
- `template`
- `image`
- `acl`
- `role`
- `group`
- `pool`
- `pool_resource`
- `link`
- `node`
- `drawing`
- `role_priv`
- `user_from_group`
- `snapshot`
- `class`
- `exercise`
</details>

<details>
<summary><strong>create</strong> (High-level creation commands)</summary>

- `user`
- `group`
- `role`
- `acl`
- `template`
- `project`
- `project_load`
- `add_pool`
- `create_compute`
- `qemu_img`
- `node`
- `link_create`
- `drawing_create`
- `snapshot_create`
- `add_applience_version`
- `create`
- `project_node_from_template`
- `disk_img`
- `node_file`
- `class`
- `exercise`
</details>

<details>
<summary><strong>add</strong> (Add resources/members)</summary>

- `group_member`
- `ressouce_to_pool`
</details>

<details>
<summary><strong>update</strong> (Update resources)</summary>

- `iou_license`
- `me`
- `user`
- `group`
- `acl`
- `template`
- `project`
- `compute`
- `pool`
- `role`
- `role_privs`
- `node`
- `drawing`
- `link`
- `disk_image`
</details>

<details>
<summary><strong>fuzzy</strong> (Interactive fuzzy-finder commands)</summary>

- `user-info`
- `ui`
- `group-info`
- `gi`
- `group-info-with-usernames`
- `gim`
- `user-info-and-group-membership`
- `uig`
- `chpw`
- `change-password`
</details>

<details>
<summary><strong>stream</strong> (Notification streaming)</summary>

- `notifications`
- `project-id`
</details>

<details>
<summary><strong>install</strong> (Remote installation commands)</summary>

- `ssl` &mdash; Install and configure HTTPS (Caddy reverse proxy) on the remote GNS3 server via SSH.
</details>

---

For more details, see the [PyPI page](https://pypi.org/project/gns3util/) or the [GitHub repository](https://github.com/Stefanistkuhl/gns3-api-util).
