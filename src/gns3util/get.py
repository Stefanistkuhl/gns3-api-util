import click
import sys
import rich
import json
import os
from . import auth
from .api.get_endpoints import GNS3GetAPI
from .utils import fzf_select

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


# TODO refactor all fo this to have a general fuzzy functions and instead of the bool for the meber toggles have a flag

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


def fuzzy_user_info(ctx, show_groups=bool):
    users_raw = get_client(ctx).users()
    if not users_raw[0]:
        sys.exit("An error occurred getting all the data for the users")
    users_data = users_raw[1]
    users = []
    for user in users_data:
        users.append(user['username'])
    selected = fzf_select(users)
    for user in users_data:
        if user['username'] in selected:
            print(f"{GREY}---{RESET}")
            for key, value in user.items():
                print(f"{CYAN}{key}{RESET}: {value}")
            print(f"{GREY}---{RESET}")
            if show_groups:
                groups_raw = get_client(ctx).users_groups(user['user_id'])
                if not groups_raw[0]:
                    sys.exit(
                        "An error occurred getting all the data for the groups members")
                groups = groups_raw[1]
                if groups == []:
                    sys.exit("this user is in no groups")
                for group in groups:
                    print(f"{GREY}---{RESET}")
                    for key, value in group.items():
                        print(f"{CYAN}{key}{RESET}: {value}")
                    print(f"{GREY}---{RESET}")


def fuzzy_group_info(ctx, show_members=bool):
    groups_raw = get_client(ctx).groups()
    if not groups_raw[0]:
        sys.exit("An error occurred getting all the data for the groups")
    groups_data = groups_raw[1]
    groups = []
    for group in groups_data:
        groups.append(group['name'])
    selected = fzf_select(groups)
    for group in groups_data:
        if group['name'] in selected:
            print(f"{GREY}---{RESET}")
            for key, value in group.items():
                print(f"{CYAN}{key}{RESET}: {value}")
            print(f"{GREY}---{RESET}")
    if show_members:
        for group in groups_data:
            if group['name'] in selected:
                group_members_raw = get_client(
                    ctx).group_members(group['user_group_id'])
                if not group_members_raw[0]:
                    sys.exit(
                        "An error occurred getting all the data for the groups members")
                group_members = group_members_raw[1]
                if group_members == []:
                    sys.exit("this group has no members")
                for group_member in group_members:
                    print(f"{GREY}---{RESET}")
                    for key, value in group_member.items():
                        print(f"{CYAN}{key}{RESET}: {value}")
                    print(f"{GREY}---{RESET}")


@get.command(name="find-user-info", help="find user info using fzf")
@click.pass_context
def find_user_info(ctx):
    fuzzy_user_info(ctx)


@get.command(name="fui", help="find user info using fzf")
@click.pass_context
def find_user_info_command_short(ctx):
    fuzzy_user_info(ctx)


@get.command(name="find-group-info", help="find group info using fzf")
@click.pass_context
def find_group_info(ctx):
    fuzzy_group_info(ctx)


@get.command(name="fgi", help="find group info using fzf")
@click.pass_context
def find_group_info_command_short(ctx):
    fuzzy_group_info(ctx)


@get.command(name="find-group-info-with-usernames", help="find group info with members using fzf")
@click.pass_context
def find_group_info_with_members(ctx):
    fuzzy_group_info(ctx, True)


@get.command(name="fgim", help="find group info with members using fzf")
@click.pass_context
def find_group_info_with_members_command_short(ctx):
    fuzzy_group_info(ctx, True)


@get.command(name="find-user-info-and-group-membership", help="find user info and group membership using fzf")
@click.pass_context
def find_user_info_and_groups(ctx):
    fuzzy_user_info(ctx)


@get.command(name="fuig", help="find user info and group membership using fzf")
@click.pass_context
def find_user_info_and_groups_short(ctx):
    fuzzy_user_info(ctx)
