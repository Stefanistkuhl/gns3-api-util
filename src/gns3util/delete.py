import json
import click
import os
import importlib.resources
from . import auth
from .api.delete_endpoints import GNS3DeleteAPI
from .utils import execute_and_print, get_command_description, fuzzy_delete_class_params, fuzzy_delete_class_wrapper

"""
Number of arguments: 0
Has data: False
"""
_zero_arg_no_data = {
    "prune_images": "prune_images"
}

"""
Number of arguments: 1
Has data: False
"""
_one_arg_no_data = {
    "user": "delete_user",
    "compute": "delete_compute",
    "project": "delete_project",
    "template": "delete_template",
    "image": "delete_image",
    "acl": "delete_acl",
    "role": "delete_role",
    "group": "delete_group",
    "pool": "delete_pool"
}

"""
Number of arguments: 2
Has data: False
"""
_two_arg_no_data = {
    "pool_resource": "delete_pool_resource",
    "link": "delete_link",
    "node": "delete_node",
    "drawing": "delete_drawing",
    "role_priv": "delete_role_priv",
    "user_from_group": "delete_user_from_group",
    "snapshot": "delete_snapshot"
}


@click.group()
def delete():
    """Delete commands."""
    pass


def get_client(ctx):
    """Helper function to create GNS3DeleteAPI instance."""
    server_url = ctx.parent.obj['server']
    success, key = auth.load_and_try_key(ctx)
    if success:
        return GNS3DeleteAPI(server_url, key['access_token'])
    else:
        os._exit(1)

with importlib.resources.files("gns3util.help_texts").joinpath("help_delete.json").open("r", encoding="utf-8") as f:
    help_dict = json.load(f)

# Create click commands with zero arguments and no data
for cmd, func in _zero_arg_no_data.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "zero_arg")

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.pass_context
        def cmd_func(ctx):
            api_delete_client = get_client(ctx)
            execute_and_print(
                ctx, api_delete_client, lambda client: getattr(api_delete_client, func)())
        return cmd_func
    delete.command(name=cmd, help=current_help_option,
                   epilog=epiloge)(make_cmd())

# Create click commands with one argument minus JSON
for cmd, func in _one_arg_no_data.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "one_arg")

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.argument('arg')
        @click.pass_context
        def cmd_func(ctx, arg):
            api_delete_client = get_client(ctx)
            execute_and_print(ctx, api_delete_client, lambda client: getattr(
                api_delete_client, func)(arg))
        return cmd_func
    delete.command(name=cmd, help=current_help_option,
                   epilog=epiloge)(make_cmd())

# Create click commands with two arguments minus JSON
for cmd, func in _two_arg_no_data.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "two_arg")

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.argument('arg1')
        @click.argument('arg2')
        @click.pass_context
        def cmd_func(ctx, arg1, arg2):
            api_delete_client = get_client(ctx)
            execute_and_print(ctx, api_delete_client, lambda client: getattr(
                api_delete_client, func)(arg1, arg2))
        return cmd_func
    delete.command(name=cmd, help=current_help_option,
                   epilog=epiloge)(make_cmd())


@delete.command(name="class", help="delete a class and it's studets")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.option(
    "-y", "--confirm", default=True, is_flag=True, help="Require confirmation for deletion"
)
@click.option(
    "-n", "--non_interactive", help="Run the command non interactively"
)
@click.pass_context
def fuzzy_delete_class(ctx, multi, confirm, non_interactive):
    if non_interactive:
        pass
    else:
        params = fuzzy_delete_class_params(
            ctx=ctx,
            client=get_client,
            method="groups",
            key="name",
            multi=multi,
            confirm=confirm)
        fuzzy_delete_class_wrapper(params)
