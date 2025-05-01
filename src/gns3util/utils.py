import json
import uuid
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


def add_user_to_group(ctx, user_id: str, group_id: str) -> GNS3Error:
    from . import put
    client = put.get_client(ctx)
    add_user_to_group_error, result = client.add_group_member(
        group_id, user_id)
    if GNS3Error.has_error(add_user_to_group_error):
        return add_user_to_group_error
    return add_user_to_group_error


def create_class(ctx, filename: str):
    error_load, data = parse_json(filename)

    if error_load:
        click.echo(
            f"Failed to load the file {filename}. Error: {data}", err=True
        )
        return

    class_name = list(data.keys())[0]
    class_obj = data[class_name]
    class_id, create_group_error = create_user_group(ctx, class_name)
    if GNS3Error.has_error(create_group_error):
        print("handle this later")
        return
    for group_name, group_obj in class_obj.items():
        group_id, create_group_error = create_user_group(ctx, group_name)
        if GNS3Error.has_error(create_group_error):
            GNS3Error.print_error(create_group_error)
            return
        students = group_obj["students"]
        for student in students:
            user_id, create_user_error = create_user(ctx, student)
            if GNS3Error.has_error(create_user_error):
                GNS3Error.print_error(create_user_error)
                return
            add_user_to_class_error = add_user_to_group(ctx, user_id, class_id)
            if GNS3Error.has_error(add_user_to_class_error):
                GNS3Error.print_error(add_user_to_class_error)
            add_user_to_group_error = add_user_to_group(ctx, user_id, group_id)
            if GNS3Error.has_error(add_user_to_group_error):
                GNS3Error.print_error(add_user_to_group_error)


def create_user_group(ctx, group_name) -> (str, GNS3Error):
    from . import post
    error = GNS3Error()
    input_data = {"name": group_name}
    client = post.get_client(ctx)
    create_group_error, result = client.create_group(input_data)
    if GNS3Error.has_error(create_group_error):
        return "", error
    click.echo(f"Successfully created the group {group_name}")
    return result['user_group_id'], error


def create_user(ctx, user_dict: dict) -> (str, GNS3Error):
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
<<<<<<< HEAD
        return "", error
    click.echo(f"Successfully created the user {input_data["username"]}")
    return result['user_id'], error
=======
        return error
    click.echo(f"Successfully created the user {input_data["username"]}")
    return error
>>>>>>> 530457d (feat: creating acls and adding group deletion to the user managment)


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


<<<<<<< HEAD
def get_role_id(ctx, name: str) -> (str, GNS3Error):
    from . import get
    client = get.get_client(ctx)
    get_roles_error, roles = client.roles()
    if GNS3Error.has_error(get_roles_error):
        return get_roles_error
    for role in roles:
        if role['name'] == name:
            return role['role_id'], get_roles_error


def create_project(ctx, name: str) -> (str, GNS3Error):
    from . import post
    project_id = str(uuid.uuid4())
=======
def create_project(ctx, name: str) -> GNS3Error:
    from . import post
    project_id = str(uuid.uuid4())
    error = GNS3Error()
>>>>>>> b63b16b (feat: project creation for exercies)
    input_data = {
        "name": name, "project_id": project_id}
    client = post.get_client(ctx)
    create_project_error, result = client.create_project(input_data)
    if GNS3Error.has_error(create_project_error):
<<<<<<< HEAD
        return project_id, create_project_error
    close_project_error = close_project(ctx, project_id)
    if GNS3Error.has_error(close_project_error):
        GNS3Error.print_error(close_project_error)
        return project_id, close_project_error
    click.echo(f"Successfully created the project {input_data["name"]}")
    return project_id, create_project_error


def close_project(ctx, project_id: str) -> GNS3Error:
    from . import post
    client = post.get_client(ctx)
    close_project_error, _ = client.close_project(project_id)
    if GNS3Error.has_error(close_project_error):
        return close_project_error
    return close_project_error
=======
        return create_project_error
    click.echo(f"Successfully created the project {input_data["name"]}")
    return create_project_error
>>>>>>> b63b16b (feat: project creation for exercies)


def get_groups_in_class(ctx, class_name: str) -> (list, GNS3Error):
    from . import get
    group_list = []
    error = GNS3Error()
    client = get.get_client(ctx)
    get_groups_error, groups = client.groups()
    if GNS3Error.has_error(get_groups_error):
        return group_list, get_groups_error
    for group in groups:
        if class_name in group['name'] and class_name != group['name']:
            group_number = group['name'].split("-")[-1]
            group_dict = {
                "group_id": group["user_group_id"], "group_number": group_number}
            group_list.append(group_dict)

    return group_list, error


<<<<<<< HEAD
def create_acl(ctx, params: create_acl_params) -> GNS3Error:
    from . import post
    client = post.get_client(ctx)
    if params.isGroup:
        input_data = {"ace_type": params.ace_type,
                      "allowed": params.allowed, "group_id": params.id, "path": params.path, "propagate": params.propagate, "role_id": params.role_id}
    else:
        input_data = {"ace_type": params.ace_type,
                      "allowed": params.allowed, "user_id": params.id, "path": params.path, "propagate": params.propagate, "role_id": params.role_id}

    create_acl_error, result = client.create_acl(input_data)
    if GNS3Error.has_error(create_acl_error):
        return create_acl_error
    click.echo(f"Successfully created the acl for project {params.path}")
    return create_acl_error


<<<<<<< HEAD
def create_pool(ctx, pool_name: str) -> (str, GNS3Error):
    from . import post
    client = post.get_client(ctx)
    input_data = {"name": pool_name}
    create_pool_error, result = client.create_pool(input_data)
    if GNS3Error.has_error(create_pool_error):
        return "", create_pool_error
    return result['resource_pool_id'], create_pool_error


def add_resource_to_pool(ctx, pool_id: str, resource_id: str) -> (GNS3Error):
    from . import put
    client = put.get_client(ctx)
    print("pool id:", pool_id)
    add_to_pool_error, result = client.add_resource_to_pool(
        pool_id, resource_id)
    if GNS3Error.has_error(add_to_pool_error):
        return add_to_pool_error
    return add_to_pool_error


=======
>>>>>>> 530457d (feat: creating acls and adding group deletion to the user managment)
def create_Exercise(ctx, class_name: str, exercise_name: str) -> bool:
    sucess = True
    role_id, get_role_id_error = get_role_id(ctx, "User")
    if GNS3Error.has_error(get_role_id_error):
        GNS3Error.print_error(get_role_id_error)
        sucess = False

    groups, get_groups_error = get_groups_in_class(ctx, class_name)
    for group in groups:
        project_name = f"{class_name}-{exercise_name}-{group['group_number']}"
        project_id, create_project_error = create_project(ctx, project_name)
        if GNS3Error.has_error(create_project_error):
            sucess = False
            GNS3Error.print_error(create_project_error)
        params = create_acl_params(
            ctx=ctx,
            ace_type="group",
            allowed=True,
            isGroup=True,
            id=group['group_id'],
            path=f"/projects/{project_id}",
            propagate=True,
            role_id=role_id
        )
        create_acl_error = create_acl(ctx, params)
        if GNS3Error.has_error(create_acl_error):
            sucess = False
            GNS3Error.print_error(create_acl_error)
    if sucess:
        click.echo(
            f"Exercise {exercise_name} and it's acls created sucessfully")
    return sucess


def safe_json(resp):
    if resp.headers.get("Content-Length") == "0" or not resp.text:
        return None
    return resp.json()
=======
def create_Exercise(ctx, class_name: str, exercise_name: str):
    groups, get_groups_error = get_groups_in_class(ctx, class_name)
    print(groups)
    for group in groups:
        project_name = f"{class_name}-{exercise_name}-{group['group_number']}"
        create_project_error = create_project(ctx, project_name)
        print(create_project_error)
>>>>>>> b63b16b (feat: project creation for exercies)
