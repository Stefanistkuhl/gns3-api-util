import json
from enum import Enum
import importlib
import uuid
import getpass
import rich
import click
from dataclasses import dataclass
import subprocess
from typing import Callable, Any, Optional
from .api.client import GNS3Error


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


@dataclass
class create_acl_params:
    ctx: Any
    ace_type: str = "str"
    allowed: bool = False
    isGroup: bool = False
    id: str = "str"
    path: str = "str"
    propagate: bool = False
    role_id: str = "str"


class fuzzy_params_type(Enum):
    user_info = 1
    group_info = 2
    group_info_with_usernames = 3
    user_info_and_group_membership = 4


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
                click.secho(
                    "Error: ", nl=False, fg="red", err=True)
                click.secho(
                    "fzf is not installed. Please install it to use this feature.", bold=True, err=True)
                return []
            else:
                click.secho(
                    "Error running fzf: ", nl=False, fg="red", err=True)
                click.secho(
                    f"{error}", bold=True, err=True)
                return []

        if output:
            return [line.strip() for line in output.strip().split('\n')]
        else:
            return []

    except FileNotFoundError:
        click.secho(
            "Error: ", nl=False, fg="red", err=True)
        click.secho(
            "fzf executable not found in PATH. Please ensure it's installed and accessible.", bold=True, err=True)
        return []


def call_client_method(ctx, module_name: str, method_name: str, *args: Any) -> tuple[GNS3Error, Any]:
    module = importlib.import_module(f".{module_name}", package=__package__)
    client = module.get_client(ctx)
    method = getattr(client, method_name)
    return method(*args)


def print_key_value_with_secho(key, value, color="cyan", reset="reset"):
    click.secho(f"{key}: ", fg=color, nl=False)
    click.secho(value)


def print_separator_with_secho(color="white"):
    click.secho("---", fg=color)


def fuzzy_info(params=fuzzy_info_params) -> GNS3Error:
    error = GNS3Error()
    fzf_input_data, api_data, get_fzf_input_error = get_values_for_fuzzy_input(
        params)
    if GNS3Error.has_error(get_fzf_input_error):
        GNS3Error.print_error(get_fzf_input_error)
        return get_fzf_input_error
    selected = fzf_select(fzf_input_data, multi=params.multi)
    matched = set()
    for selected_item in selected:
        for a in api_data:
            if a[params.key] == selected_item and a[params.key] not in matched:
                print_separator_with_secho()
                for k, v in a.items():
                    print_key_value_with_secho(k, v)
                print_separator_with_secho()
                if params.opt_data:
                    opt_data_error, opt_data = getattr(params.client(
                        params.ctx), params.opt_method)(a[params.opt_key])
                    if GNS3Error.has_error(opt_data_error):
                        GNS3Error.print_error(opt_data_error)
                        error.request_network_error = True
                        return error
                    if opt_data == []:
                        click.secho(f"Empty data returned from method {
                            params.opt_method} for the {a[params.key]} value", bold=True, err=True)
                    else:
                        for d in opt_data:
                            print_separator_with_secho()
                            for k2, v2 in d.items():
                                print_key_value_with_secho(k2, v2)
                            print_separator_with_secho()
                break
    return error


def get_values_for_fuzzy_input(params) -> (list, list, GNS3Error):
    get_data_error, api_data = call_client_method(
        params.ctx, "get", params.method)
    fuzzy_error = GNS3Error()
    if GNS3Error.has_error(get_data_error):
        fuzzy_error.connection = True
        return None, None, fuzzy_error
    fzf_input_data = []
    for data in api_data:
        fzf_input_data.append(data[params.key])
    return fzf_input_data, api_data, fuzzy_error


def fuzzy_change_password(params=fuzzy_password_params) -> GNS3Error:
    fzf_input_data, api_data, get_fzf_input_error = get_values_for_fuzzy_input(
        params)
    if GNS3Error.has_error(get_fzf_input_error):
        return get_fzf_input_error
    selected = fzf_select(fzf_input_data, multi=params.multi)
    matched = set()
    for selected_item in selected:
        for a in api_data:
            if a[params.key] == selected_item and a[params.key] not in matched:
                click.secho("Changing the password for user ", nl=False)
                click.secho(f"{a['username']}", bold=True)
                pw = getpass.getpass("Enter the desired password:\n")
                input_data = {"password": pw}
                change_password_error, result = call_client_method(
                    params.ctx, "put", "update_user", a['user_id'], input_data)
                if GNS3Error.has_error(change_password_error):
                    GNS3Error.print_error(change_password_error)
                    return change_password_error
                click.secho("Success: ", nl=False, fg="green")
                click.secho("changed the password for user ", nl=False)
                click.secho(f"{a['username']}", bold=True)
                break
    return change_password_error


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


def add_user_to_group(ctx, user_id: str, group_id: str) -> GNS3Error:
    add_user_to_group_error, result = call_client_method(
        ctx, "put", "add_group_member", group_id, user_id)
    if GNS3Error.has_error(add_user_to_group_error):
        return add_user_to_group_error
    return add_user_to_group_error


def create_class(ctx, filename: str) -> tuple[str, bool]:
    error_load, data = parse_json(filename)

    if error_load:
        click.secho("Error: ", nl=False, fg="red", err=True)
        click.secho("Failed to load file: ", nl=False, err=True)
        click.secho(f"{data}", bold=True, err=True)
        return "", False

    class_name = list(data.keys())[0]
    class_obj = data[class_name]
    class_id, create_group_error = create_user_group(ctx, class_name)
    if GNS3Error.has_error(create_group_error):
        GNS3Error.print_error(create_group_error)
        return class_name, False
    click.secho("Success: ", nl=False, fg="green")
    click.secho(f"created the group{class_name}")
    for group_name, group_obj in class_obj.items():
        group_id, create_group_error = create_user_group(ctx, group_name)
        if GNS3Error.has_error(create_group_error):
            GNS3Error.print_error(create_group_error)
            return class_name, False
        click.secho("Success: ", nl=False, fg="green")
        click.secho(f"created the group{group_name}")
        students = group_obj["students"]
        for student in students:
            user_id, create_user_error = create_user(ctx, student)
            if GNS3Error.has_error(create_user_error):
                GNS3Error.print_error(create_user_error)
                return class_name, False
            click.secho("Success: ", nl=False, fg="green")
            click.secho(f"created the user {student['userName']}")
            add_user_to_class_error = add_user_to_group(ctx, user_id, class_id)
            if GNS3Error.has_error(add_user_to_class_error):
                GNS3Error.print_error(add_user_to_class_error)
                return class_name, False
            add_user_to_group_error = add_user_to_group(ctx, user_id, group_id)
            if GNS3Error.has_error(add_user_to_group_error):
                GNS3Error.print_error(add_user_to_group_error)
                return class_name, False
    return class_name, True


def create_user_group(ctx, group_name) -> (str, GNS3Error):
    input_data = {"name": group_name}
    create_group_error, result = call_client_method(
        ctx, "post", "create_group", input_data)
    if GNS3Error.has_error(create_group_error):
        return "", create_group_error
    return result['user_group_id'], create_group_error


def create_user(ctx, user_dict: dict) -> (str, GNS3Error):
    if user_dict["fullName"] != "":
        input_data = {
            "username": user_dict["userName"], "full_name": user_dict["fullName"], "email": user_dict["email"], "password": user_dict["password"]}
    else:
        input_data = {
            "username": user_dict["userName"], "email": user_dict["email"], "password": user_dict["password"]}
    create_user_error, result = call_client_method(
        ctx, "post", "create_user", input_data)
    if GNS3Error.has_error(create_user_error):
        return "", create_user_error
    return result['user_id'], create_user_error


def get_fuzzy_info_params(input: fuzzy_params_type, ctx, get_client, multi: bool) -> fuzzy_info_params:
    if input == fuzzy_params_type.user_info:
        return fuzzy_info_params(
            ctx=ctx,
            client=get_client,
            method="users",
            key="username",
            multi=multi,
            opt_data=False
        )
    elif input == fuzzy_params_type.group_info:
        return fuzzy_info_params(
            ctx=ctx,
            client=get_client,
            method="groups",
            key="name",
            multi=multi,
            opt_data=False
        )
    elif input == fuzzy_params_type.group_info_with_usernames:
        return fuzzy_info_params(
            ctx=ctx,
            client=get_client,
            opt_method="group_members",
            opt_key="user_group_id",
            method="groups",
            key="name",
            multi=multi,
            opt_data=True
        )
    elif input == fuzzy_params_type.user_info_and_group_membership:
        return fuzzy_info_params(
            ctx=ctx,
            client=get_client,
            opt_method="users_groups",
            opt_key="user_id",
            method="users",
            key="username",
            multi=multi,
            opt_data=True
        )


def fuzzy_info_wrapper(params):
    error = fuzzy_info(params)
    if GNS3Error.has_error(error):
        if error.connection:
            click.secho("Error: ", fg="red", nl=False, err=True)
            click.secho(
                "Failed to fetch data from the API check your Network connection to the server", bold=True, err=True)
            return
        GNS3Error.print_error(error)


def fuzzy_put_wrapper(params):
    error = fuzzy_change_password(params)
    if GNS3Error.has_error(error):
        if error.connection:
            click.secho("Error: ", fg="red", nl=False, err=True)
            click.secho(
                "Failed to fetch data from the API check your Network connection to the server", bold=True, err=True)
            return
        GNS3Error.print_error(error)


def execute_and_print(ctx, client, func):
    error, data = func(client)
    if GNS3Error.has_error(error):
        GNS3Error.print_error(error)
    else:
        rich.print_json(json.dumps(data, indent=2))


def get_role_id(ctx, name: str) -> (str, GNS3Error):
    get_roles_error, roles = call_client_method(ctx, "get", "roles")
    if GNS3Error.has_error(get_roles_error):
        return "", get_roles_error
    for role in roles:
        if role['name'] == name:
            return role['role_id'], get_roles_error


def create_project(ctx, name: str) -> (str, GNS3Error):
    project_id = str(uuid.uuid4())
    input_data = {
        "name": name, "project_id": project_id}
    create_project_error, result = call_client_method(
        ctx, "post", "create_project", input_data)
    if GNS3Error.has_error(create_project_error):
        return project_id, create_project_error
    close_project_error = close_project(ctx, project_id)
    if GNS3Error.has_error(close_project_error):
        return project_id, close_project_error
    return project_id, create_project_error


def close_project(ctx, project_id: str) -> GNS3Error:
    close_project_error, _ = call_client_method(
        ctx, "post", "close_project", project_id)
    if GNS3Error.has_error(close_project_error):
        return close_project_error
    return close_project_error


def get_groups_in_class(ctx, class_name: str) -> (list, GNS3Error):
    group_list = []
    get_groups_error, groups = call_client_method(ctx, "get", "groups")
    if GNS3Error.has_error(get_groups_error):
        return group_list, get_groups_error
    for group in groups:
        if class_name in group['name'] and class_name != group['name']:
            group_number = group['name'].split("-")[-1]
            group_dict = {
                "group_id": group["user_group_id"], "group_number": group_number}
            group_list.append(group_dict)

    return group_list, get_groups_error


def create_acl(ctx, params: create_acl_params) -> GNS3Error:
    if params.isGroup:
        input_data = {"ace_type": params.ace_type,
                      "allowed": params.allowed, "group_id": params.id, "path": params.path, "propagate": params.propagate, "role_id": params.role_id}
    else:
        input_data = {"ace_type": params.ace_type,
                      "allowed": params.allowed, "user_id": params.id, "path": params.path, "propagate": params.propagate, "role_id": params.role_id}
    create_acl_error, result = call_client_method(
        ctx, "post", "create_acl", input_data)
    return create_acl_error


def create_pool(ctx, pool_name: str) -> (str, GNS3Error):
    input_data = {"name": pool_name}
    create_pool_error, result = call_client_method(
        ctx, "post", "create_pool", input_data)
    if GNS3Error.has_error(create_pool_error):
        return "", create_pool_error
    return result['resource_pool_id'], create_pool_error


def add_resource_to_pool(ctx, pool_id: str, resource_id: str) -> (GNS3Error):
    add_to_pool_error, result = call_client_method(
        ctx, "put", "add_resource_to_pool", pool_id, resource_id)
    return add_to_pool_error


def create_Exercise(ctx, class_name: str, exercise_name: str) -> bool:
    role_id, get_role_id_error = get_role_id(ctx, "User")
    if GNS3Error.has_error(get_role_id_error):
        GNS3Error.print_error(get_role_id_error)
        return False

    groups, get_groups_error = get_groups_in_class(ctx, class_name)
    for group in groups:
        project_name = f"{class_name}-{exercise_name}-{group['group_number']}"
        project_id, create_project_error = create_project(ctx, project_name)
        if GNS3Error.has_error(create_project_error):
            GNS3Error.print_error(create_project_error)
            return False
        click.secho("Success: ", nl=False, fg="green")
        click.secho(f"created the project {project_name}")
        pool_name = project_name + "-pool"
        pool_id, create_pool_error = create_pool(ctx, pool_name)
        if GNS3Error.has_error(create_pool_error):
            GNS3Error.print_error(create_pool_error)
            return False
        add_to_pool_error = add_resource_to_pool(ctx, pool_id, project_id)
        if GNS3Error.has_error(add_to_pool_error):
            if add_to_pool_error.not_found:
                GNS3Error.print_error(
                    add_to_pool_error, "project id: "+project_id, "pool_id: missing")
            else:
                GNS3Error.print_error(add_to_pool_error)
            return False
        params = create_acl_params(
            ctx=ctx,
            ace_type="group",
            allowed=True,
            isGroup=True,
            id=group['group_id'],
            path=f"/pools/{pool_id}",
            propagate=True,
            role_id=role_id
        )
        create_acl_error = create_acl(ctx, params)
        if GNS3Error.has_error(create_acl_error):
            GNS3Error.print_error(create_acl_error)
            return False
        click.secho("Success: ", nl=False, fg="green")
        click.secho(f"created the acl for resource {params.path}")
    return True


def safe_json(resp):
    if resp.headers.get("Content-Length") == "0" or not resp.text:
        return None
    return resp.json()


def print_usernames_and_ids(ctx):
    error, users = call_client_method(ctx, "get", "users")
    if GNS3Error.has_error(error):
        GNS3Error.print_error(error)
    else:
        click.secho("List of all users and their id:", fg="green")
        for user in users:
            print_separator_with_secho()
            username = user.get('username', 'N/A')
            user_id = user.get('user_id', 'N/A')
            click.secho("Username: ", fg="cyan", nl=False)
            click.secho(f"{username}")
            click.secho("ID: ", fg="cyan", nl=False)
            click.secho(f"{user_id}")
            print_separator_with_secho()

def get_command_description(cmd: str, help_dict: dict, arg_type: str) -> tuple[str,str]:
    """
    Retrieves the description of a command from the help dictionary.

    Args:
        cmd (str): The command name to look up.
        help_dict (dict): The dictionary containing help metadata.
        arg_type (str): The key in help_dict corresponding to the argument count category
                        (e.g., "zero_arg", "one_arg", etc.).

    Returns:
        tuple[str,str]: The description and example of the command, or an empty string if not found.
    """
    current_help_option = ""
    epiloge = ""

    for key, value in help_dict[arg_type].items():
        if key == cmd:
            current_help_option = str(value["description"])
            epiloge = str(value["example"])
            break
    
    return current_help_option, epiloge
