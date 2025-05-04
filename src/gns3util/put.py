import click
import os
import json
from . import auth
from .api.put_endpoints import GNS3PutAPI
from .utils import execute_and_print, fuzzy_password_params, fuzzy_put_wrapper

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
Number of arguments: 2
Has data: False
"""
_two_arg_no_data = {
    "add_group_member": "add_group_member",
    "add_ressouce_to_pool": "add_resource_to_pool",
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


@click.group()
def put():
    """put commands."""
    pass


def get_client(ctx):
    """Helper function to create GNS3PutAPI instance."""
    server_url = ctx.parent.obj['server']
    _, key = auth.load_and_try_key(ctx)
    return GNS3PutAPI(server_url, key['access_token'])


# Create click commands with zero arguments
for cmd, func in _zero_arg.items():
    def make_cmd(func=func):
        @click.argument('json_data')
        @click.pass_context
        def cmd_func(ctx, json_data):
            api_put_client = get_client(ctx)
            try:
                data = json.loads(json_data)
                execute_and_print(
                    ctx, api_put_client, lambda client: getattr(api_put_client, func)(data))
            except json.JSONDecodeError:
                click.secho("Error: Invalid JSON input", err=True, fg="red")
                return
        return cmd_func
    put.command(name=cmd)(make_cmd())

# Create click commands with one argument plus JSON
for cmd, func in _one_arg.items():
    def make_cmd(func=func):
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
                click.secho("Error: Invalid JSON input", err=True, fg="red")
                return
        return cmd_func
    put.command(name=cmd)(make_cmd())


# Create click commands with two arguments minus JSON
for cmd, func in _two_arg_no_data.items():
    def make_cmd(func=func):
        @click.argument('arg1')
        @click.argument('arg2')
        @click.pass_context
        def cmd_func(ctx, arg1, arg2):
            api_put_client = get_client(ctx)
            execute_and_print(ctx, api_put_client, lambda client: getattr(
                api_put_client, func)(arg1, arg2))
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
            api_put_client = get_client(ctx)
            try:
                data = json.loads(json_data)
                execute_and_print(ctx, api_put_client, lambda client: getattr(
                    api_put_client, func)(arg1, arg2, data))
            except json.JSONDecodeError:
                click.secho("Error: Invalid JSON input", err=True, fg="red")
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
            api_put_client = get_client(ctx)
            try:
                data = json.loads(json_data)
                execute_and_print(ctx, api_put_client, lambda client: getattr(
                    api_put_client, func)(arg1, arg2, arg3, data))
            except json.JSONDecodeError:
                click.secho("Error: Invalid JSON input", err=True, fg="red")
                return
        return cmd_func
    put.command(name=cmd)(make_cmd())


@put.command(name="fchpw", help="find user info using fzf and change their password")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_user_info_and_groups_short(ctx, multi):
    params = fuzzy_password_params(
        ctx=ctx,
        client=get_client,
        method="users",
        key="username",
        multi=multi,
    )
    fuzzy_put_wrapper(params)
