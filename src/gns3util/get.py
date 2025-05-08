import click
import os
from . import auth
from .api.get_endpoints import GNS3GetAPI
from .api import GNS3Error
from .utils import fzf_select, fuzzy_info, fuzzy_info_params, fuzzy_info_wrapper, execute_and_print, print_separator_with_secho, print_usernames_and_ids, get_fuzzy_info_params, fuzzy_params_type, get_command_description
import json

get = click.Group('get')

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
    "images": "images",
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
    server_url = ctx.parent.obj['server']
    _, key = auth.load_and_try_key(ctx)
    return GNS3GetAPI(server_url, key['access_token'])


help_path = os.path.join(os.getcwd(), "src", "gns3util", "help_texts", "help_get.json")
with open(help_path, "r") as f:
    help_dict = json.load(f)

# Create click commands with zero arguments
for cmd, func in _zero_arg.items():
    current_help_option,epiloge = get_command_description(cmd, help_dict, "zero_arg")
    def make_cmd(func=func, help_option=current_help_option,epilog=epiloge):
        @click.pass_context
        def cmd_func(ctx):
            api_get_client = get_client(ctx)
            execute_and_print(
                ctx, api_get_client, lambda client: getattr(api_get_client, func)())
        return cmd_func
    get.command(name=cmd, help=current_help_option,epilog=epiloge)(make_cmd())

# Create click commands with one argument
for cmd, func in _one_arg.items():
    current_help_option,epiloge = get_command_description(cmd, help_dict, "one_arg")
    def make_cmd(func=func, help_option=current_help_option,epilog=epiloge):
        @click.argument('arg')
        @click.pass_context
        def cmd_func(ctx, arg):
            api_get_client = get_client(ctx)
            execute_and_print(
                ctx, api_get_client, lambda client: getattr(api_get_client, func)(arg))
        return cmd_func
    get.command(name=cmd, help=current_help_option,epilog=epiloge)(make_cmd())

# Create click commands with two arguments
for cmd, func in _two_arg.items():
    current_help_option,epiloge = get_command_description(cmd, help_dict, "two_arg")
    def make_cmd(func=func, help_option=current_help_option,epilog=epiloge):
        @click.argument('project_id')
        @click.argument('id')
        @click.pass_context
        def cmd_func(ctx, project_id, id):
            api_get_client = get_client(ctx)
            execute_and_print(ctx, api_get_client, lambda client: getattr(
                api_get_client, func)(project_id, id))
        return cmd_func
    get.command(name=cmd, help=current_help_option,epilog=epiloge)(make_cmd())

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
    print_usernames_and_ids(ctx)


@get.command(name="uai", help="Listing all users and their ids")
@click.pass_context
def usernames_and_ids_short(ctx):
    print_usernames_and_ids(ctx)


@get.command(name="find-user-info", help="find user info using fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_user_info(ctx, multi):
    params = get_fuzzy_info_params(
        fuzzy_params_type.user_info, ctx, get_client, multi)
    fuzzy_info_wrapper(params)


@get.command(name="fui", help="find user info using fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_user_info_command_short(ctx, multi):
    params = get_fuzzy_info_params(
        fuzzy_params_type.user_info, ctx, get_client, multi)
    fuzzy_info_wrapper(params)


@get.command(name="find-group-info", help="find group info using fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_group_info(ctx, multi):
    params = get_fuzzy_info_params(
        fuzzy_params_type.group_info, ctx, get_client, multi)
    fuzzy_info_wrapper(params)


@get.command(name="fgi", help="find group info using fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_group_info_command_short(ctx, multi):
    params = get_fuzzy_info_params(
        fuzzy_params_type.group_info, ctx, get_client, multi)
    fuzzy_info_wrapper(params)


@get.command(name="find-group-info-with-usernames", help="find group info with members using fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_group_info_with_members(ctx, multi):
    params = get_fuzzy_info_params(
        fuzzy_params_type.group_info_with_usernames, ctx, get_client, multi)
    fuzzy_info_wrapper(params)


@get.command(name="fgim", help="find group info with members using fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_group_info_with_members_command_short(ctx, multi):
    params = get_fuzzy_info_params(
        fuzzy_params_type.group_info_with_usernames, ctx, get_client, multi)
    fuzzy_info_wrapper(params)


@get.command(name="find-user-info-and-group-membership", help="find user info and group membership using fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_user_info_and_groups(ctx, multi):
    params = get_fuzzy_info_params(
        fuzzy_params_type.user_info_and_group_membership, ctx, get_client, multi)
    fuzzy_info_wrapper(params)


@get.command(name="fuig", help="find user info and group membership using fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_user_info_and_groups_short(ctx, multi):
    params = get_fuzzy_info_params(
        fuzzy_params_type.user_info_and_group_membership, ctx, get_client, multi)
    fuzzy_info_wrapper(params)
