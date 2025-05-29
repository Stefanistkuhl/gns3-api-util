import click
import os
import json
import importlib.resources
from . import auth
from .api.put_endpoints import GNS3PutAPI
from .utils import execute_and_print, get_command_description

"""
Number of arguments: 0
Has data: True
"""
_zero_arg = {
    "iou_license": "iou_license",
    "me": "me"
}

"""
Number of arguments: 1
Has data: True
"""
_one_arg = {
    "user": "update_user",
    "group": "update_group",
    "acl": "update_ace",
    "template": "update_template",
    "project": "update_project",
    "compute": "update_compute",
    "pool": "update_pool",
    "role": "update_role"

}

"""
Number of arguments: 2
Has data: False
"""
_two_arg_no_data = {
    "role_privs": "update_role_privs"
}

_two_arg = {
    "node": "update_node",
    "drawing": "update_drawing",
    "link": "update_link"

}

_three_arg = {
    "disk_image": "update_disk_image"

}


@click.group()
def update():
    """Put commands."""
    pass


def get_client(ctx):
    """Helper function to create GNS3PutAPI instance."""
    server_url = ctx.parent.obj['server']
    verify = ctx.parent.obj['verify']
    success, key = auth.load_and_try_key(ctx)
    if success:
        return GNS3PutAPI(server_url, key['access_token'], verify)
    else:
        os._exit(1)


with importlib.resources.files("gns3util.help_texts").joinpath("help_put.json").open("r", encoding="utf-8") as f:
    help_dict = json.load(f)

# Create click commands with zero arguments
for cmd, func in _zero_arg.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "zero_arg")

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.argument('json_data')
        @click.pass_context
        def cmd_func(ctx, json_data):
            api_put_client = get_client(ctx)
            try:
                data = json.loads(json_data)
                execute_and_print(
                    ctx, api_put_client, lambda client: getattr(api_put_client, func)(data))
            except json.JSONDecodeError:
                click.secho("Error: ", nl=True, fg="red", err=True)
                click.secho("Invalid JSON input", bold=True, err=True)
                return
        return cmd_func
    update.command(name=cmd, help=current_help_option,
                   epilog=epiloge)(make_cmd())

# Create click commands with one argument plus JSON
for cmd, func in _one_arg.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "one_arg")

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.argument('arg')
        @click.argument('json_data')
        @click.pass_context
        def cmd_func(ctx, arg, json_data):
            api_put_client = get_client(ctx)
            try:
                data = json.loads(json_data)
                execute_and_print(ctx, api_put_client, lambda client: getattr(
                    api_put_client, func)(arg, data))
            except json.JSONDecodeError:
                click.secho("Error: ", nl=True, fg="red", err=True)
                click.secho("Invalid JSON input", bold=True, err=True)
                return
        return cmd_func
    update.command(name=cmd, help=current_help_option,
                   epilog=epiloge)(make_cmd())


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
    update.command(name=cmd, help=current_help_option,
                   epilog=epiloge)(make_cmd())

# Create click commands with two arguments plus JSON
for cmd, func in _two_arg.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "two_arg")

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.argument('arg1')
        @click.argument('arg2')
        @click.argument('json_data')
        @click.pass_context
        def cmd_func(ctx, arg1, arg2, json_data):
            api_put_client = get_client(ctx)
            try:
                data = json.loads(json_data)
                execute_and_print(ctx, api_put_client, lambda client: getattr(
                    api_put_client, func)(arg1, arg2, data))
            except json.JSONDecodeError:
                click.secho("Error: ", nl=True, fg="red", err=True)
                click.secho("Invalid JSON input", bold=True, err=True)
                return
        return cmd_func
    update.command(name=cmd, help=current_help_option,
                   epilog=epiloge)(make_cmd())

# Create click commands with three arguments plus JSON
for cmd, func in _three_arg.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "three_arg")

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.argument('arg1')
        @click.argument('arg2')
        @click.argument('arg3')
        @click.argument('json_data')
        @click.pass_context
        def cmd_func(ctx, arg1, arg2, arg3, json_data):
            api_put_client = get_client(ctx)
            try:
                data = json.loads(json_data)
                execute_and_print(ctx, api_put_client, lambda client: getattr(
                    api_put_client, func)(arg1, arg2, arg3, data))
            except json.JSONDecodeError:
                click.secho("Error: ", nl=True, fg="red", err=True)
                click.secho("Invalid JSON input", bold=True, err=True)
                return
        return cmd_func
    update.command(name=cmd, help=current_help_option,
                   epilog=epiloge)(make_cmd())
