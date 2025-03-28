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
    "check-version": "check_version",
    "user": "create_user",
    "group": "create_group",
    "role": "create_role",
    "acl": "create_acl",
    "template": "create_template",
    "project": "create_project",
    "project_load": "load_project",
    "add_pool": "create_pool"
}

"""
Number of arguments: 0
Has data: False
"""
_zero_arg_no_data = {
    "reload": "reload_node",
    "shutdown": "shutdown_controller",
    "install_img": "install_image"
}

"""
Number of arguments: 1
Has data: True
"""
_one_arg = {
    "qemu_img": "create_qemu_image",
    "node": "create_node",
    "link_create": "create_link",
    "drawing_create": "create_drawing",
    "snapshot_create": "create_snapshot",
    "create_compute": "create_compute",
    "auto_idlepc": "set_auto_idlepc",
    "add_applience_version": "create_appliance_version"
}

"""
Number of arguments: 1
Has data: False
"""
_one_arg_no_data = {
    "duplicate_template": "duplicate_template",
    "project_close": "close_project",
    "project_open": "open_project",
    "project_lock": "lock_project",
    "project_unlock": "unlock_project",
    "start_nodes": "start_node",
    "stop_nodes": "stop_node",
    "suspend_nodes": "suspend_node",
    "reload_nodes": "reload_node",
    "nodes_console_reset": "reset_nodes_console",
    "symbol_create": "create_symbol",
    "connect_compute": "connect_compute"
}

"""
Number of arguments: 2
Has data: False
"""
_two_arg_no_data = {
    "upload_img": "upload_image",
    "project_import": "import_project",
    "project_write_file": "write_project_file",
    "node_isolate": "isolate_node",
    "node_unisolate": "unisolate_node",
    "node_console_reset": "reset_node_console",
    "reset_link": "reset_link",
    "stop_link_capture": "stop_link_capture",
    "snapshot_restore": "restore_snapshot",
    "add_applience_version": "install_appliance_version"
}


_two_arg = {
    "project_node_from_template": "create_project_node_from_template",
    "duplicate_node": "duplicate_node",
    "start_link_capture": "start_link_capture"
}

_three_arg = {
    "create_disk_img": "create_disk_image"
}

_three_arg_no_data = {
    "create_node_file": "create_node_file"
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
