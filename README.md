# GNS3 API Util
<center>
<img width=256 src="https://i.imgur.com/t1PNyl4.gif" alt="surely a temporary logo" />
</center>

![Python Version](https://img.shields.io/python/required-version-toml?tomlFilePath=https%3A%2F%2Fraw.githubusercontent.com%2FStefanistkuhl%2Fgns3-api-util%2Frefs%2Fheads%2Fmaster%2Fpyproject.toml)
![GitHub Issues or Pull Requests by label](https://img.shields.io/github/issues/stefanistkuhl/gns3-api-util)
![language count](https://img.shields.io/github/languages/count/stefanistkuhl/gns3-api-util)
![repo size](https://img.shields.io/github/repo-size/stefanistkuhl/gns3-api-util)

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
   - [bottle](https://github.com/bottlepy/bottle)
   - [InquirerPy](https://github.com/kazhala/InquirerPy)

   Additionally [fzf](https://github.com/junegunn/fzf) can be be installed for interactive selections.

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
