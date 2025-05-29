import click
import os
import json
import importlib.resources
from . import auth
from .api.put_endpoints import GNS3PutAPI
from .utils import execute_and_print, get_command_description

"""
Number of arguments: 2
Has data: False
"""
_two_arg_no_data = {
    "group_member": "add_group_member",
    "ressouce_to_pool": "add_resource_to_pool",
}


@click.group()
def add():
    """Put commands."""
    pass


def get_client(ctx):
    """Helper function to create GNS3PutAPI instance."""
    server_url = ctx.parent.obj['server']
    success, key = auth.load_and_try_key(ctx)
    if success:
        return GNS3PutAPI(server_url, key['access_token'])
    else:
        os._exit(1)


with importlib.resources.files("gns3util.help_texts").joinpath("help_put.json").open("r", encoding="utf-8") as f:
    help_dict = json.load(f)


# Create click commands with two arguments minus JSON
for cmd, func in _two_arg_no_data.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "two_arg_no_data")

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.argument('arg1')
        @click.argument('arg2')
        @click.pass_context
        def cmd_func(ctx, arg1, arg2):
            api_put_client = get_client(ctx)
            execute_and_print(ctx, api_put_client, lambda client: getattr(
                api_put_client, func)(arg1, arg2))
        return cmd_func
    add.command(name=cmd, help=current_help_option, epilog=epiloge)(make_cmd())
