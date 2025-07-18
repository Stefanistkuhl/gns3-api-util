import click
import json
import importlib.resources

from . import auth
from .api.post_endpoints import GNS3PostAPI
from .utils import (
    execute_and_print,
    create_class,
    create_Exercise,
    get_command_description,
)
from .server import start_and_get_data

# endpoint definitions

_zero_arg_no_data = {
    "reload": "reload_node",
    "shutdown": "shutdown_controller",
    "install_img": "install_image",
}

_one_arg_no_data = {
    "duplicate_template": "duplicate_template",
    "project_close": "close_project",
    "project_open": "open_project",
    "project_lock": "lock_project",
    "project_unlock": "unlock_project",
    "start_nodes": "start_nodes",
    "stop_nodes": "stop_nodes",
    "suspend_nodes": "suspend_nodes",
    "reload_nodes": "reload_nodes",
    "nodes_console_reset": "reset_nodes_console",
    "connect_compute": "connect_compute",
}

_two_arg_no_data = {
    "upload_img": "upload_image",
    "project_import": "import_project",
    "project_write_file": "write_project_file",
    "node_isolate": "isolate_node",
    "start_node": "start_node",
    "stop_node": "stop_node",
    "suspend_node": "suspend_node",
    "reload_node": "reload_node",
    "node_unisolate": "unisolate_node",
    "node_console_reset": "reset_node_console",
    "reset_link": "reset_link",
    "stop_link_capture": "stop_link_capture",
    "snapshot_restore": "restore_snapshot",
    "add_applience_version": "install_appliance_version",
}

_two_arg = {
    "duplicate_node": "duplicate_node",
    "start_link_capture": "start_link_capture",
}

# load help texts
with importlib.resources.files("gns3util.help_texts") \
        .joinpath("help_post.json") \
        .open("r", encoding="utf-8") as f:
    help_dict = json.load(f)


@click.group()
def post():
    """Misc post commands."""
    pass


def get_client(ctx: click.Context):
    """Helper function to create GNS3PostAPI instance."""
    server_url = ctx.parent.obj['server']
    verify = ctx.parent.obj['verify']
    success, key = auth.load_and_try_key(ctx)
    if success and key:
        return GNS3PostAPI(server_url, key.access_token, verify=verify)
    else:
        ctx.exit(1)

# define sub‐command groups


@post.group()
def controller():
    """Controller operations."""
    pass


@post.group()
def project():
    """Project operations."""
    pass


@post.group()
def node():
    """Node operations."""
    pass


@post.group()
def image():
    """Image operations."""
    pass


@post.group()
def link():
    """Link operations."""
    pass


@post.group()
def snapshot():
    """Snapshot operations."""
    pass


@post.group()
def compute():
    """Compute operations."""
    pass


# helper to attach a cmd to a group
def attach(group, name, func, help_text, epilog):
    group.command(name=name, help=help_text, epilog=epilog)(func)


# zero‐arg, no data
for cmd, api_name in _zero_arg_no_data.items():
    help_txt, ep = get_command_description(cmd, help_dict, "zero_arg_no_data")

    def make_cmd(api_name=api_name):
        @click.pass_context
        def _cmd(ctx: click.Context):
            client = get_client(ctx)
            execute_and_print(ctx, client, lambda c: getattr(c, api_name)())
        return _cmd

    grp = controller if cmd in ("reload", "shutdown") else image
    attach(grp, cmd, make_cmd(), help_txt, ep)


# one‐arg, no data
for cmd, api_name in _one_arg_no_data.items():
    help_txt, ep = get_command_description(cmd, help_dict, "one_arg_no_data")

    def make_cmd(api_name=api_name):
        @click.argument("arg")
        @click.pass_context
        def _cmd(ctx: click.Context, arg):
            client = get_client(ctx)
            execute_and_print(ctx, client, lambda c: getattr(c, api_name)(arg))
        return _cmd

    if cmd.startswith("project_") or cmd == "duplicate_template":
        grp = project
    elif cmd.endswith("_nodes") or cmd.endswith("nodes_reset"):
        grp = node
    else:
        grp = compute
    attach(grp, cmd, make_cmd(), help_txt, ep)


# two‐arg, no data
for cmd, api_name in _two_arg_no_data.items():
    help_txt, ep = get_command_description(cmd, help_dict, "two_arg_no_data")

    def make_cmd(api_name=api_name):
        @click.argument("arg1")
        @click.argument("arg2")
        @click.pass_context
        def _cmd(ctx: click.Context, arg1, arg2):
            client = get_client(ctx)
            execute_and_print(ctx, client, lambda c: getattr(
                c, api_name)(arg1, arg2))
        return _cmd

    if cmd.startswith("project_"):
        grp = project
    elif cmd.startswith("node_"):
        grp = node
    elif cmd.startswith("upload_img") or cmd == "add_applience_version":
        grp = image
    elif cmd.startswith("reset_link") or cmd.endswith("capture"):
        grp = link
    elif cmd.startswith("snapshot_restore"):
        grp = snapshot
    else:
        grp = image
    attach(grp, cmd, make_cmd(), help_txt, ep)


# two‐arg + JSON
for cmd, api_name in _two_arg.items():
    help_txt, ep = get_command_description(cmd, help_dict, "two_arg")

    def make_cmd(api_name=api_name):
        @click.argument("arg1")
        @click.argument("arg2")
        @click.argument("json_data")
        @click.pass_context
        def _cmd(ctx: click.Context, arg1, arg2, json_data):
            client = get_client(ctx)
            try:
                data = json.loads(json_data)
            except json.JSONDecodeError:
                click.secho("Error: Invalid JSON input",
                            fg="red", bold=True, err=True)
                return
            execute_and_print(ctx, client, lambda c: getattr(
                c, api_name)(arg1, arg2, data))
        return _cmd

    grp = node if cmd == "duplicate_node" else link
    attach(grp, cmd, make_cmd(), help_txt, ep)


# misc: class & exercise under top‐level
@post.command(name="class", help="Setup a class and its students")
@click.argument("filename", required=False, type=click.Path(exists=True))
@click.option(
    "-c",
    "--create",
    is_flag=True,
    help="Launch a webpage to enter class data interactively",
)
@click.pass_context
def make_class(ctx: click.Context, filename, create):
    if not filename and not create:
        click.secho("Use -c or provide a JSON file", fg="red")
        return

    if create:
        data = start_and_get_data(host="localhost", port=8080, debug=True)
        if not data:
            click.secho("No data received", fg="red")
            return
        class_name, ok = create_class(ctx, None, data)
    else:
        class_name, ok = create_class(ctx, filename)

    if ok:
        click.secho(f"Class '{class_name}' created", fg="green")
    else:
        click.secho(f"Failed to create class '{
                    class_name}'", fg="red", err=True)


@post.command(name="exercise", help="Create an exercise for a class")
@click.argument("class_name")
@click.argument("exercise_name")
@click.pass_context
def make_exercise(ctx: click.Context, class_name, exercise_name):
    ok = create_Exercise(ctx, class_name, exercise_name)
    if ok:
        click.secho(f"Exercise '{exercise_name}' created", fg="green")
    else:
        click.secho(f"Failed to create exercise '{
                    exercise_name}'", fg="red", err=True)


@post.command(
    help="Check server version against provided JSON data",
    epilog="Example: gns3util -s [server] post check_version '{...}'",
)
@click.argument("json_data")
@click.pass_context
def check_version(ctx: click.Context, json_data):
    client = get_client(ctx)
    try:
        data = json.loads(json_data)
    except json.JSONDecodeError:
        click.secho("Error: Invalid JSON", fg="red", bold=True, err=True)
        return
    execute_and_print(ctx, client, lambda c: c.check_version(data))


@post.command(
    help="Create or authenticate a user with JSON data",
    epilog="Example: gns3util -s [server] post user_authenticate '{...}'",
)
@click.argument("json_data")
@click.pass_context
def user_authenticate(ctx: click.Context, json_data):
    client = get_client(ctx)
    try:
        data = json.loads(json_data)
    except json.JSONDecodeError:
        click.secho("Error: Invalid JSON", fg="red", bold=True, err=True)
        return
    execute_and_print(ctx, client, lambda c: c.user_authenticate(data))
