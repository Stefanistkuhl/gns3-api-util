import click
import sys
import rich
import json
import os
from . import auth
from .api.get_endpoints import GNS3GetAPI
from .utils import fzf_select, fuzzy_info, fuzzy_info_params, fuzzy_info_wrapper

get = click.Group('get')

GREY = "\033[90m"
CYAN = "\033[96m"
RESET = "\033[0m"

# Commands with no arguments
_zero_arg = {
    "version": "version",
    "iou-license": "iou_license",
    "statistics": "statistics",
    "me": "current_user_info",
    "users": "users",
    "projects": "projects",
    "groups": "groups",
    "roles": "roles",
    "privileges": "privileges",
    "acl-endpoints": "acl_endpoints",
    "acl": "acl",
    "templates": "templates",
    "symbols": "symbols",
    "default-symbols": "default_symbols",
    "computes": "computes",
    "appliances": "appliances",
    "pools": "pools"
}

# Commands with one argument
_one_arg = {
    "user": "user",
    "user-groups": "users_groups",
    "project": "project",
    "project-stats": "project_stats",
    "project-locked": "project_locked",
    "group": "groups_by_id",
    "group-members": "group_members",
    "role": "role_by_id",
    "role-privileges": "role_privileges",
    "template": "template_by_id",
    "compute": "compute_by_id",
    "docker-images": "compute_by_id_docker_images",
    "virtualbox-vms": "compute_by_id_virtualbox_vms",
    "vmware-vms": "compute_by_id_vmware_vms",
    "images": "images",
    "images_by_path": "images_by_path",
    "snapshots": "snapshots",
    "appliance": "appliance",
    "pool": "pool",
    "pool-resources": "pool_resources",
    "drawings": "drawings",
    "symbol": "symbol",
    "acl-rule": "acl_by_id",
    "links": "links",
    "nodes": "nodes"
}

# Commands with two arguments (assumed: project_id and id)
_two_arg = {
    "node": "node_by_id",
    "node-links": "node_links_by_id",
    "link": "link",
    "link-filters": "link_filters",
    "drawing": "drawing"
}


def get_client(ctx):
    """Helper function to create GNS3GetAPI instance."""
    key_file = os.path.expanduser("~/.gns3key")
    server_url = ctx.parent.obj['server']
    key = auth.load_key(key_file)  # updated from loadKey to load_key
    return GNS3GetAPI(server_url, key)


def execute_and_print(ctx, func):
    client = get_client(ctx)
    success, data = func(client)
    if success:
        rich.print_json(json.dumps(data, indent=2))


# Create click commands with zero arguments
for cmd, func in _zero_arg.items():
    def make_cmd(func=func):
        @click.pass_context
        def cmd_func(ctx):
            execute_and_print(ctx, lambda client: getattr(client, func)())
        return cmd_func
    get.command(name=cmd)(make_cmd())

# Create click commands with one argument
for cmd, func in _one_arg.items():
    def make_cmd(func=func):
        @click.argument('arg')
        @click.pass_context
        def cmd_func(ctx, arg):
            execute_and_print(ctx, lambda client: getattr(client, func)(arg))
        return cmd_func
    get.command(name=cmd)(make_cmd())

# Create click commands with two arguments
for cmd, func in _two_arg.items():
    def make_cmd(func=func):
        @click.argument('project_id')
        @click.argument('id')
        @click.pass_context
        def cmd_func(ctx, project_id, id):
            execute_and_print(ctx, lambda client: getattr(
                client, func)(project_id, id))
        return cmd_func
    get.command(name=cmd)(make_cmd())

# Special commands with timeout options


@get.command()
@click.option('--timeout', '-t', 'timeout_seconds', default=60, help='Notification stream timeout in seconds')
@click.pass_context
def notifications(ctx, timeout_seconds):
    get_client(ctx).notifications(timeout_seconds)


@get.command(name="project-id")
@click.argument('project_id')
@click.option('--timeout', '-t', 'timeout_seconds', default=60, help='Notification stream timeout in seconds')
@click.pass_context
def project_notifications(ctx, project_id, timeout_seconds):
    get_client(ctx).project_notifications(project_id, timeout_seconds)


@get.command(name="usernames-and-ids", help="Listing all users and their ids")
@click.pass_context
def usernames_and_ids(ctx):
    users_raw = get_client(ctx).users()
    if not users_raw[0]:
        sys.exit("An error occurred getting all the data for the users")
    users = users_raw[1]
    print("List of all users and their id:")
    for user in users:
        username = user.get('username', 'N/A')
        user_id = user.get('user_id', 'N/A')
        print(f"Username: {username}")
        print(f"ID: {user_id}")
        print("-" * 10)


@get.command(name="find-user-info", help="find user info using fzf")
@click.pass_context
def find_user_info(ctx):
    params = fuzzy_info_params(
        ctx=ctx,
        client=get_client,
        method="users",
        key="username",
        multi=True,
        opt_data=False
    )
    fuzzy_info_wrapper(params)


@get.command(name="fui", help="find user info using fzf")
@click.pass_context
def find_user_info_command_short(ctx):
    params = fuzzy_info_params(
        ctx=ctx,
        client=get_client,
        method="users",
        key="username",
        multi=True,
        opt_data=False
    )
    fuzzy_info_wrapper(params)


@get.command(name="find-group-info", help="find group info using fzf")
@click.pass_context
def find_group_info(ctx):
    params = fuzzy_info_params(
        ctx=ctx,
        client=get_client,
        method="groups",
        key="name",
        multi=True,
        opt_data=False
    )
    fuzzy_info_wrapper(params)


@get.command(name="fgi", help="find group info using fzf")
@click.pass_context
def find_group_info_command_short(ctx):
    params = fuzzy_info_params(
        ctx=ctx,
        client=get_client,
        method="groups",
        key="name",
        multi=True,
        opt_data=False
    )
    fuzzy_info_wrapper(params)


@get.command(name="find-group-info-with-usernames", help="find group info with members using fzf")
@click.pass_context
def find_group_info_with_members(ctx):
    params = fuzzy_info_params(
        ctx=ctx,
        client=get_client,
        opt_method="group",
        opt_key="user_group_id",
        method="groups",
        key="name",
        multi=True,
        opt_data=True
    )
    fuzzy_info_wrapper(params)


@get.command(name="fgim", help="find group info with members using fzf")
@click.pass_context
def find_group_info_with_members_command_short(ctx):
    params = fuzzy_info_params(
        ctx=ctx,
        client=get_client,
        opt_method="group_members",
        opt_key="user_group_id",
        method="groups",
        key="name",
        multi=True,
        opt_data=True
    )
    fuzzy_info_wrapper(params)


@get.command(name="find-user-info-and-group-membership", help="find user info and group membership using fzf")
@click.pass_context
def find_user_info_and_groups(ctx):
    params = fuzzy_info_params(
        ctx=ctx,
        client=get_client,
        opt_method="group_members",
        opt_key="user_group_id",
        method="groups",
        key="name",
        multi=True,
        opt_data=True
    )
    fuzzy_info_wrapper(params)


@get.command(name="fuig", help="find user info and group membership using fzf")
@click.pass_context
def find_user_info_and_groups_short(ctx):
    params = fuzzy_info_params(
        ctx=ctx,
        client=get_client,
        opt_method="users_groups",
        opt_key="user_id",
        method="users",
        key="username",
        multi=True,
        opt_data=True
    )
    fuzzy_info_wrapper(params)
