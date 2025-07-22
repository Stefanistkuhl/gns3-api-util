import click
import json
from . import auth
from .api.put_endpoints import GNS3PutAPI
from .utils import (
    execute_and_print,
    is_valid_uuid,
    call_client_data,
    get_data_for_update,
    resolve_ids,
)
from gns3util.schemas import (
    IOULicense,
    LoggedInUserUpdate,
    UserGroupUpdate,
    RoleUpdate,
    ACEUpdate,
    ACE,
    Project,
    ProjectUpdate,
    Node,
    NodeUpdate,
    ResourcePoolUpdate,
)
from pydantic import ValidationError


@click.group()
def update():
    """Put commands."""
    pass


def get_client(ctx: click.Context):
    """Helper function to create GNS3PutAPI instance."""
    server_url = ctx.parent.obj["server"]
    verify = ctx.parent.obj["verify"]
    success, key = auth.load_and_try_key(ctx)
    if success and key:
        return GNS3PutAPI(server_url, key.access_token, verify)
    else:
        ctx.exit(1)


@update.command(
    help="Update the IOULicense",
    epilog='Example: gns3util -s [server] update iou_license --iourc_content "some str"',
)
@click.option(
    "-ic", "--iourc_content", type=str, default=None, help="contents of the license"
)
@click.option(
    "-lc",
    "--license_check",
    is_flag=True,
    default=False,
    help="enable license checking",
)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def iou_license(ctx: click.Context, iourc_content, license_check, use_json):
    if not iou_license and not license_check and not use_json:
        raise click.UsageError(
            "For this command the -ic and -lc options are required or the -j on it's own."
        )
    if not use_json:
        try:
            data = IOULicense(
                iourc_content=iourc_content,
                license_check=license_check,
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
    execute_and_print(ctx, client, lambda c: c.iou_license(data))


@update.command(
    help="Update the logged in user",
    epilog="Example: gns3util -s [server] update me -p password -e something@exmaple.com -f someName",
)
@click.option(
    "-p",
    "--password",
    type=str,
    default=None,
    help="Password to set for the current User",
)
@click.option(
    "-e", "--email", type=str, default=None, help="Email to set for the current User"
)
@click.option(
    "-f",
    "--full-name",
    type=str,
    default=None,
    help="Full name to set for the current User",
)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def me(ctx: click.Context, password, email, full_name, use_json):
    if not password and not email and not full_name and not use_json:
        raise click.UsageError(
            "For this command the any of the -p, -e and -f options are required or the -j option on it's own."
        )
    if not use_json:
        try:
            data = LoggedInUserUpdate(
                password=password,
                email=email,
                full_name=full_name,
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
    execute_and_print(ctx, client, lambda c: c.me(data))


@update.command(
    help="Update a  given User with a given ID or name which will be resolved to a ID if a User with a matching name exists.",
    epilog="Example: gns3util -s [server] create user -u alice -p password [user-id]",
)
@click.argument("user-id", required=True, type=str)
@click.option(
    "-u", "--username", type=str, default=None, help="Desired username for the User"
)
@click.option(
    "-i",
    "--is-active",
    is_flag=True,
    default=False,
    help="Marking the user as currently active",
)
@click.option(
    "-e", "--email", type=str, default=None, help="Desired email for the user"
)
@click.option(
    "-f", "--full-name", type=str, default=None, help="Desired full name for the user"
)
@click.option(
    "-p", "--password", type=str, default=None, help="Desired password for the user"
)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def user(
    ctx: click.Context,
    user_id,
    username,
    is_active,
    email,
    full_name,
    password,
    use_json,
):
    if (
        not username
        and not is_active
        and not email
        and not full_name
        and not password
        and not use_json
    ):
        raise click.UsageError(
            "For this command any of the options required or the -j option on it's own."
        )

    if not is_valid_uuid(user_id):
        user_id, ok = resolve_ids(ctx, "user", user_id)
        if not ok:
            click.secho(f"{user_id}", err=True)
            ctx.exit(1)

    args = [
        ("username", username),
        ("is_active", is_active),
        ("email", email),
        ("full_name", full_name),
        ("password", password),
    ]

    if not use_json:
        data = {}
        for key, value in args:
            if value is not None:
                data[key] = value
    else:
        try:
            data = json.loads(use_json)
        except json.JSONDecodeError:
            click.secho("Error: Invalid JSON", fg="red", bold=True, err=True)
            ctx.exit(1)

    client = get_client(ctx)
    execute_and_print(ctx, client, lambda c: c.update_user(user_id, data))


@update.command(
    help="Update a group",
    epilog="Example: gns3util -s [server] update group -n some-name [group-id]",
)
@click.argument("group-id", required=True, type=str)
@click.option("-n", "--name", type=str, default=None, help="Desired name for the group")
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def group(ctx: click.Context, role_id, group_id, name, use_json):
    if not name and not use_json:
        raise click.UsageError(
            "For this command either the -n option is required or the -j option on it's own."
        )

    if not is_valid_uuid(group_id):
        group_id, ok = resolve_ids(ctx, "group", group_id)
        if not ok:
            click.secho(f"{group_id}", err=True)
            ctx.exit(1)

    if not use_json:
        try:
            data = UserGroupUpdate(
                name=name,
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
    execute_and_print(ctx, client, lambda c: c.update_group(group_id, data))


@update.command(
    help="Update a role",
    epilog="Example: gns3util -s [server] update role -n some-name [role-id]",
)
@click.argument("role-id", required=True, type=str)
@click.option("-n", "--name", type=str, default=None, help="Desired name for the role.")
@click.option(
    "-d",
    "--description",
    type=str,
    default=None,
    help="Desired description for the role.",
)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def role(ctx: click.Context, role_id, name, description, use_json):
    if not name and not description and not use_json:
        raise click.UsageError(
            "For this command either any option is required or the -j option on it's own."
        )

    if not is_valid_uuid(role_id):
        role_id, ok = resolve_ids(ctx, "role", role_id)
        if not ok:
            click.secho(f"{role_id}", err=True)
            ctx.exit(1)

    if not use_json:
        try:
            data = RoleUpdate(
                name=name,
                description=description,
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
    execute_and_print(ctx, client, lambda c: c.update_role(role_id, data))


@update.command(
    help="Update an ACE. user, group and role id will try to resolve the input to a valid id on the server if no valid UUIDv4 is given so names can be used instead of ids.",
    epilog="Example: gns3util -s [server] create ace -at user -p /[some endpoint]/[id] -r",
)
@click.argument("ace-id", required=True, type=str)
@click.option(
    "-at",
    "--ace-type",
    type=click.Choice(["user", "group"]),
    default=None,
    help="Desired type for the ACE.",
)
@click.option(
    "-p", "--path", type=str, default=None, help="Desired path for the ace to affect."
)
@click.option(
    "-pr",
    "--propagate",
    is_flag=True,
    default=True,
    help="Apply ACE rules to all nested endpoints in the path. Default: True",
)
@click.option(
    "-a",
    "--allow",
    is_flag=True,
    default=True,
    help="Wheater to allow or deny acces to the set path. Default: True",
)
@click.option(
    "-u",
    "--user-id",
    type=str,
    default=None,
    help="Desired user id to use for this ACE.",
)
@click.option(
    "-g",
    "--group-id",
    type=str,
    default=None,
    help="Desired group id to use for this ACE.",
)
@click.option(
    "-r",
    "--role-id",
    type=str,
    default=None,
    help="Desired role id to use for this ACE.",
)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def ace(
    ctx: click.Context,
    ace_id,
    ace_type,
    path,
    propagate,
    allow,
    user_id,
    group_id,
    role_id,
    use_json,
):
    if (
        not ace_type
        and not path
        and not user_id
        and not group_id
        and not role_id
        and not use_json
    ):
        raise click.UsageError(
            "For this command any of the options is required or the -j option on it's own."
        )

    if group_id and not is_valid_uuid(group_id):
        group_id, ok = resolve_ids(ctx, "group", group_id)
        if not ok:
            click.secho(f"{group_id}", err=True)
            ctx.exit(1)

    if user_id and not is_valid_uuid(user_id):
        user_id, ok = resolve_ids(ctx, "user", user_id)
        if not ok:
            click.secho(f"{user_id}", err=True)
            ctx.exit(1)

    if role_id and not is_valid_uuid(role_id):
        role_id, ok = resolve_ids(ctx, "role", role_id)
        if not ok:
            click.secho(f"{role_id}", err=True)
            ctx.exit(1)

    original_ace: ACE = get_data_for_update(
        call_client_data(ctx=ctx, package="get", method="acl"), ace_id
    )
    if not original_ace:
        click.secho("Error while getting the data of the orignal ACE.")
        ctx.exit(1)

    args = [
        ("ace_type", ace_type),
        ("path", path),
        ("propagate", propagate),
        ("allowed", allow),
        ("user_id", user_id),
        ("group_id", group_id),
        ("role_id", role_id),
    ]

    for arg in args:
        if arg[1]:
            setattr(original_ace, arg[0], arg[1])

    if not use_json:
        try:
            data = ACEUpdate(
                ace_type=original_ace.ace_type,
                path=original_ace.path,
                propagate=original_ace.propagate,
                allowed=original_ace.allowed,
                user_id=original_ace.user_id,
                group_id=original_ace.group_id,
                role_id=original_ace.role_id,
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
    execute_and_print(ctx, client, lambda c: c.update_ace(ace_id, data))


@update.command(
    help="Update a Project.",
    epilog="Example: gns3util -s [server] update project -n some_name ID/Name",
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
@click.argument("project-id", required=True, type=str)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def project(
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
        and not use_json
    ):
        raise click.UsageError(
            "For this command any of the options are required or the -j option on it's own."
        )

    if project_id and not is_valid_uuid(project_id):
        project_id, ok = resolve_ids(ctx, "project", project_id)
        if not ok:
            click.secho(f"{project_id}", err=True)
            ctx.exit(1)

    original_project: Project = get_data_for_update(
        call_client_data(ctx=ctx, package="get", method="projects"), project_id
    )
    if not original_project:
        click.secho("Error while getting the data of the orignal ACE.")
        ctx.exit(1)

    args = [
        ("name", name),
        ("project_id", project_id),
        ("path", path),
        ("auto_close", auto_close),
        ("auto_open", auto_open),
        ("auto_start", auto_start),
        ("scene_height", scene_height),
        ("scene_width", scene_width),
        ("zoom", zoom),
        ("show_layers", show_layers),
        ("snap_to_grid", snap_to_grid),
        ("show_grid", show_grid),
        ("grid_size", grid_size),
        ("drawing_grid_size", drawing_grid_size),
        ("show_interface_labels", show_interface_labels),
    ]

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
            data = ProjectUpdate(
                name=original_project.name,
                project_id=original_project.project_id,
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
    execute_and_print(ctx, client, lambda c: c.update_project(project_id, data))


@update.command(
    help="Update a Node in a Project. To use custom adapters the --use-json option has to be used.",
    epilog="Example: gns3util -s [server] update node -lt some_name  PROJECT_ID/Name NODE_ID/Name",
)
@click.option(
    "-c",
    "--compute-id",
    type=str,
    default="local",
    help='Compute on that the Node get\'s created. Default: "local"',
)
@click.option(
    "-n",
    "--name",
    type=str,
    default=None,
    help="Desired name for the Node.",
)
@click.option(
    "-nt",
    "--node_type",
    type=click.Choice(
        [
            "cloud",
            "nat",
            "ethernet_hub",
            "ethernet_switch",
            "frame_relay_switch",
            "atm_switch",
            "docker",
            "dynamips",
            "vpcs",
            "virtualbox",
            "vmware",
            "iou",
            "qemu",
        ]
    ),
    default=None,
    help="Type of Node.",
)
@click.option(
    "-ni",
    "--new-node-id",
    type=str,
    default=None,
    help="Desired ID for the node. If this is empty one will be generated.",
)
@click.option(
    "-cp",
    "--console-port",
    type=click.IntRange(min=0, max=65535, clamp=False),
    default=None,
    help="TCP port of the console.",
)
@click.option(
    "-ct",
    "--console-type",
    type=click.Choice(
        ["vnc", "telnet", "http", "https", "spice", "spice+agent", "none"]
    ),
    default=None,
    help="Type of the console interface.",
)
@click.option(
    "-ca",
    "--console-auto-start",
    is_flag=True,
    default=False,
    help="Automatically start the console when the node has started.",
)
@click.option(
    "-a",
    "--aux",
    type=click.IntRange(min=0, max=65535, clamp=False),
    default=None,
    help="Auxiliary console TCP port.",
)
@click.option(
    "-at",
    "--aux-type",
    type=click.Choice(
        ["vnc", "telnet", "http", "https", "spice", "spice+agent", "none"]
    ),
    default=None,
    help="Type of the aux console.",
)
@click.option(
    "-p",
    "--properties",
    type=dict,
    default=None,
    help="Custom properties for the emulator.",
)
@click.option(
    "-lt",
    "--label-text",
    type=str,
    default=None,
    help="Text of the label of the Node.",
)
@click.option(
    "-ls",
    "--label-style-attribute",
    type=str,
    default=None,
    help="SVG style attribute. Apply default style if null.",
)
@click.option(
    "-lx",
    "--label-x-position",
    type=int,
    default=None,
    help="X-Position of the label.",
)
@click.option(
    "-ly",
    "--label-y-position",
    type=int,
    default=None,
    help="Y-Position of the label.",
)
@click.option(
    "-lr",
    "--label-rotation",
    type=click.IntRange(min=-359, max=360, clamp=False),
    default=None,
    help="Rotation of the label.",
)
@click.option(
    "-s",
    "--symbol",
    type=str,
    default=None,
    help="Name of the desired symbol.",
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
@click.option(
    "-l",
    "--locked",
    is_flag=True,
    default=False,
    help="Whether the node is locked or not.",
)
@click.option(
    "-pf",
    "--port-name-format",
    type=str,
    default=None,
    help="Name format for the port for example: Ethernet{0}",
)
@click.option(
    "-ps",
    "--port-segment-size",
    type=int,
    default=None,
)
@click.option(
    "-fs",
    "--first-port-name",
    type=str,
    default=None,
    help="Name of the first port.",
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
    compute_id,
    name,
    node_type,
    new_node_id,
    console_port,
    console_type,
    console_auto_start,
    aux,
    aux_type,
    properties,
    label_text,
    label_style_attribute,
    label_x_position,
    label_y_position,
    label_rotation,
    symbol,
    x,
    y,
    z,
    locked,
    port_name_format,
    port_segment_size,
    first_port_name,
    project_id,
    node_id,
    use_json,
):
    if (
        not name
        and not node_type
        and not new_node_id
        and not console_port
        and not console_type
        and not console_auto_start
        and not aux
        and not aux_type
        and not properties
        and not label_text
        and not label_style_attribute
        and not label_x_position
        and not label_y_position
        and not label_rotation
        and not symbol
        and not x
        and not y
        and not port_name_format
        and not port_segment_size
        and not first_port_name
        and not use_json
    ):
        raise click.UsageError(
            "For this command any of the options are required or the -j option on it's own."
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

    if compute_id and not is_valid_uuid(compute_id) and compute_id != "local":
        compute_id, ok = resolve_ids(ctx, "compute", compute_id)
        if not ok:
            click.secho(f"{compute_id}", err=True)
            ctx.exit(1)

    original_node: Node = get_data_for_update(
        call_client_data(ctx=ctx, package="get", method="nodes", args=[project_id]),
        node_id,
    )
    if not original_node:
        click.secho("Error while getting the data of the orignal ACE.")
        ctx.exit(1)

    args = [
        ("compute_id", compute_id),
        ("name", name),
        ("node_type", node_type),
        ("node_id", new_node_id),
        ("console", console_port),
        ("console_type", console_type),
        ("console_auto_start", console_auto_start),
        ("aux", aux),
        ("aux_type", aux_type),
        ("properties", properties),
        ("symbol", symbol),
        ("x", x),
        ("y", y),
        ("z", z),
        ("locked", locked),
        ("port_name_format", port_name_format),
        ("port_segment_size", port_segment_size),
        ("first_port_name", first_port_name),
    ]

    for arg in args:
        if arg[1]:
            setattr(original_node, arg[0], arg[1])

    if not use_json:
        try:
            if (
                label_text is None
                and label_rotation
                or label_x_position
                or label_y_position
                or label_style_attribute is not None
            ):
                raise click.UsageError(
                    "When adding a label the text has to be set using the -lt option."
                )
            elif label_text is not None:
                original_node.label.text = label_text
                original_node.label.style = label_style_attribute
                original_node.label.x = label_x_position
                original_node.label.y = label_y_position
                original_node.label.rotation = label_rotation

            data = NodeUpdate(
                compute_id=original_node.compute_id,
                name=original_node.name,
                node_type=original_node.node_type,
                console=original_node.console,
                console_type=original_node.console_type,
                console_auto_start=original_node.console_auto_start,
                aux=original_node.aux,
                aux_type=original_node.aux_type,
                properties=original_node.properties,
                label=original_node.label,
                symbol=original_node.symbol,
                x=original_node.x,
                y=original_node.y,
                z=original_node.z,
                locked=original_node.locked,
                port_name_format=original_node.port_name_format,
                port_segment_size=original_node.port_segment_size,
                first_port_name=original_node.first_port_name,
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
    execute_and_print(ctx, client, lambda c: c.update_node(project_id, node_id, data))


@update.command(
    help="Update a new disk for a node in a project.",
    epilog="Example: gns3util -s [server] create qemu-disk-image -f qcow2 -s 1024 PROJECT_ID/Name NODE_ID/Name DISK_NAME",
    name="qemu-disk-image",
)
@click.option(
    "-f",
    "--format",
    type=click.Choice(["qcow2", "qcow", "vpc", "vdi", "vdmk", "raw"]),
    default=None,
    help="Type of image format.",
)
@click.option(
    "-s",
    "--size",
    type=int,
    default=None,
    help="Size of disk in megabytes",
)
@click.option(
    "-p",
    "--preallocation",
    type=click.Choice(["off", "metadata", "falloc", "full"]),
    default=None,
    help="Desired Qemu disk image pre-allocation option.",
)
@click.option(
    "-c",
    "--cluster-size",
    type=int,
    default=None,
    help="Desired cluster size.",
)
@click.option(
    "-r",
    "--refcount-bits",
    type=int,
    default=None,
    help="Desired amount of refcount bits",
)
@click.option(
    "-l",
    "--lazy_refcounts",
    type=click.Choice(["on", "off"]),
    default=None,
    help="Enableling or disabeling lazy refcounts.",
)
@click.option(
    "-sf",
    "--subformat",
    type=click.Choice(
        [
            "dynamic",
            "fixed",
            "streamOptimized",
            "twoGbMaxExtentSparse",
            "twoGbMaxExtentFlat",
            "monolithicSparse",
            "monolithicFlat",
        ]
    ),
    default=None,
    help="Desired image sub-format options.",
)
@click.option(
    "-st",
    "--static",
    type=click.Choice(["on", "off"]),
    default=None,
    help="",
)
@click.option(
    "-zg",
    "--zeroed-grain",
    type=click.Choice(["on", "off"]),
    default=None,
    help="",
)
@click.option(
    "-at",
    "--adapter-type",
    type=click.Choice(["idle", "lsilogic", "buslogic", "legacyESX"]),
    default=None,
    help="",
)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.argument("project-id", required=True, type=str)
@click.argument("node-id", required=True, type=str)
@click.argument("disk-name", required=True, type=str)
@click.pass_context
def qemu_disk_image_create(
    ctx: click.Context,
    format,
    size,
    preallocation,
    cluster_size,
    refcount_bits,
    lazy_refcounts,
    subformat,
    static,
    zeroed_grain,
    adapter_type,
    project_id,
    node_id,
    disk_name,
    use_json,
):
    if (
        not format
        and not size
        and not preallocation
        and not cluster_size
        and not refcount_bits
        and not lazy_refcounts
        and not subformat
        and not static
        and not zeroed_grain
        and not adapter_type
        and not use_json
    ):
        raise click.UsageError(
            "For this command and of the options are required or the -j option on it's own."
        )

    if project_id and not is_valid_uuid(project_id):
        project_id, ok = resolve_ids(ctx, "project", project_id)
        if not ok:
            click.secho(f"{project_id}", err=True)
            ctx.exit(1)

    if node_id and not is_valid_uuid(node_id):
        node_id, ok = resolve_ids(ctx, "node", node_id, args=[project_id])
        if not ok:
            click.secho(f"{node_id}", err=True)
            ctx.exit(1)

    args = [
        ("format", format),
        ("size", size),
        ("preallocation", preallocation),
        ("cluster_size", cluster_size),
        ("refcount_bits", refcount_bits),
        ("lazy_refcounts", lazy_refcounts),
        ("subformat", subformat),
        ("static", static),
        ("zeroed_grain", zeroed_grain),
        ("adapter_type", adapter_type),
    ]

    if not use_json:
        data = {}
        for key, value in args:
            if value is not None:
                data[key] = value
    else:
        try:
            data = json.loads(use_json)
        except json.JSONDecodeError:
            click.secho("Error: Invalid JSON", fg="red", bold=True, err=True)
            ctx.exit(1)

    client = get_client(ctx)
    execute_and_print(
        ctx, client, lambda c: c.update_disk_image(project_id, node_id, disk_name, data)
    )


@update.command(
    epilog="Example: gns3util -s [server] update link PROJECT_ID/Name LINK_ID JSON_DATA",
)
@click.argument("project-id", required=True, type=str)
@click.argument("link-id", required=True, type=str)
@click.argument("json-data", required=True, type=str)
@click.pass_context
def link(
    ctx: click.Context,
    project_id,
    link_id,
    json_data,
):
    """
    Update a link between 2 nodes in a project.

    \b
    Required data schema: https://apiv3.gns3.net/#/Links/update_link_v3_projects__project_id__links__link_id__put
    """

    if project_id and not is_valid_uuid(project_id):
        project_id, ok = resolve_ids(ctx, "project", project_id)
        if not ok:
            click.secho(f"{project_id}", err=True)
            ctx.exit(1)

    try:
        data = json.loads(json_data)
    except json.JSONDecodeError:
        click.secho("Error: Invalid JSON", fg="red", bold=True, err=True)
        ctx.exit(1)

    client = get_client(ctx)
    execute_and_print(ctx, client, lambda c: c.update_link(project_id, link_id, data))


@update.command(
    help="Update a drawing in a project.",
    epilog='Example: gns3util -s [server] update drawing -s "some svg" PROJECT_ID/Name DRAWING_ID',
)
@click.option(
    "-x",
    "--x",
    type=int,
    default=None,
    help="X-Position of the drawing.",
)
@click.option(
    "-y",
    "--y",
    type=int,
    default=None,
    help="Y-Position of the drawing.",
)
@click.option(
    "-z",
    "--z",
    type=int,
    default=1,
    help="Z-Position (layer) of the drawing. Default: 1",
)
@click.option(
    "-l",
    "--locked",
    is_flag=True,
    default=False,
    help="Lock the drawing.",
)
@click.option(
    "-r",
    "--rotation",
    type=click.IntRange(min=-359, max=360, clamp=False),
    default=0,
    help="Rotation of the drawing.",
)
@click.option(
    "-s",
    "--svg",
    type=str,
    default=None,
    help="Raw SVG data for the drawing.",
)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.argument("project-id", required=True, type=str)
@click.argument("drawing-id", required=True, type=str)
@click.pass_context
def drawing(
    ctx: click.Context,
    x,
    y,
    z,
    locked,
    rotation,
    svg,
    project_id,
    drawing_id,
    use_json,
):
    if project_id and not is_valid_uuid(project_id):
        project_id, ok = resolve_ids(ctx, "project", project_id)
        if not ok:
            click.secho(f"{project_id}", err=True)
            ctx.exit(1)

    args = [
        ("x", x),
        ("y", y),
        ("z", z),
        ("locked", locked),
        ("rotation", rotation),
        ("svg", svg),
    ]

    if not use_json:
        data = {}
        for key, value in args:
            if value is not None:
                data[key] = value
    else:
        try:
            data = json.loads(use_json)
        except json.JSONDecodeError:
            click.secho("Error: Invalid JSON", fg="red", bold=True, err=True)
            ctx.exit(1)

    client = get_client(ctx)
    execute_and_print(
        ctx, client, lambda c: c.update_drawing(project_id, drawing_id, data)
    )


@update.command(
    help="Update a compute.",
    epilog="Example: gns3util -s [server] update compute -n some_name",
)
@click.option(
    "-pr",
    "--protocol",
    type=click.Choice(["http", "https"]),
    default=None,
    help="Protcol for connection to the compute.",
)
@click.option(
    "-h",
    "--host",
    type=str,
    default=None,
    help="IP or Domain of the remote host.",
)
@click.option(
    "-p",
    "--port",
    type=click.IntRange(min=0, max=65535, clamp=False),
    default=None,
    help="TCP port to connect with the remote host.",
)
@click.option(
    "-u",
    "--user",
    type=str,
    default=None,
    envvar="GNS3_USER",
    help="Username to connect as. (env: GNS3_USER)",
)
@click.option(
    "-pw",
    "--password",
    type=str,
    default=None,
    envvar="GNS3_PASSWORD",
    help="Password for the user. (env: GNS3_PASSWORD)",
)
@click.option(
    "-n",
    "--name",
    type=str,
    default=None,
    help="Desired name for the compute.",
)
@click.option(
    "-i",
    "--compute-id",
    type=str,
    default=None,
    help="Desired id for the compute.",
)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def compute(
    ctx: click.Context,
    protocol,
    host,
    port,
    user,
    password,
    name,
    compute_id,
    use_json,
):
    if (
        not protocol
        and not host
        and not port
        and not user
        and not password
        and not name
        and not compute_id
        and not use_json
    ):
        raise click.UsageError(
            "For this command either and of the options are required or the -j option on it's own."
        )

    if compute_id and not is_valid_uuid(compute_id):
        compute_id, ok = resolve_ids(ctx, "compute", compute_id)
        if not ok:
            click.secho(f"{compute_id}", err=True)
            ctx.exit(1)

    args = [
        ("protocol", protocol),
        ("host", host),
        ("port", port),
        ("name", name),
        ("compute_id", compute_id),
    ]

    if not use_json:
        data = {}
        for key, value in args:
            if value is not None:
                data[key] = value
    else:
        try:
            data = json.loads(use_json)
        except json.JSONDecodeError:
            click.secho("Error: Invalid JSON", fg="red", bold=True, err=True)
            ctx.exit(1)

    client = get_client(ctx)
    execute_and_print(ctx, client, lambda c: c.update_compute(compute_id, data))


@update.command(
    help="Update a resource pool.",
    epilog="Example: gns3util -s [server] update pool -n some-name POOL_ID/Name",
)
@click.option(
    "-n",
    "--name",
    type=str,
    default=None,
    help="Desired name for the pool",
)
@click.argument("pool-id", required=True, type=str)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def pool(ctx: click.Context, name, pool_id, use_json):
    if not name and not use_json:
        raise click.UsageError(
            "For this command either the -n option is required or the -j option on it's own."
        )

    if pool_id and not is_valid_uuid(pool_id):
        pool_id, ok = resolve_ids(ctx, "pool", pool_id)
        if not ok:
            click.secho(f"{pool_id}", err=True)
            ctx.exit(1)

    if not use_json:
        try:
            data = ResourcePoolUpdate(
                name=name,
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
    execute_and_print(ctx, client, lambda c: c.update_pool(pool_id, data))
