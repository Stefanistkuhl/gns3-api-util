import click
import json
import rich
import os
from . import auth
from .api.delete_endpoints import GNS3DeleteAPI

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
    "compute": "compute",
    "project": "project",
    "template": "template",
    "image": "image",
    "acl": "acl",
    "role": "role",
    "group": "group",
    "pool": "pool"
}

"""
Number of arguments: 2
Has data: False
"""
_two_arg_no_data = {
    "pool_resource": "pool_resource",
    "link": "link",
    "node": "node",
    "drawing": "drawing",
    "role_priv": "role_priv",
    "user_from_group": "user_from_group",
    "snapshot": "snapshot"
}


@click.group()
def delete():
    """delete commands."""
    pass


def get_client(ctx):
    """Helper function to create GNS3deleteAPI instance."""
    key_file = os.path.expanduser("~/.gns3key")
    server_url = ctx.parent.obj['server']
    key = auth.loadKey(key_file)
    return GNS3DeleteAPI(server_url, key)


def execute_and_print(ctx, func):
    client = get_client(ctx)
    success, data = func(client)
    if success:
        rich.print_json(json.dumps(data, indent=2))


# Create click commands with zero arguments and no data
for cmd, func in _zero_arg_no_data.items():
    def make_cmd(func=func):
        @click.pass_context
        def cmd_func(ctx):
            execute_and_print(
                ctx, lambda client: getattr(client, func)())
        return cmd_func
    delete.command(name=cmd)(make_cmd())


# Create click commands with one argument minus JSON
for cmd, func in _one_arg_no_data.items():
    def make_cmd(func=func):
        @click.argument('arg')
        @click.pass_context
        def cmd_func(ctx, arg):
            execute_and_print(ctx, lambda client: getattr(
                client, func)(arg))
        return cmd_func
    delete.command(name=cmd)(make_cmd())

# Create click commands with two arguments minus JSON
for cmd, func in _two_arg_no_data.items():
    def make_cmd(func=func):
        @click.argument('arg1')
        @click.argument('arg2')
        @click.pass_context
        def cmd_func(ctx, arg1, arg2):
            execute_and_print(ctx, lambda client: getattr(
                client, func)(arg1, arg2))
        return cmd_func
    delete.command(name=cmd)(make_cmd())
