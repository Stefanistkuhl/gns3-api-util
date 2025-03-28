import click
import json
import rich
import os
from . import auth
from .api.put_endpoints import GNS3PutAPI

"""
Number of arguments: 0
Has data: True
"""
_zero_arg = {
    "iou_license": "iou_license",
    "me": "me"
}

"""
Number of arguments: 0
Has data: False
"""
_zero_arg_no_data = {

}

"""
Number of arguments: 1
Has data: True
"""
_one_arg = {
    "update_user": "update_user",
    "update_group": "update_group",
    "update_acl": "update_ace",
    "update_template": "update_template",
    "update_project": "update_project",
    "update_compute": "update_compute",
    "update_pool": "update_pool",
    "update_role": "update_role"

}

"""
Number of arguments: 1
Has data: False
"""
_one_arg_no_data = {

}

"""
Number of arguments: 2
Has data: False
"""
_two_arg_no_data = {
    "add_group_member": "add_group_member",
    "add_ressouce_to_pool": "add_ressouce_to_pool",
    "update_role_privs": "update_role_privs"
}


_two_arg = {
    "update_node": "update_node",
    "update_drawing": "update_drawing",
    "update_link": "update_link"

}

_three_arg = {
    "update_disk_image": "update_disk_image"

}

_three_arg_no_data = {

}


@click.group()
def put():
    """put commands."""
    pass


def get_client(ctx):
    """Helper function to create GNS3PutAPI instance."""
    key_file = os.path.expanduser("~/.gns3key")
    server_url = ctx.parent.obj['server']
    key = auth.loadKey(key_file)
    return GNS3PutAPI(server_url, key)


def execute_and_print(ctx, func):
    client = get_client(ctx)
    success, data = func(client)
    if success:
        rich.print_json(json.dumps(data, indent=2))


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
    put.command(name=cmd)(make_cmd())

# Create click commands with zero arguments and no data
for cmd, func in _zero_arg_no_data.items():
    def make_cmd(func=func):
        @click.pass_context
        def cmd_func(ctx):
            execute_and_print(
                ctx, lambda client: getattr(client, func)())
        return cmd_func
    put.command(name=cmd)(make_cmd())

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
    put.command(name=cmd)(make_cmd())

# Create click commands with one argument minus JSON
for cmd, func in _one_arg_no_data.items():
    def make_cmd(func=func):
        @click.argument('arg')
        @click.pass_context
        def cmd_func(ctx, arg):
            execute_and_print(ctx, lambda client: getattr(
                client, func)(arg))
        return cmd_func
    put.command(name=cmd)(make_cmd())

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
    put.command(name=cmd)(make_cmd())

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
    put.command(name=cmd)(make_cmd())

# Create click commands with three arguments plus JSON
for cmd, func in _three_arg.items():
    def make_cmd(func=func):
        @click.argument('arg1')
        @click.argument('arg2')
        @click.argument('arg3')
        @click.argument('json_data')
        @click.pass_context
        def cmd_func(ctx, arg1, arg2, arg3, json_data):
            try:
                data = json.loads(json_data)
                execute_and_print(ctx, lambda client: getattr(
                    client, func)(arg1, arg2, arg3, data))
            except json.JSONDecodeError:
                print("Error: Invalid JSON input")
                return
        return cmd_func
    put.command(name=cmd)(make_cmd())

# Create click commands with two arguments minus JSON
for cmd, func in _three_arg_no_data.items():
    def make_cmd(func=func):
        @click.argument('arg1')
        @click.argument('arg2')
        @click.argument('arg3')
        @click.pass_context
        def cmd_func(ctx, arg1, arg2, arg3):
            execute_and_print(ctx, lambda client: getattr(
                client, func)(arg1, arg2, arg3))
        return cmd_func
    put.command(name=cmd)(make_cmd())
