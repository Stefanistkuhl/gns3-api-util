import click
import json
import rich
import os
from . import auth
from .api.post_endpoints import GNS3PostAPI

"""
Number of arguments: 0
Has data: True
"""
_zero_arg = {
    "check-version": "version_check",
    "user": "user_create",
    "group": "group_create",
    "role": "role_create",
    "acl": "acl_create",
    "template": "template_create",
    "project": "project_create",
    "project_load": "project_load",
    "add_pool": "add_pool"
}

"""
Number of arguments: 0
Has data: False
"""
_zero_arg_no_data = {
    "reload": "node_reload",
    "shutdown": "controller_shutdown",
    "install_img": "image_install"
}

"""
Number of arguments: 1
Has data: True
"""
_one_arg = {
    "qemu_img": "qemu_image_create",
    "node": "node_create",
    "link_create": "link_create",
    "drawing_create": "drawing_create",
    "snapshot_create": "snapshot_create",
    "create_compute": "create_compute",
    "auto_idlepc": "auto_idlepc",
    "add_applience_version": "add_applience_version"
}

"""
Number of arguments: 1
Has data: False
"""
_one_arg_no_data = {
    "duplicate_template": "template_duplicate",
    "project_close": "project_close",
    "project_open": "project_open",
    "project_lock": "project_lock",
    "project_unlock": "project_unlock",
    "start_nodes": "node_start",
    "stop_nodes": "node_stop",
    "suspend_nodes": "node_suspend",
    "reload_nodes": "node_reload",
    "nodes_console_reset": "nodes_console_reset",
    "symbol_create": "symbol_create",
    "connect_compute": "connect_compute"
}

"""
Number of arguments: 2
Has data: False
"""
_two_arg_no_data = {
    "upload_img": "image_upload",
    "project_import": "project_import",
    "project_write_file": "project_file_write",
    "node_isolate": "node_isolate",
    "node_unisolate": "node_unisolate",
    "node_console_reset": "node_console_reset",
    "link_reset": "link_reset",
    "stop_link_capture": "stop_link_capture",
    "snapshot_restore": "snapshot_restore",
    "add_applience_version": "add_applience_version"
}


_two_arg = {
    "project_node_from_template": "project_node_create_from_template",
    "duplicate_node": "node_duplicate",
    "start_link_capture": "start_link_capture"
}

_three_arg = {
    "create_disk_img": "create_disk_img"
}

_three_arg_no_data = {
    "node_create_file": "node_create_file"
}


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

# Create click commands with zero arguments and no data
for cmd, func in _zero_arg_no_data.items():
    def make_cmd(func=func):
        @click.pass_context
        def cmd_func(ctx):
            execute_and_print(
                ctx, lambda client: getattr(client, func)())
        return cmd_func
    post.command(name=cmd)(make_cmd())

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

# Create click commands with one argument minus JSON
for cmd, func in _one_arg_no_data.items():
    def make_cmd(func=func):
        @click.argument('arg')
        @click.pass_context
        def cmd_func(ctx, arg):
            execute_and_print(ctx, lambda client: getattr(
                client, func)(arg))
        return cmd_func
    post.command(name=cmd)(make_cmd())

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
    post.command(name=cmd)(make_cmd())

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
    post.command(name=cmd)(make_cmd())

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
    post.command(name=cmd)(make_cmd())
