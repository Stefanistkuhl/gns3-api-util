import click
import paramiko
import os
import re
import importlib.resources
from dataclasses import dataclass
from typing import Optional


@click.group()
def install():
    """Install something in the GNS3 server"""
    pass


@dataclass
class install_ssl_args:
    firewall_allow: Optional[str]
    firewall_block: Optional[bool | None]
    verbose: Optional[bool]
    reverse_proxy_port: int = 443
    domain: str = ""
    gns3_port: int = 3080
    subject: str = "/CN=localhost"


def edit_script_with_flags(input: str, params: install_ssl_args) -> str:
    modified_lines = []
    for line in input.splitlines():
        if 'UFW=""' in line:
            if params.firewall_block:
                modified_lines.append(line.replace('UFW=""', 'UFW="ufw"'))
        elif 'RP_PORT=""' in line:
            modified_lines.append(line.replace(
                'RP_PORT=""', f'RP_PORT="{params.reverse_proxy_port}"'))
        elif 'GNS3_PORT=""' in line:
            modified_lines.append(line.replace(
                'GNS3_PORT=""', f'GNS3_PORT="{params.gns3_port}"'))
        elif 'DOMAIN=""' in line:
            if params.domain != '""':
                modified_lines.append(line.replace(
                    'DOMAIN=""', f'DOMAIN="{params.domain}"'))
            else:
                modified_lines.append(line.replace(
                    'DOMAIN=""', 'DOMAIN=""'))
        elif 'SUBJ=""' in line:
            modified_lines.append(line.replace(
                'SUBJ=""', f'SUBJ="{params.subject}"'))
        elif 'UFW_ENABLE' in line:
            if params.firewall_block:
                if params.firewall_allow and params.firewall_allow != '""':
                    modified_lines.append("echo 'y' | $SUDO ufw enable")
                    modified_lines.append(line.replace(
                        'UFW_ENABLE', f'$SUDO ufw allow proto tcp from {params.firewall_allow} to any port {params.gns3_port}'))
                else:
                    modified_lines.append("echo 'y' | $SUDO ufw enable")
                    modified_lines.append(line.replace(
                        'UFW_ENABLE', f'$SUDO ufw deny {params.gns3_port}'))
            else:
                modified_lines.append(line.replace(
                    'UFW_ENABLE', ''))

        else:
            modified_lines.append(line)

    script = "\n".join(modified_lines)
    return script


def file_opts_to_params(input: str) -> install_ssl_args:
    reverse_proxy_port = 443
    gns3_port = 3080
    domain = ""
    subj = "/CN=localhost"
    firewall_allow = None
    firewall_block = False
    verbose = False
    for line in input.splitlines():
        if "REVERSE_PROXY_PORT" in line:
            split = line.split("=")
            try:
                reverse_proxy_port = int(split[-1])
            except ValueError:
                raise click.UsageError(
                    "Please enter an integer as the port number for the reverse proxy.")
        elif "GNS3_PORT" in line:
            split = line.split("=")
            try:
                gns3_port = int(split[-1])
            except ValueError:
                raise click.UsageError(
                    "Please enter an integer as the port number for the GNS3 server")
        elif "DOMAIN" in line:
            split = line.split("=")
            if "" in line:
                domain = '""'
            else:
                domain = line[-1]
        elif "SUBJECT" in line:
            index_of_first_equals = line.find("=")
            if index_of_first_equals != -1:
                subj = line[index_of_first_equals+1:]
        elif "FIREWALL_ALLOW" in line:
            split = line.split("=")
            if split[-1] == "":
                firewall_allow = None
            else:
                firewall_allow = split[-1]
        elif "FIREWALL_BLOCK" in line:
            split = line.split("=")
            if split[-1].rstrip().lower() == "false":
                firewall_block = False
            elif split[-1].rstrip().lower() == "true":
                firewall_block = True
        elif "VERBOSE" in line:
            split = line.split("=")
            if split[-1].rstrip().lower() == "false":
                verbose = False
            elif split[-1].rstrip().lower() == "true":
                verbose = True

    return install_ssl_args(
        firewall_allow=firewall_allow,
        firewall_block=firewall_block,
        domain=domain,
        subject=subj,
        reverse_proxy_port=reverse_proxy_port,
        gns3_port=gns3_port,
        verbose=verbose
    )


def is_valid_subj(subj: str) -> bool:
    """Return True if dn matches the required /key=value[/…] format."""
    subj_pattern = re.compile(
        r"^(?=.*?/CN=[^/=]+)"
        r"(?:/"
        r"(?:C=[A-Z]{2}"
        r"|ST=[^/=]+"
        r"|L=[^/=]+"
        r"|O=[^/=]+"
        r"|OU=[^/=]+"
        r"|CN=[^/=]+"
        r"|emailAddress=[^/=]+)"
        r")+$"
    )
    return bool(subj_pattern.match(subj))


def is_valid_domain(input: str):
    pattern = r"^[A-Za-z0-9-]{1,63}\.[A-Za-z]{2,6}$"

    return bool(re.match(pattern, input))


def is_valid_ip(input: str):
    pattern = re.compile(
        r"^"
        r"("
        r"(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\."
        r"){3}"
        r"(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])"
        r"/"
        r"(3[0-2]|[12]?[0-9])"
        r"$"
    )
    return bool(re.match(pattern, input))


def validate_install_ssl_input(
    firewall_allow: Optional[str | None],
    firewall_block: Optional[bool],
    reverse_proxy_port: int = 443,
    domain: str = "",
    gns3_port: int = 3080,
    subject: str = "/CN=localhost",
    verbose=False
) -> install_ssl_args:
    if firewall_allow and not firewall_allow:
        raise click.UsageError(
            "If a firewall allow ip range is set the --firewall-block flag must be set aswell.")
    if firewall_allow and firewall_block:
        if firewall_allow != '""':
            if not is_valid_ip(firewall_allow):
                raise click.UsageError(
                    "Please enter a Network in the following format: \n xxx.xxx.xxx.xxx/xx")

    if reverse_proxy_port > 65535 or gns3_port > 65535:
        raise click.UsageError(
            "Please use a valid port number that is smaller than 65535")
    if subject != "/CN=localhost":
        if not is_valid_subj(subject):
            raise click.UsageError(
                "Invalid subj please follow this formatting:\n /C=COUNTRY-CODE/ST=STATE/L=CITY/O=Organization/OU=Organizational-Unit/CN=common-name/emailAddress=email\n You only need this at minumum which get's used if the option isn't set: \n /CN=localhost \n")
    if domain != '""':
        if not is_valid_domain(domain):
            raise click.UsageError(
                f"{domain} is not a valid domain name if please enter a valid one, or not use this option if you don't want to use one.")
    return install_ssl_args(
        firewall_allow=firewall_allow,
        firewall_block=firewall_block,
        reverse_proxy_port=reverse_proxy_port,
        domain=domain,
        gns3_port=gns3_port,
        subject=subject,
        verbose=verbose
    )


def push_and_run_script_via_heredoc(
    ssh_client: paramiko.SSHClient,
    package: str,
    resource_name: str,
    params: install_ssl_args,
    remote_path: str = "/tmp/setup_https.sh",
) -> bool:
    """
    1) Loads `resource_name` from `package`
    2) Pushes it to `remote_path` via a single `cat << 'EOF' … EOF`
    3) chmod +x, executes it, then removes it.
    Returns True if script ran successfully, False otherwise.
    """
    script_text = importlib.resources.files(package) \
        .joinpath(resource_name) \
        .read_text(encoding="utf-8")

    script_text = edit_script_with_flags(script_text, params)

    heredoc = (
        f"cat << 'EOF' > {remote_path}\n"
        f"{script_text.rstrip()}\n"
        "EOF\n"
        f"chmod +x {remote_path}\n"
    )

    stdin, stdout, stderr = ssh_client.exec_command(heredoc)
    out = stdout.read().decode()
    err = stderr.read().decode()
    if out:
        if params.verbose:
            click.secho("HEREDOC STDOUT: ", fg="white", nl=False)
            click.secho(out, bold=True)
    if err:
        click.secho("HEREDOC STDERR: ", fg="red", nl=False)
        click.secho(err, bold=True)

    stdin, stdout, stderr = ssh_client.exec_command(f"bash {remote_path}")
    out = stdout.read().decode()
    err = stderr.read().decode()
    exit_status = stdout.channel.recv_exit_status()
    if out:
        if params.verbose:
            click.secho("SCRIPT STDOUT: ", fg="white", nl=False)
            click.secho(out, bold=True)
    if err:
        click.secho("SCRIPT STDERR: ", fg="red", nl=False)
        click.secho(err, bold=True)

    ssh_client.exec_command(f"rm -f {remote_path}")
    return exit_status == 0


def parse_server_url_for_ssh(server_url, port_option):
    """
    Parses a server URL (e.g., 'https://10.10.10.10:3080' or '10.10.10.10')
    into a hostname and port suitable for SSH connection.
    Prioritizes the 'port_option' value if provided.
    Defaults to port 22 if no explicit port is given.
    """
    hostname_part = re.sub(r'https?://', '', server_url)
    parts = hostname_part.split(':')
    hostname = parts[0]
    default_ssh_port = 22

    if port_option is not None:
        port = port_option
    elif len(parts) > 1 and parts[1].isdigit():
        port = int(parts[1])
    else:
        port = default_ssh_port

    return hostname, port


def ssh_connect_with_key_or_password(hostname, username, port,
                                     custom_private_key_path=None, verbose: bool = False):
    """
    Connects to an SSH server trying common private key locations first,
    then a custom path if provided, and finally falls back to password.
    """
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

    default_key_names = ["id_rsa", "id_dsa", "id_ecdsa", "id_ed25519"]
    potential_key_paths = []

    if custom_private_key_path:
        potential_key_paths.append(os.path.expanduser(custom_private_key_path))

    ssh_dir = os.path.expanduser("~/.ssh")
    if os.path.isdir(ssh_dir):
        for key_name in default_key_names:
            key_path = os.path.join(ssh_dir, key_name)
            if os.path.exists(key_path) and key_path not in potential_key_paths:
                potential_key_paths.append(key_path)

    key_classes = [
        paramiko.RSAKey,
        paramiko.DSSKey,
        paramiko.ECDSAKey,
        paramiko.Ed25519Key,
    ]

    for private_key_path in potential_key_paths:
        if not os.path.exists(private_key_path):
            if verbose:
                click.secho(f"Info: Private key file not found: {
                            private_key_path}")
            continue

        for key_cls in key_classes:
            try:
                pkey = key_cls.from_private_key_file(private_key_path)
            except paramiko.PasswordRequiredException:
                if verbose:
                    click.secho(
                        f"Info: Key {private_key_path} is encrypted; skipping."
                    )
                break
            except paramiko.SSHException:
                continue
            except Exception as e:
                if verbose:
                    click.secho(
                        f"Warning: Error loading {key_cls.__name__} from "
                        f"{private_key_path}: {e}"
                    )
                continue

            if verbose:
                click.secho(
                    f"Attempting connection to {hostname}:{port} with "
                    f"{key_cls.__name__} key: {private_key_path}..."
                )
            try:
                client.connect(
                    hostname,
                    port=port,
                    username=username,
                    pkey=pkey,
                )
                click.secho("Successfully connected with private key.")
                return client
            except paramiko.SSHException as e:
                if verbose:
                    click.secho(
                        f"Warning: Failed to connect with {private_key_path} "
                        f"({key_cls.__name__}): {e}. Trying next key."
                    )
            except Exception as e:
                if verbose:
                    click.secho(
                        f"Warning: Error using {private_key_path} "
                        f"({key_cls.__name__}): {e}. Trying next key."
                    )

    click.secho("All key-based authentication attempts failed or no keys found.")
    click.secho("Falling back to password authentication.")

    try:
        password = click.prompt("Enter SSH password", hide_input=True)
        if verbose:
            click.secho(
                f"Attempting connection to {hostname}:{port} with password..."
            )
        client.connect(
            hostname,
            port=port,
            username=username,
            password=password,
        )
        click.secho("Successfully connected with password.")
        return client
    except paramiko.AuthenticationException:
        click.secho("Authentication failed. Please check your credentials.")
    except paramiko.SSHException as e:
        click.secho(f"Could not establish SSH connection: {e}")
    except Exception as e:
        click.secho(f"An unexpected error occurred: {e}")

    return None


interactive_options_text = """REVERSE_PROXY_PORT=443
GNS3_PORT=3080
DOMAIN=""
SUBJECT=/CN=localhost
FIREWALL_ALLOW=""
FIREWALL_BLOCK=False
VERBOSE=False
"""


@install.command(name="ssl", help="Install caddy on the remote host over ssh as a reverse proxy for GNS3 to have HTTPS only works linux servers")
@click.argument("user", required=True, type=str)
@click.option(
    "-p", "--port", required=False, type=int, default=22, help="SSH port."
)
@click.option(
    "-k",
    "--key",
    "private_key_path",
    required=False,
    type=click.Path(exists=True, dir_okay=False, resolve_path=True),
    help="Path to a custom SSH private key file.",
)
@click.option(
    "-rp",
    "--reverse-proxy-port",
    "reverse_proxy_port",
    type=int,
    default=443,
    help="Port for the reverse proxy to use.",
)
@click.option(
    "-gp",
    "--gns3-port",
    "gns3_port",
    type=int,
    default=3080,
    help="Port of the GNS3 Server.",
)
@click.option(
    "-d", "--domain", default="",  help="Domain to use for the reverse proxy."
)
@click.option(
    "-s",
    "--subject",
    default="/CN=localhost",
    help="Set the subject alternative name for the SSL certificate.",
)
@click.option(
    "-fa",
    "--firewall-allow",
    "firewall_allow",
    type=str,
    default=None,
    required=False,
    help=(
        "Block all connections to the GNS3 server port and only allow a given subnet. Example: 10.0.0.0/24"
    ),
)
@click.option(
    "-fb",
    "--firewall-block",
    "firewall_block",
    is_flag=True,
    default=False,
    required=False,
    help="Block all connections to the port of the GNS3 server.",
)
@click.option(
    "-i",
    "--interactive",
    is_flag=True,
    help="Set the options for this command interactively.",
)
@click.option(
    "-v", "--verbose", "verbose", is_flag=True, help="Run this command with extra logging."
)
@click.pass_context
def install_ssl(
    ctx: click.Context,
    user,
    port,
    private_key_path,
    reverse_proxy_port,
    gns3_port,
    domain,
    subject,
    firewall_allow,
    firewall_block,
    interactive,
    verbose,
):
    """
    Connects to the server via SSH and sets up SSL using
    Caddy as a reverse proxy—but only if the remote user
    is root or has passwordless sudo.
    """
    if interactive:
        opts = click.edit(text=interactive_options_text)
        interavtive_params = file_opts_to_params(opts)
        params = validate_install_ssl_input(
            interavtive_params.firewall_allow, interavtive_params.firewall_block, interavtive_params.reverse_proxy_port, interavtive_params.domain, interavtive_params.gns3_port, interavtive_params.subject, verbose=interavtive_params.verbose)
    else:
        params = validate_install_ssl_input(
            firewall_allow, firewall_block, reverse_proxy_port, domain, gns3_port, subject, verbose=verbose)
    server_url = ctx.parent.obj.get("server")
    hostname, ssh_port = parse_server_url_for_ssh(server_url, port)
    if params.verbose:
        click.secho("Attempting to connect to SSH at: ", fg="white", nl=False)
        click.secho(f"{hostname}:{ssh_port}", bold=True)

    ssh_client = ssh_connect_with_key_or_password(
        hostname, user, ssh_port, private_key_path
    )
    if not ssh_client:
        click.secho("Error: ", fg="red", nl=False, err=True)
        click.secho(
            "Failed to establish SSH connection for SSL installation.", err=True, bold=True
        )
        return

    try:
        stdin, stdout, stderr = ssh_client.exec_command("id -u")
        uid = int(stdout.read().decode().strip() or "-1")
        if uid != 0:
            stdin, stdout, stderr = ssh_client.exec_command(
                "sudo -n true"
            )
            exit_status = stdout.channel.recv_exit_status()
            if exit_status != 0:
                click.secho(
                    "Error: ",
                    fg="red", nl=False
                )
                click.secho(
                    "remote user is not root "
                    "and lacks passwordless sudo privileges.", err=True, bold=True
                )
                ssh_client.close()
                return
        if params.verbose:
            click.secho(
                "Privilege check passed (root or passwordless sudo).",
                fg="green", bold=True
            )
    except Exception as e:
        click.secho(
            "Error checking remote privileges: ", fg="red", nl=False, err=True)
        click.secho(f"{e}", bold=True, err=True)
        ssh_client.close()
        return

    if params.verbose:
        click.secho("\nSSH connection established.", fg="green", bold=True)
    try:
        success = push_and_run_script_via_heredoc(
            ssh_client,
            package="gns3util.resources",
            resource_name="setup_https.sh",
            params=params,
            remote_path="/tmp/setup_https.sh",
        )
    finally:
        ssh_client.close()
        if params.verbose:
            click.secho("SSH connection closed.", fg="white", bold=True)
        if 'success' in locals() and success:
            click.secho("Success: ", fg="green", nl=False)
            click.secho(
                f"Setup caddy as reverse proxy on the remote server and listening on port {
                    params.reverse_proxy_port}.",
                bold=True
            )
