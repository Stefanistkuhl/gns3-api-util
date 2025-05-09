import click
from .api.client import GNS3Error
import json
import os
import getpass


def insert_as_first_val(dict_obj, key, value):
    """Insert a key-value pair as the first item in a dictionary."""
    new_dict = {key: value, **dict_obj}
    return new_dict


def get_user_credentials() -> tuple[bool, tuple[str, str]]:
    """Prompt user for username and password."""
    try:
        username = input("Enter the user to login in as:\n")
        password = getpass.getpass("Enter your password:\n")
        return False, (username, password)
    except KeyboardInterrupt:
        click.secho("\nOperation cancelled by user.", err=True)
        return True, ("", "")
    except Exception as e:
        click.secho("Error getting credentials: ",
                    nl=False, err=True, fg="red")
        click.secho(f"{str(e)}", bold=True, err=True)
        return True, ("", "")


def authenticate_user(ctx, credentials: tuple[str, str]) -> tuple[GNS3Error, any]:
    """Authenticate user against GNS3 server and return the response."""
    input_data = {
        "username": credentials[0], "password": credentials[1]}
    from .api.post_endpoints import GNS3PostAPI
    server_url = ctx.parent.obj['server']
    client = GNS3PostAPI(server_url, key=None)
    auth_error, result = client.user_authenticate(input_data)
    return auth_error, result


def save_auth_data(auth_data, server_url, username, key_file) -> bool:
    """Save authentication data to a file."""
    if not os.path.exists(os.path.abspath(key_file)):
        open(key_file, 'a').close()

    try:
        key_entries = []

        with open(key_file, "a") as f:
            resp_dic = insert_as_first_val(auth_data, "user", username)
            resp_dic = insert_as_first_val(
                resp_dic, "server_url", server_url)
            f.write(json.dumps(resp_dic) + "\n")

        with open(key_file, "r") as f:
            for line in f:
                key_entries.append(line)

        seen = set()
        unique_list = []
        for key in key_entries:
            try:
                entry = json.loads(key)
                pair = (entry['server_url'], entry['user'])
                if pair not in seen:
                    seen.add(pair)
                    unique_list.append(entry)
            except json.JSONDecodeError as e:
                click.secho("Error decoding JSON: ",
                            nl=False, fg="red", err=True)
                click.secho(f"{e}", nl=False, bold=True, err=True)
                click.secho("in line: ", nl=False, err=True)
                click.secho(f"{line}", bold=True, err=True)

        with open(key_file, "w") as f:
            for key in unique_list:
                f.write(json.dumps(key) + "\n")
            return True

    except IOError as e:
        click.secho("Error writing to file: ", nl=False, fg="red", err=True)
        click.secho(f"{key_file}: {str(e)}", bold=False, err=True)
        return False
    except Exception as e:
        click.secho("Unexpected error saving authentication data: ",
                    nl=False, fg="red", err=True)
        click.secho(f"{str(e)}", bold=True, err=True)
        return False


def load_key(key_file) -> tuple[bool, list]:
    try:
        data_arr = []
        with open(key_file) as f:
            data = f.read()
        if not data:
            return False, data_arr
        with open(key_file) as f:
            for line in f:
                data_arr.append(json.loads(line))
        return True, data_arr
    except FileNotFoundError:
        return False, data_arr


def try_key(ctx, key) -> tuple[GNS3Error, any]:
    from .api.get_endpoints import GNS3GetAPI
    server_url = ctx.parent.obj['server']
    client = GNS3GetAPI(server_url, key)
    try_key_error, result = client.current_user_info()
    return try_key_error, result


@click.group()
def auth():
    """Authentication commands."""
    pass


auth = click.Group('auth')


def load_and_try_key(ctx) -> tuple[bool, str, str]:
    key_file = os.path.expanduser("~/.gns3key")
    load_success, keyData = load_key(key_file)
    if load_success:
        for key in keyData:
            token = key['access_token']
            user = key['user']
            try_key_error, result = try_key(ctx, token)
            if GNS3Error.has_error(try_key_error):
                if try_key_error.unauthorized:
                    return False, "", ""
                else:
                    return False, "", ""
            else:
                return True, user, token
    return False, "", ""


@auth.command()
@click.pass_context
def login(ctx):
    """Perform authentication."""
    try:
        load_and_try_key_success, user, result = load_and_try_key(ctx)
        if load_and_try_key_success:
            click.secho("Success: ", nl=False, fg="green")
            click.secho("API key works, logged in as ", nl=False)
            click.secho(f"{user}", bold=True)
            return
        key_file = os.path.expanduser("~/.gns3key")
        server_url = ctx.parent.obj['server']
        get_credentials_error, credentials = get_user_credentials()
        if get_credentials_error:
            return

        auth_error, auth_data = authenticate_user(
            ctx, credentials)
        if GNS3Error.has_error(auth_error):
            if auth_error.unauthorized:
                click.secho(
                    "Authentication failed: ", fg="red", err=True, nl=False)
                click.secho(
                    "Invalid username or password.", err=True)
                return
            else:
                GNS3Error.print_error(auth_error)
                return

        save_data_error = save_auth_data(
            auth_data, server_url, credentials[0], key_file)
        if save_data_error:
            click.secho("Success: ", nl=False, fg="green")
            click.secho("authenticated as ", nl=False)
            click.secho(f"{credentials[0]}", nl=False, bold=True)
            click.secho("and token saved in ", nl=False)
            click.secho(f"{key_file}", bold=True, nl=False)
            return
        else:
            click.secho(
                "Authentication failed: ", fg="red", err=True, nl=False)
            click.secho(
                f"Failed to write token to {key_file}", err=True)
            return

    except KeyboardInterrupt:
        click.secho("\nAuthentication interrupted by user.", err=True)
        return

    except Exception as e:
        click.secho("An unexpected error occured: ",
                    nl=False, fg="red", err=True)
        click.secho(f"{str(e)}", bold=True, err=True)
        return


@auth.command()
@click.pass_context
def status(ctx):
    """Display authentication status."""
    load_and_try_key_success, user, result = load_and_try_key(ctx)
    if load_and_try_key_success:
        click.secho("Success: ", nl=False, fg="green")
        click.secho("API key works, logged in as ", nl=False)
        click.secho(f"{user}", bold=True)
        return
