import click
from .api.client import GNS3Error
from click.exceptions import Abort
from click import UsageError
from typing import Any
from gns3util.schemas import Token, AuthFileEntry
from gns3util.utils import validate_response
from dacite import from_dict
from dataclasses import asdict
import json
import sys
import os


def authenticate_user(
    ctx: click.Context, credentials: tuple[str, str]
) -> tuple[GNS3Error, Any]:
    """Authenticate user against GNS3 server and return the response."""
    input_data = {"username": credentials[0], "password": credentials[1]}
    from .api.post_endpoints import GNS3PostAPI

    server_url = ctx.parent.obj["server"]
    verify = ctx.parent.obj["verify"]
    client = GNS3PostAPI(server_url, key=None, verify=verify)
    auth_error, result = client.user_authenticate(input_data)
    return auth_error, result


def save_auth_data(auth_data: Token, server_url: str, username: str, key_file) -> bool:
    """Save authentication data to a file."""
    if not os.path.exists(os.path.abspath(key_file)):
        open(key_file, "a").close()

    try:
        key_entries = []

        with open(key_file, "a") as f:
            key_entry = AuthFileEntry(
                server_url=server_url,
                user=username,
                access_token=auth_data.access_token,
                token_type=auth_data.token_type,
            )
            f.write(json.dumps(asdict(key_entry)) + "\n")

        with open(key_file, "r") as f:
            for line in f:
                key_entries.append(line)

        seen = set()
        unique_list = []
        for key in key_entries:
            try:
                entry_raw = json.loads(key)
                entry = from_dict(data_class=AuthFileEntry, data=entry_raw)
                pair = (entry.server_url, entry.user)
                if pair not in seen:
                    seen.add(pair)
                    unique_list.append(entry)
            except json.JSONDecodeError as e:
                click.secho("Error decoding JSON: ", nl=False, fg="red", err=True)
                click.secho(f"{e}", nl=False, bold=True, err=True)
                click.secho("in line: ", nl=False, err=True)
                click.secho(f"{line}", bold=True, err=True)

        with open(key_file, "w") as f:
            for key in unique_list:
                f.write(json.dumps(asdict(key)) + "\n")
            return True

    except IOError as e:
        click.secho("Error writing to file: ", nl=False, fg="red", err=True)
        click.secho(f"{key_file}: {str(e)}", bold=False, err=True)
        return False
    except Exception as e:
        click.secho(
            "Unexpected error saving authentication data: ",
            nl=False,
            fg="red",
            err=True,
        )
        click.secho(f"{str(e)}", bold=True, err=True)
        return False


def load_key(key_file) -> tuple[bool, list[AuthFileEntry]]:
    try:
        data_arr = []
        with open(key_file) as f:
            data = f.read()
        if not data:
            return False, data_arr
        with open(key_file) as f:
            for line in f:
                data_arr.append(
                    from_dict(data_class=AuthFileEntry, data=json.loads(line))
                )
        return True, data_arr
    except ValueError:
        return False, data_arr
    except FileNotFoundError:
        return False, data_arr


def try_key(ctx: click.Context, key) -> tuple[GNS3Error, Any]:
    from .api.get_endpoints import GNS3GetAPI

    server_url = ctx.parent.obj["server"]
    verify = ctx.parent.obj["verify"]
    client = GNS3GetAPI(server_url, key, verify=verify)
    try_key_error, result = client.current_user_info()
    return try_key_error, result


@click.group()
def auth():
    """Authentication commands."""
    pass


def load_and_try_key(ctx: click.Context) -> tuple[bool, AuthFileEntry | None]:
    key_file = ctx.parent.obj["key_file"] or os.path.expanduser("~/.gns3key")
    load_success, keyData = load_key(key_file)
    if not load_success:
        no_confirm = ctx.obj.get("no_keyfile_confirm", False)
        if no_confirm or click.confirm(
            "Your keyfile contains invalid characters do you want the file to get overwritten?"
        ):
            with open(key_file, "w") as f:
                f.write("")
            load_success = True
        else:
            click.secho("Authentication cancelled.")
            sys.exit(1)
    if load_success:
        for key in keyData:
            if key.server_url == ctx.parent.obj["server"]:
                token = key.access_token
                try_key_error, result = try_key(ctx, token)
                if GNS3Error.has_error(try_key_error):
                    if try_key_error.connection:
                        GNS3Error.print_error(try_key_error)
                        if "https://" in ctx.parent.obj["server"]:
                            click.secho(
                                "You are probably using a self-signed SSL-Cert so try again with the ",
                                nl=False,
                            )
                            click.secho("-i", bold=True, nl=False)
                            click.secho(" flag")
                        ctx.exit(1)
                    if not try_key_error.unauthorized:
                        GNS3Error.print_error(try_key_error)
                        return False, None
                else:
                    return True, key

    if str(ctx.command) == "<Command login>":
        return False, {}

    click.secho("Error: ", fg="red", nl=False, err=True)
    click.secho("couldn't load the keyfile at", err=True, nl=False)
    click.secho(f" {key_file} ", bold=True, err=True, nl=False)
    click.secho("it is either empty or doesn't exist please use", err=True, nl=False)
    click.secho(" gns3util -s [server_url] auth login ", bold=True, err=True, nl=False)
    click.secho("to create it and log into this server", err=True)
    return False, {}


@auth.command()
@click.option(
    "-u",
    "--user",
    default=None,
    envvar="GNS3_USER",
    help="Username for authentication (env: GNS3_USER)",
)
@click.option(
    "-p",
    "--password",
    default=None,
    envvar="GNS3_PASSWORD",
    help="Password for authentication (env: GNS3_PASSWORD). "
    "Use '-' to read from stdin.",
)
@click.option(
    "--no-keyfile-confirm",
    is_flag=True,
    help="Skip confirmation when keyfile is corrupt",
)
@click.pass_context
def login(ctx: click.Context, user, password, no_keyfile_confirm):
    """Perform authentication."""
    ctx.obj = ctx.obj or {}
    if no_keyfile_confirm:
        ctx.obj["no_keyfile_confirm"] = True

    try:
        ok, key = load_and_try_key(ctx)
        if ok and key:
            click.secho("Success: ", fg="green", nl=False)
            click.secho("API key works, logged in as ", nl=False)
            click.secho(f"{key.user}", bold=True)
            return

        if not user and password == "-":
            raise click.UsageError(
                "When using a password from stdin use -u to set a user aswell"
            )

        if not user:
            user = click.prompt("Enter the user to login in as", type=str)

        if password == "-":
            password = sys.stdin.read().rstrip("\n")
        elif password is None:
            try:
                password = click.prompt(
                    "Enter your password", hide_input=True, confirmation_prompt=False
                )
            except Abort:
                click.secho("Authentication cancelled.", err=True)
                return

        auth_error, auth_data_raw = authenticate_user(ctx, (user, password))
        if GNS3Error.has_error(auth_error):
            if auth_error.unauthorized:
                click.secho("Authentication failed: ", fg="red", err=True, nl=False)
                click.secho("Invalid username or password.", err=True)
            else:
                GNS3Error.print_error(auth_error)
            return
        auth_data: Token = validate_response("user_authenticate", auth_data_raw)
        key_file = ctx.parent.obj["key_file"] or os.path.expanduser("~/.gns3key")
        server_url = ctx.parent.obj["server"]
        saved = save_auth_data(auth_data, server_url, user, key_file)
        if not saved:
            click.secho("Authentication failed: ", fg="red", err=True, nl=False)
            click.secho(f"Failed to write token to {key_file}", err=True)
            return

        click.secho("Success: ", fg="green", nl=False)
        click.secho("authenticated as ", nl=False)
        click.secho(f"{user} ", bold=True, nl=False)
        click.secho("and token saved in ", nl=False)
        click.secho(f"{key_file}", bold=True, nl=False)

    except (Abort, KeyboardInterrupt):
        click.secho("Authentication cancelled.", err=True)
        return

    except UsageError:
        raise click.UsageError(
            "When using a password from stdin use -u to set a user aswell"
        )

    except Exception as e:
        click.secho("An unexpected error occurred: ", fg="red", err=True, nl=False)
        click.secho(f"{str(e)}", bold=True, err=True)
        ctx.exit(1)


@auth.command()
@click.pass_context
def status(ctx: click.Context):
    """Display authentication status."""
    load_and_try_key_success, result = load_and_try_key(ctx)
    if load_and_try_key_success and result:
        click.secho("Success: ", nl=False, fg="green")
        click.secho("API key works, logged in as ", nl=False)
        click.secho(f"{result.user}", bold=True)
        return
