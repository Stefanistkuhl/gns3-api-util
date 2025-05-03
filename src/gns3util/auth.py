import click
import requests
import json
import os
import getpass


def insert_as_first_val(dict_obj, key, value):
    """Insert a key-value pair as the first item in a dictionary."""
    new_dict = {key: value, **dict_obj}
    return new_dict


def get_user_credentials():
    """Prompt user for username and password."""
    try:
        username = input("Enter the user to login in as:\n")
        password = getpass.getpass("Enter your password:\n")
        return username, password
    except KeyboardInterrupt:
        click.secho("\nOperation cancelled by user.", err=True)
        return
    except Exception as e:
        click.secho(f"Error getting credentials: {str(e)}", err=True)
        return


def authenticate_user(username, password, server_url):
    """Authenticate user against GNS3 server and return the response."""
    try:
        url = f'{server_url}/v3/access/users/authenticate'
        headers = {'Content-Type': 'application/json'}
        data = {'username': username, 'password': password}

        response = requests.post(url, json=data, headers=headers, timeout=10)

        if response.status_code == 200:
            return response.json()
        elif response.status_code == 401:
            click.secho(
                "Authentication failed: Invalid username or password.", err=True)
            return None
        else:
            click.secho(f"Server returned error: {
                        response.status_code}", err=True)
            click.secho(f"Response: {response.text}", err=True)
            return None

    except requests.exceptions.ConnectionError:
        click.secho(f"Connection error: Could not connect to {
                    server_url}", err=True)
        return None
    except requests.exceptions.Timeout:
        click.secho(
            "Connection timeout: The server took too long to respond.", err=True)
        return None
    except requests.exceptions.RequestException as e:
        click.secho(f"Request error: {str(e)}", err=True)
        return None
    except Exception as e:
        click.secho(f"Unexpected error during authentication: {
                    str(e)}", err=True)
        return None


def save_auth_data(auth_data, server_url, username, key_file):
    """Save authentication data to a file."""
    try:
        os.makedirs(os.path.dirname(os.path.abspath(key_file)), exist_ok=True)

        with open(key_file, "w") as f:
            resp_dic = insert_as_first_val(auth_data, "user", username)
            resp_dic = insert_as_first_val(resp_dic, "server_url", server_url)
            json.dump(resp_dic, f, indent=4)
            return resp_dic
    except IOError as e:
        click.secho(f"Error writing to file {key_file}: {str(e)}", err=True)
        return None
    except Exception as e:
        click.secho(f"Unexpected error saving authentication data: {
                    str(e)}", err=True)
        return None


def load_key(key_file):
    try:
        with open(key_file) as f:
            data = f.read()
        data = json.loads(data)
        return data
    except FileNotFoundError:
        return False


def try_key(key, server_url):
    url = f'{server_url}/v3/access/users/me'
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }

    response = requests.get(url, headers=headers, timeout=10)
    try:
        if response.status_code == 200:
            return True, response.json()
        elif response.status_code == 401:
            click.secho(
                "User unautorized please log in to refresh your key", err=True)
            return False, None
        else:
            click.secho(f"Server returned error: {
                        response.status_code}", err=True)
            click.secho(f"Response: {response.text}", err=True)
            return False, None
    except requests.exceptions.ConnectionError:
        click.secho(f"Connection error: Could not connect to {
                    server_url}", err=True)
        return False, None
    except requests.exceptions.Timeout:
        click.secho(
            "Connection timeout: The server took too long to respond.", err=True)
        return False, None
    except requests.exceptions.RequestException as e:
        click.secho(f"Request error: {str(e)}", err=True)
        return False, None
    except Exception as e:
        click.secho(f"Unexpected error during user check: {str(e)}", err=True)
        return False, None


@click.group()
def auth():
    """Authentication commands."""
    pass


auth = click.Group('auth')


@auth.command()
@click.pass_context
def login(ctx):
    """Perform authentication."""
    try:
        key_file = os.path.expanduser("~/.gns3key")
        server_url = ctx.parent.obj['server']
        keyData = load_key(key_file)
        if keyData:
            resp, usr = try_key(keyData, server_url)
            if resp:
                click.secho(f"API key works, logged in as {
                            usr.get('username', 'unknown')}")
                return

        username, password = get_user_credentials()

        auth_data = authenticate_user(username, password, server_url)
        if not auth_data:
            return

        saved_data = save_auth_data(auth_data, server_url, username, key_file)
        if saved_data:
            click.secho(
                "Authentication successful. Credentials saved.")
            return
        else:
            return

    except KeyboardInterrupt:
        click.secho("\nAuthentication interrupted by user.", err=True)
        return

    except Exception as e:
        click.secho(f"An unexpected error occurred: {str(e)}", err=True)
        return


@auth.command()
@click.pass_context
def status(ctx):
    """Display authentication status."""
    try:
        key_file = os.path.expanduser("~/.gns3key")
        server_url = ctx.parent.obj['server']
        keyData = load_key(key_file)
        if keyData:
            resp, usr = try_key(keyData, server_url)
            if resp:
                click.secho(f"Logged in as: {usr.get('username', 'unknown')}")
                return
            else:
                click.secho("No active login found.", err=True)
                return
        else:
            click.secho(
                "No saved credentials found. Please authenticate.", err=True)
            return

    except Exception as e:
        click.secho(f"An unexpected error occurred: {str(e)}", err=True)
        return
