import click
import uuid
import json
import importlib.resources
from . import auth
from .api.post_endpoints import GNS3PostAPI
from .utils import (
    execute_and_print,
    get_command_description,
    resolve_ids,
    get_data_for_update,
    is_valid_uuid,
    call_client_data,
    close_project,
)
from gns3util.schemas import (
    Version,
    Credentials,
    ProjectDuplicate,
    Project,
    NodeDuplicate,
    AutoIdlePC,
)
from pydantic import ValidationError


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


# load help texts
with (
    importlib.resources.files("gns3util.help_texts")
    .joinpath("help_post.json")
    .open("r", encoding="utf-8") as f
):
    help_dict = json.load(f)


@click.group()
def post():
    """Misc post commands."""
    pass


def get_client(ctx: click.Context):
    """Helper function to create GNS3PostAPI instance."""
    server_url = ctx.parent.obj["server"]
    verify = ctx.parent.obj["verify"]
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
            execute_and_print(ctx, client, lambda c: getattr(c, api_name)(arg1, arg2))

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



@post.command(
    help="Check server version against provided data",
    epilog="Example: gns3util -s [server] post check_version --version 3.0.5",
)
@click.option(
    "-ch",
    "--controller-host",
    type=str,
    default=None,
    help="Host to use to check, leave empty to use the server that you set with `-s/--server`",
)
@click.option(
    "-v",
    "--version",
    type=str,
    help="Version to check against.",
)
@click.option(
    "-l",
    "--local",
    is_flag=True,
    default=None,
    help="Wheater to use the local controller or not.",
)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def check_version(ctx: click.Context, controller_host, version, local, use_json):
    if not version and not use_json:
        raise click.UsageError(
            "For this command at least the -v or -j option is required."
        )

    if not use_json:
        try:
            version_data = Version(
                controller_host=controller_host,
                version=version,
                local=local,
            )
        except ValidationError as e:
            click.secho("Invalid input data", err=True)
            ctx.exit(1)
        data = json.loads(version_data.model_dump_json())
    else:
        try:
            data = json.loads(use_json)
        except json.JSONDecodeError:
            click.secho("Error: Invalid JSON", fg="red", bold=True, err=True)
            ctx.exit(1)

    client = get_client(ctx)
    execute_and_print(ctx, client, lambda c: c.check_version(data))


@post.command(
    help="Authenticate as a user",
    epilog="Example: gns3util -s [server] post user_authenticate --user alice --password 1234",
)
@click.option(
    "-u",
    "--user",
    type=str,
    default=None,
    envvar="GNS3_USER",
    help="desired user to authenticate as. (env: GNS3_USER)",
)
@click.option(
    "-p",
    "--password",
    type=str,
    default=None,
    envvar="GNS3_PASSWORD",
    help="password for that user. (env: GNS3_PASSWORD)",
)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def user_authenticate(ctx: click.Context, user, password, use_json):
    if not user and not password and not use_json:
        raise click.UsageError(
            "For this command the -u and -p options or the -j option is required."
        )

    if not use_json:
        try:
            data = Credentials(
                username=user,
                password=password,
            )
        except ValidationError as e:
            click.secho("Invalid input data", err=True)
            ctx.exit(1)
        data = json.loads(data.model_dump_json())
    else:
        try:
            data = json.loads(use_json)
        except json.JSONDecodeError:
            click.secho("Error: Invalid JSON", fg="red", bold=True, err=True)
            ctx.exit(1)

    client = get_client(ctx)
    execute_and_print(ctx, client, lambda c: c.user_authenticate(data))


@project.command(
    help="Load a project from a given path.",
    epilog='Example: gns3util -s [server] post load-project -p "/opt/gns3/projects/id"/project filename.gns3',
    name="load",
)
@click.option(
    "-p",
    "--path",
    type=str,
    default=None,
)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def load_project(ctx: click.Context, path, use_json):
    if not path and not use_json:
        raise click.UsageError(
            "For this command the -p option or  the -j option is required."
        )

    if not use_json:
        data = {"path": path}
    else:
        try:
            data = json.loads(use_json)
        except json.JSONDecodeError:
            click.secho("Error: Invalid JSON", fg="red", bold=True, err=True)
            ctx.exit(1)

    client = get_client(ctx)
    execute_and_print(ctx, client, lambda c: c.load_project(data))

@project.command(
    help="Duplicate a Project.",
    epilog="Example: gns3util -s [server] duplicate project -n some_name ID/Name",
    name="duplicate"
)
@click.option(
    "-n",
    "--name",
    type=str,
    default=None,
    help="Desired name for the project",
)
@click.option(
    "-id",
    "--project-id",
    type=str,
    default=None,
    help="Desired id for the project, leave empty for a generated one",
)
@click.option(
    "-p",
    "--path",
    type=str,
    default=None,
    help="Filepath for the project.",
)
@click.option(
    "-ac",
    "--auto-close",
    is_flag=True,
    default=True,
    help="Close project when last client leaves. Default: True",
)
@click.option(
    "-ao",
    "--auto-open",
    is_flag=True,
    default=False,
    help="Project opens when GNS3 starts. Default: False",
)
@click.option(
    "-as",
    "--auto-start",
    is_flag=True,
    default=False,
    help="Project starts when opened. Default: False",
)
@click.option(
    "-sh",
    "--scene-height",
    type=int,
    default=None,
    help="Height of the drawing area.",
)
@click.option(
    "-sw",
    "--scene-width",
    type=int,
    default=None,
    help="Width of the drawing area.",
)
@click.option(
    "-z",
    "--zoom",
    type=int,
    default=None,
    help="Zoom of the drawing area.",
)
@click.option(
    "-sl",
    "--show-layers",
    is_flag=True,
    default=False,
    help="Show layers on the drawing area. Default: False",
)
@click.option(
    "-sg",
    "--snap-to-grid",
    is_flag=True,
    default=False,
    help="Snap to grid on the drawing area. Default: False",
)
@click.option(
    "-shg",
    "--show-grid",
    is_flag=True,
    default=False,
    help="Show the grid on the drawing area. Default: False",
)
@click.option(
    "-gz",
    "--grid-size",
    type=int,
    default=None,
    help="Grid size for the drawing area for nodes.",
)
@click.option(
    "-dgz",
    "--drawing-grid-size",
    type=int,
    default=None,
    help="Grid size for the drawing area for drawings.",
)
@click.option(
    "-si",
    "--show-interface-labels",
    is_flag=True,
    default=False,
    help="Show interface labels on the drawing area. Default: False",
)
@click.option(
    "-supl",
    "--supplier-logo",
    type=str,
    default=None,
    help="Path to the project supplier logo.",
)
@click.option(
    "-su",
    "--supplier-url",
    type=str,
    default=None,
    help="URL to the project supplier site.",
)
@click.option(
    "-rm",
    "--reset-mac-addresses",
    is_flag=True,
    default=False,
    help="Reset MAC addresses for this project.",
)
@click.option(
    "-caf",
    "--close-after-creation",
    is_flag=True,
    default=True,
    help="Wheater to close the project after it's creation. Default: True",
)
@click.argument("project-id", required=True, type=str)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def duplicate_project(
    ctx: click.Context,
    name,
    path,
    auto_close,
    auto_open,
    auto_start,
    scene_height,
    scene_width,
    zoom,
    show_layers,
    snap_to_grid,
    show_grid,
    grid_size,
    drawing_grid_size,
    show_interface_labels,
    supplier_logo,
    supplier_url,
    reset_mac_addresses,
    close_after_creation,
    project_id,
    use_json,
):
    if (
        not name
        and not project_id
        and not path
        and not auto_close
        and not auto_open
        and not auto_start
        and not scene_height
        and not scene_width
        and not zoom
        and not show_layers
        and not snap_to_grid
        and not show_grid
        and not grid_size
        and not drawing_grid_size
        and not show_interface_labels
        and not supplier_logo
        and not supplier_url
        and not reset_mac_addresses
        and not use_json
    ):
        raise click.UsageError(
            "For this command the -n option is required or the -j option on it's own."
        )
    if name is None:
        raise click.UsageError(
            "For this command the -n option is required or the -j option on it's own."
        )

    if project_id and not is_valid_uuid(project_id):
        project_id, ok = resolve_ids(ctx, "project", project_id)
        if not ok:
            click.secho(f"{project_id}", err=True)
            ctx.exit(1)

    original_project: Project = get_data_for_update(
        call_client_data(ctx=ctx, package="get", method="projects"), project_id)
    if not original_project:
        click.secho("Error while getting the data of the orignal ACE.")
        ctx.exit(1)

    args = [("name", name), ("project_id", project_id),("path",path), ("auto_close",auto_close), ("auto_open", auto_open) ,("auto_start", auto_start), ("scene_height", scene_height), ("scene_width", scene_width), ("zoom",zoom), ("show_layers", show_layers), ("snap_to_grid", snap_to_grid), ("show_grid", show_grid), ("grid_size", grid_size), ("drawing_grid_size", drawing_grid_size), ("show_interface_labels", show_interface_labels)]

    for arg in args:
        if arg[1]:
            setattr(original_project, arg[0], arg[1])

    if supplier_logo:
        if original_project.supplier:
            original_project.supplier.logo = supplier_logo
    if supplier_url:
        if original_project.supplier:
            original_project.supplier.url = supplier_url

    if not use_json:
        try:
            data = ProjectDuplicate(
                name=original_project.name,
                project_id=uuid.uuid4(),
                path=original_project.path,
                auto_close=original_project.auto_close,
                auto_open=original_project.auto_open,
                auto_start=original_project.auto_start,
                scene_height=original_project.scene_height,
                scene_width=original_project.scene_width,
                zoom=original_project.zoom,
                show_layers=original_project.show_layers,
                snap_to_grid=original_project.snap_to_grid,
                show_grid=original_project.show_grid,
                grid_size=original_project.grid_size,
                drawing_grid_size=original_project.drawing_grid_size,
                show_interface_labels=original_project.show_interface_labels,
                supplier=original_project.supplier,
                variables=original_project.variables,
                reset_mac_addresses=reset_mac_addresses,
            )
        except ValidationError as e:
            click.secho("Invalid input data", err=True)
            ctx.exit(1)
        data = json.loads(data.model_dump_json())
    else:
        try:
            data = json.loads(use_json)
        except json.JSONDecodeError:
            click.secho("Error: Invalid JSON", fg="red", bold=True, err=True)
            ctx.exit(1)

    client = get_client(ctx)
    execute_and_print(ctx, client, lambda c: c.duplicate_project(project_id, data))
    if close_after_creation:
        close_project(ctx, str(project_id))

@node.command(
    help="Duplicate a Node in a Project.",
    epilog="Example: gns3util -s [server] post duplicate-node -x 10 -y 10 PROJECT_ID/Name NODE_ID/Name",
    name="duplicate",
)
@click.option(
    "-x",
    "--x",
    type=int,
    default=None,
    help="X-Position of the node.",
)
@click.option(
    "-y",
    "--y",
    type=int,
    default=None,
    help="Y-Position of the node.",
)
@click.option(
    "-z",
    "--z",
    type=int,
    default=1,
    help="Z-Position (layer) of the node. Default: 1",
)
@click.argument("project-id", required=True, type=str)
@click.argument("node-id", required=True, type=str)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def node(
    ctx: click.Context,
    x,
    y,
    z,
    project_id,
    node_id,
    use_json,
):
    if (
        not x
        and not y
        and not use_json
    ):
        raise click.UsageError(
            "For this command the -x and -y options are required or the -j option on it's own."
        )

    if project_id and not is_valid_uuid(project_id):
        project_id, ok = resolve_ids(ctx, "project", project_id)
        if not ok:
            click.secho(f"{project_id}", err=True)
            ctx.exit(1)

    if node_id and not is_valid_uuid(node_id):
        node_id, ok = resolve_ids(ctx, "node", node_id, [project_id])
        if not ok:
            click.secho(f"{node_id}", err=True)
            ctx.exit(1)

    if not use_json:
        try:
            data = NodeDuplicate(
                x=x,
                y=y,
                z=z,
            )
        except ValidationError as e:
            click.secho("Invalid input data", err=True)
            ctx.exit(1)
        data = json.loads(data.model_dump_json())
    else:
        try:
            data = json.loads(use_json)
        except json.JSONDecodeError:
            click.secho("Error: Invalid JSON", fg="red", bold=True, err=True)
            ctx.exit(1)

    client = get_client(ctx)
    execute_and_print(ctx, client, lambda c: c.duplicate_node(project_id, node_id, data))


@project.command(
    help="Start a packet caputre in a project on a given link.",
    epilog="Example: gns3util -s [server] post start-capture PROJECT_ID/Name LINK_ID",
    name="start-capture",
)
@click.option(
    "-p",
    "--properties",
    type=dict,
    default={},
    help="Additional properties for the request.",
)
@click.argument("project-id", required=True, type=str)
@click.argument("link-id", required=True, type=str)
@click.pass_context
def node(
    ctx: click.Context,
    properties,
    project_id,
    link_id,
):

    if project_id and not is_valid_uuid(project_id):
        project_id, ok = resolve_ids(ctx, "project", project_id)
        if not ok:
            click.secho(f"{project_id}", err=True)
            ctx.exit(1)

    client = get_client(ctx)
    execute_and_print(ctx, client, lambda c: c.start_link_capture(project_id, link_id, properties))



@image.command(
    help="Find a suitable Idle-PC value for a given IOS image. This may take a few minutes.",
    epilog="Example: gns3util -s [server] create auto-idle-pc -p c7200 -i /path/to/image -r 256",
    name="auto-idle-pc"
)
@click.option(
    "-p",
    "--platform",
    type=str,
    default=None,
    help="Cisco platform",
)
@click.option(
    "-i",
    "--image",
    type=str,
    default=None,
    help="Image path",
)
@click.option(
    "-r",
    "--ram",
    type=click.IntRange(min=0, max=65535, clamp=False),
    default=None,
    help="Amount of RAM in MB.",
)
@click.argument("compute-id", required=True, type=str)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def auto_idle_pc(
    ctx: click.Context,
    platform,
    image,
    ram,
    compute_id,
    use_json,
):
    if not platform and image and ram and not use_json:
        raise click.UsageError(
            "For this command the -p, -i and -r options are required or the -j option on it's own."
        )

    if compute_id and not is_valid_uuid(compute_id):
        compute_id, ok = resolve_ids(ctx, "compute", compute_id)
        if not ok:
            click.secho(f"{compute_id}", err=True)
            ctx.exit(1)

    if not use_json:
        try:
            data = AutoIdlePC(
                platform=platform,
                image=image,
                ram=ram
            )
        except ValidationError as e:
            click.secho("Invalid input data", err=True)
            ctx.exit(1)
        data = json.loads(data.model_dump_json())
    else:
        try:
            data = json.loads(use_json)
        except json.JSONDecodeError:
            click.secho("Error: Invalid JSON", fg="red", bold=True, err=True)
            ctx.exit(1)

    client = get_client(ctx)
    execute_and_print(ctx, client, lambda c: c.set_auto_idlepc(compute_id, data))
