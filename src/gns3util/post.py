import click
import json
import rich
import os
from . import auth
from .api.post_endpoints import GNS3PostAPI


@click.group()
def post():
    """Post commands."""
    pass


def get_client(ctx):
    """Helper function to create GNS3PostAPI instance."""
    key_file = os.path.expanduser("~/.gns3key")
    server_url = ctx.parent.obj['server']
    key = auth.loadKey(key_file)
    return GNS3PostAPI(server_url, key)


def execute_and_print(ctx, func):
    client = get_client(ctx)
    success, data = func(client)
    if success:
        rich.print_json(json.dumps(data, indent=2))


# Commands with no arguments
_zero_arg = {
    "check-version": "check_version",
    "user": "user"
}

# Create click commands with zero arguments
for cmd, func in _zero_arg.items():
    def make_cmd(func=func):
        @click.argument('json_data')
        @click.pass_context
        def cmd_func(ctx, json_data):
            try:
                data = json.loads(json_data)
                execute_and_print(
                    ctx, lambda client: getattr(client, func)(data))
            except json.JSONDecodeError:
                print("Error: Invalid JSON input")
                return
        return cmd_func
    post.command(name=cmd)(make_cmd())

_zero_arg_no_data = {
    "reload": "reload",
    "shutdown": "shutdown"
}

# Create click commands with zero arguments and no data
for cmd, func in _zero_arg_no_data.items():
    def make_cmd(func=func):
        @click.pass_context
        def cmd_func(ctx):
            execute_and_print(
                ctx, lambda client: getattr(client, func)())
        return cmd_func
    post.command(name=cmd)(make_cmd())

# Commands with one argument plus JSON
_one_arg = {
    # Will be populated as needed
}

# Create click commands with one argument plus JSON
for cmd, func in _one_arg.items():
    def make_cmd(func=func):
        @click.argument('arg')
        @click.argument('json_data')
        @click.pass_context
        def cmd_func(ctx, arg, json_data):
            try:
                data = json.loads(json_data)
                execute_and_print(ctx, lambda client: getattr(
                    client, func)(arg, data))
            except json.JSONDecodeError:
                print("Error: Invalid JSON input")
                return
        return cmd_func
    post.command(name=cmd)(make_cmd())

# Commands with two arguments plus JSON
_two_arg = {
    # Will be populated as needed
}

# Create click commands with two arguments plus JSON
for cmd, func in _two_arg.items():
    def make_cmd(func=func):
        @click.argument('arg1')
        @click.argument('arg2')
        @click.argument('json_data')
        @click.pass_context
        def cmd_func(ctx, arg1, arg2, json_data):
            try:
                data = json.loads(json_data)
                execute_and_print(ctx, lambda client: getattr(
                    client, func)(arg1, arg2, data))
            except json.JSONDecodeError:
                print("Error: Invalid JSON input")
                return
        return cmd_func
    post.command(name=cmd)(make_cmd())
