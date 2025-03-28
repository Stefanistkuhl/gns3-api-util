# GNS3 API Util

A command-line utility for interacting with the GNS3 API. This tool streamlines common API operations—such as authentication, GET, POST, PUT, and DELETE requests—against a GNS3 server, making it easier to integrate and automate tasks in your network emulation environments.

## Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/Stefanistkuhl/gns3-api-util.git
   cd gns3-api-util
   ```

2. **Install Dependencies**

   Use a virtual environment. Then install the project along with its dependencies:

   ```bash
   python -m venv venv
   source venv/bin/activate
   pip install -e .
   ```

   This will install the required packages:
   
   - [click](https://click.palletsprojects.com/)
   - [requests](https://docs.python-requests.org/)
   - [rich](https://github.com/Textualize/rich)

## Usage

After installing, the utility can be executed directly from the command line using the entry point `gns3util`.

### Running the CLI

At a minimum, provide the `--server` (or `-s`) option with the URL of your GNS3 server:

```bash
gns3util --server http://<GNS3_SERVER_ADDRESS>
```

### Commands

The CLI supports several subcommands to interact with the GNS3 API:

- **auth**: Manage authentication.
- **get**: Perform GET requests.
- **post**: Perform POST requests.
- **put**: Perform PUT requests.
- **delete**: Perform DELETE requests.

For example, to run an authentication command:

```bash
gns3util auth --server http://localhost:3080 [additional-options]
```

Replace `[additional-options]` with any parameters required by the subcommand.

### Help

You can view the help text by using the `--help` option:

```bash
gns3util --help
```

This will display usage information and options for each command.

Todo

- [ ] Add shell completion

- [ ] Provide a help option for all commands

- [ ] Store multiple keys at once and select a server more easily
