import json
import getpass
import rich
import click
from dataclasses import dataclass
import subprocess
from typing import Callable, Any, Optional
from .api.client import GNS3Error

GREY = "\033[90m"
CYAN = "\033[96m"
RESET = "\033[0m"


@dataclass
class fuzzy_info_params:
    ctx: Any
    client: Callable[[Any], Any]
    opt_method: Optional[str] = None
    opt_key: Optional[str] = None
    method: str = "str"
    key: str = "str"
    multi: bool = False
    opt_data: bool = False


@dataclass
class fuzzy_password_params:
    ctx: Any
    client: Callable[[Any], Any]
    method: str = "str"
    key: str = "str"
    multi: bool = False


def fzf_select(options, multi=False):
    """
    Opens an fzf window with the given options and returns the selected option(s).

    Args:
        options: A list of strings representing the options to choose from.

    Returns:
        A list of strings containing the selected option(s), or an empty list if none were selected or if fzf is not found.
    """
    try:
        if multi:
            fzf_process = subprocess.Popen(
                ['fzf', '--multi'],
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )
        else:
            fzf_process = subprocess.Popen(
                ['fzf'],
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )

        output, error = fzf_process.communicate('\n'.join(options))

        if error:
            if "fzf: command not found" in error:
                print(
                    "Error: fzf is not installed. Please install it to use this feature.")
                return []
            else:
                print(f"Error running fzf: {error}")
                return []

        if output:
            return [line.strip() for line in output.strip().split('\n')]
        else:
            return []

    except FileNotFoundError:
        print("Error: fzf executable not found in PATH. Please ensure it's installed and accessible.")
        return []


def fuzzy_info(params=fuzzy_info_params) -> GNS3Error:
    error = GNS3Error()
    fzf_input_data, api_data, get_fzf_input_error = get_values_for_fuzzy_input(
        params)
    if GNS3Error.has_error(get_fzf_input_error):
        return get_fzf_input_error
    selected = fzf_select(fzf_input_data, multi=params.multi)
    matched = set()
    for selected_item in selected:
        for a in api_data:
            if a[params.key] == selected_item and a[params.key] not in matched:
                for k, v in a.items():
                    print(f"{CYAN}{k}{RESET}: {v}")
                print(f"{GREY}---{RESET}")
                if params.opt_data:
                    opt_data_error, opt_data = getattr(params.client(
                        params.ctx), params.opt_method)(a[params.opt_key])
                    if GNS3Error.has_error(opt_data_error):
                        error.request_network_error = True
                        return error
                    if opt_data == []:
                        print(f"Empty data returned from method {
                              params.opt_method} for the {a[params.key]} value")
                    else:
                        for d in opt_data:
                            print(f"{GREY}---{RESET}")
                            for k2, v2 in d.items():
                                print(f"{CYAN}{k2}{RESET}: {v2}")
                            print(f"{GREY}---{RESET}")
                break
    return error


def get_values_for_fuzzy_input(params) -> (list, list, GNS3Error):
    from . import get
    fuzzy_error = GNS3Error()
    get_data_error, api_data = getattr(
        get.get_client(params.ctx), params.method)()
    if GNS3Error.has_error(get_data_error):
        fuzzy_error.connection = True
        return None, None, fuzzy_error
    fzf_input_data = []
    for data in api_data:
        fzf_input_data.append(data[params.key])
    return fzf_input_data, api_data, fuzzy_error


def fuzzy_change_password(params=fuzzy_password_params) -> GNS3Error:
    from . import put
    error = GNS3Error()
    fzf_input_data, api_data, get_fzf_input_error = get_values_for_fuzzy_input(
        params)
    if GNS3Error.has_error(get_fzf_input_error):
        return get_fzf_input_error
    selected = fzf_select(fzf_input_data, multi=params.multi)
    matched = set()
    for selected_item in selected:
        for a in api_data:
            if a[params.key] == selected_item and a[params.key] not in matched:
                print(f"Changing the password for user {a['username']}")
                pw = getpass.getpass("Enter the desired password:\n")
                input_data = {"password": pw}
                client = put.get_client(params.ctx)
                change_password_error, result = client.update_user(
                    a['user_id'], input_data)
                if GNS3Error.has_error(change_password_error):
                    return error
                print(f"Successfully changed the password for user {
                      a['username']}")
                break
    return error


def parse_json(filepath: str) -> tuple[bool, Any]:
    try:
        with open(filepath, 'r') as f:
            data = json.load(f)
        return False, data
    except FileNotFoundError:
        return True, f"File not found: {filepath}"
    except json.JSONDecodeError as e:
        return True, f"Invalid JSON in {filepath}: {e}"
    except Exception as e:
        return True, f"An unexpected error occurred: {e}"


def create_class(ctx, filename: str):
    error_load, data = parse_json(filename)

    if error_load:
        click.echo(
            f"Failed to load the file {filename}. Error: {data}", err=True
        )
        return

    class_name = list(data.keys())[0]
    class_obj = data[class_name]
    create_group_error = create_user_group(ctx, class_name)
    if GNS3Error.has_error(create_group_error):
        print("handle this later")
        return
    for group_name, group_obj in class_obj.items():
        create_group_error = create_user_group(ctx, group_name)
        if GNS3Error.has_error(create_group_error):
            print("handle this later")
            return
        students = group_obj["students"]
        for student in students:
            create_user_error = create_user(ctx, student)
            if GNS3Error.has_error(create_user_error):
                print("handle this later")
                return


def create_user_group(ctx, group_name) -> GNS3Error:
    from . import post
    error = GNS3Error()
    input_data = {"name": group_name}
    client = post.get_client(ctx)
    create_group_error, result = client.create_group(input_data)
    if GNS3Error.has_error(create_group_error):
        return error
    click.echo(f"Successfully created the group {group_name}")
    return error


def create_user(ctx, user_dict: dict) -> GNS3Error:
    from . import post
    error = GNS3Error()
    if user_dict["fullName"] != "":
        input_data = {
            "username": user_dict["userName"], "full_name": user_dict["fullName"], "email": user_dict["email"], "password": user_dict["password"]}
    else:
        input_data = {
            "username": user_dict["userName"], "email": user_dict["email"], "password": user_dict["password"]}
    client = post.get_client(ctx)
    create_user_error, result = client.create_user(input_data)
    if GNS3Error.has_error(create_user_error):
        return error
    click.echo(f"Successfully created the uesr {input_data["username"]}")
    return error


def fuzzy_info_wrapper(params):
    error = fuzzy_info(params)
    if error.connection:
        click.echo(
            "Failed to fetch data from the API check your Network connection to the server", err=True)


def fuzzy_put_wrapper(params):
    error = fuzzy_change_password(params)
    if error.connection:
        click.echo(
            "Failed to fetch data from the API check your Network connection to the server", err=True)


def execute_and_print(ctx, client, func):
    error, data = func(client)
    if not GNS3Error.has_error(error):
        rich.print_json(json.dumps(data, indent=2))
