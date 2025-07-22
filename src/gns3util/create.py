from gns3util.schemas import (
    UserCreate,
    UserGroupCreate,
    RoleCreate,
    ACECreate,
    QemuDiskImageCreate,
    TemplateCreate,
    ProjectCreate,
    Supplier,
    TemplateUsage,
    NodeCreate,
    Label,
    SnapshotCreate,
    ComputeCreate,
    ResourcePoolCreate,
)
from pydantic import ValidationError
import importlib.resources
from .server import start_and_get_data
from .utils import (
    execute_and_print,
    create_class,
    create_Exercise,
    get_command_description,
    is_valid_uuid,
    resolve_ids,
    close_project,
)
from .api.post_endpoints import GNS3PostAPI
from . import auth
import json
import click
import uuid

_one_arg_no_data = {
    "create": "create_symbol",
}


_three_arg_no_data = {"node_file": "create_node_file"}


@click.group()
def create():
    """Creation commands."""
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


# Replace help_path and open with importlib.resources
with (
    importlib.resources.files("gns3util.help_texts")
    .joinpath("help_post.json")
    .open("r", encoding="utf-8") as f
):
    help_dict = json.load(f)


# Create click commands with one argument minus JSON
for cmd, func in _one_arg_no_data.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "one_arg_no_data"
    )

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.argument("arg")
        @click.pass_context
        def cmd_func(ctx: click.Context, arg):
            api_post_client = get_client(ctx)
            execute_and_print(
                ctx, api_post_client, lambda client: getattr(api_post_client, func)(arg)
            )

        return cmd_func

    create.command(name=cmd, help=current_help_option, epilog=epiloge)(make_cmd())

# Create click commands with two arguments minus JSON
for cmd, func in _three_arg_no_data.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "three_arg_no_data"
    )

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.argument("arg1")
        @click.argument("arg2")
        @click.argument("arg3")
        @click.pass_context
        def cmd_func(ctx: click.Context, arg1, arg2, arg3):
            api_post_client = get_client(ctx)
            execute_and_print(
                ctx,
                api_post_client,
                lambda client: getattr(api_post_client, func)(arg1, arg2, arg3),
            )

        return cmd_func

    create.command(name=cmd, help=current_help_option, epilog=epiloge)(make_cmd())


@create.command(
    name="class", help="create everything need to setup a class and it's students"
)
@click.argument("filename", required=False, type=click.Path(exists=True, readable=True))
@click.option(
    "-c",
    "--create",
    is_flag=True,
    help="Launch a local webpage to enter the info to create a class",
)
@click.pass_context
def make_class(ctx: click.Context, filename, create):
    if filename is None and create is False:
        click.secho("Please either use the -c flag or give a json file as input to use")
        return

    if create:
        data = start_and_get_data(host="localhost", port=8080, debug=True)
        if data:
            class_name, success = create_class(ctx, None, data)
            if success:
                click.secho("Success: ", nl=False, fg="green")
                click.secho("created class ", nl=False)
                click.secho(f"{class_name}", bold=True)
            else:
                click.secho("Error: ", nl=False, fg="red", err=True)
                click.secho("failed to create class", bold=True, err=True)
        else:
            click.secho("no data", err=True)
            return
    else:
        file = click.format_filename(filename)
        class_name, success = create_class(ctx, file)
        if success:
            click.secho("Success: ", nl=False, fg="green")
            click.secho("created class ", nl=False)
            click.secho(f"{class_name}", bold=True)
        else:
            click.secho("Error: ", nl=False, fg="red", err=True)
            click.secho("failed to create class", bold=True, err=True)


@create.command(
    name="exercise",
    help="Create a Project for every group in a class with ACL's to lock down access.",
)
@click.option("-c", "--class_name", type=str)
@click.option("-e", "--exercise_name", type=str)
@click.pass_context
# TODO add tabcomplete for classes to use
def make_exercise(ctx: click.Context, class_name, exercise_name):
    success = create_Exercise(ctx, class_name, exercise_name)
    if success:
        click.secho("Success: ", nl=False, fg="green")
        click.secho("Exercise ", nl=False)
        click.secho(f"{exercise_name} ", bold=True, nl=False)
        click.secho("and it's acls created sucessfully")
    else:
        click.secho("Error: ", nl=False, fg="red", err=True)
        click.secho("failed to create exercise ", nl=False, err=True)
        click.secho(f"{exercise_name}", bold=True, err=True)


@create.command(
    help="Create a User",
    epilog="Example: gns3util -s [server] create user -u alice -p password",
)
@click.option(
    "-u",
    "--username",
    type=str,
    default=None,
    help="Desired username for the User",
)
@click.option(
    "-i",
    "--is-active",
    is_flag=True,
    default=False,
    help="Marking the user as currently active",
)
@click.option(
    "-e",
    "--email",
    type=str,
    default=None,
    help="Desired email for the user",
)
@click.option(
    "-f",
    "--full-name",
    type=str,
    default=None,
    help="Full name to set for the current User",
)
@click.option(
    "-p",
    "--password",
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
def user(ctx: click.Context, username, is_active, email, full_name, password, use_json):
    if not username and not password and not use_json:
        raise click.UsageError(
            "For this command -u and -p options are required or the -j option on it's own."
        )
    if (is_active or email or full_name) and not (username and password):
        raise click.UsageError(
            "For this command -u and -p options are required or the -j option on it's own."
        )
    if not use_json:
        try:
            data = UserCreate(
                username=username,
                password=password,
                email=email,
                is_active=is_active,
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
    execute_and_print(ctx, client, lambda c: c.create_user(data))


@create.command(
    help="Create a group",
    epilog="Example: gns3util -s [server] create group -n some-name",
)
@click.option(
    "-n",
    "--name",
    type=str,
    default=None,
    help="Desired name for the group",
)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def group(ctx: click.Context, name, use_json):
    if not name and not use_json:
        raise click.UsageError(
            "For this command either the -n option is required or the -j option on it's own."
        )
    if not use_json:
        try:
            data = UserGroupCreate(
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
    execute_and_print(ctx, client, lambda c: c.create_group(data))


@create.command(
    help="Creaste a role",
    epilog="Example: gns3util -s [server] create role -n some-name",
)
@click.option(
    "-n",
    "--name",
    type=str,
    default=None,
    help="Desired name for the role.",
)
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
def role(ctx: click.Context, name, description, use_json):
    if not name and not description and not use_json:
        raise click.UsageError(
            "For this command either the -n option is required or the -j option on it's own."
        )
    if description and not name:
        raise click.UsageError("For this command the -n option is required.")

    if not use_json:
        try:
            data = RoleCreate(
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
    execute_and_print(ctx, client, lambda c: c.create_role(data))


@create.command(
    help="Create an ACE. user, group and role id will try to resolve the input to a valid id on the server if no valid UUIDv4 is given so names can be used instead of ids.",
    epilog="Example: gns3util -s [server] create ace -at user -p /pools/[id] -r",
)
@click.option(
    "-at",
    "--ace-type",
    type=click.Choice(["user", "group"]),
    default=None,
    help="Desired type for the ACE.",
)
@click.option(
    "-p",
    "--path",
    type=str,
    default=None,
    help="Desired path for the ace to affect.",
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
            "For this command either the -at, -p, -r options, are required or the -j option on it's own."
        )
    if ace_type or path or role_id and use_json is None:
        raise click.UsageError(
            "For this command either the -at, -p, -r options, are required or the -j option on it's own."
        )
    if (
        ace_type == "user"
        and group_id is not None
        or ace_type == "group"
        and user_id is not None
    ):
        raise click.UsageError(
            "If you select user as ACE type you must specify a group id to use and vice versa."
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

    if not is_valid_uuid(role_id):
        role_id, ok = resolve_ids(ctx, "role", role_id)
        if not ok:
            click.secho(f"{role_id}", err=True)
            ctx.exit(1)

    if not use_json:
        try:
            data = ACECreate(
                ace_type=ace_type,
                path=path,
                propagate=propagate,
                allowed=allow,
                user_id=user_id,
                group_id=group_id,
                role_id=uuid.UUID(role_id),
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
    execute_and_print(ctx, client, lambda c: c.create_acl(data))


@create.command(
    help="Create a QEMU disk image.",
    epilog="Example: gns3util -s [server] create qemu_img -f qcow2 -s 0 path",
    name="qemu-img",
)
@click.argument("image-path", required=True, type=str)
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
@click.pass_context
def qemu_img(
    ctx: click.Context,
    format,
    size,
    preallocation,
    refcount_bits,
    cluster_size,
    lazy_refcounts,
    subformat,
    static,
    zeroed_grain,
    adapter_type,
    image_path,
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
            "For this command the -f and -s options, are required or the -j option on it's own."
        )
    if (format is None or size is None) and use_json is None:
        raise click.UsageError(
            "For this command the -f and -s options, are required or the -j option on it's own."
        )

    if not use_json:
        try:
            data = QemuDiskImageCreate(
                format=format,
                size=size,
                preallocation=preallocation,
                cluster_size=cluster_size,
                refcount_bits=cluster_size,
                lazy_refcounts=lazy_refcounts,
                subformat=subformat,
                static=static,
                zeroed_grain=zeroed_grain,
                adapter_type=adapter_type,
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
    execute_and_print(ctx, client, lambda c: c.create_qemu_image(image_path, data))


@create.command(
    help="Create a Template.",
    epilog="Example: gns3util -s [server] create template -n some_name -t vpcs",
)
@click.option(
    "-id",
    "--template-id",
    type=str,
    default=None,
    help="Desired ID for template, leave empty to use a generated one.",
)
@click.option(
    "-n",
    "--name",
    type=str,
    default=None,
    help="Desired name of the template.",
)
@click.option(
    "-v",
    "--version",
    type=str,
    default=None,
    help="",
)
@click.option(
    "-c",
    "--category",
    type=click.Choice(["router", "switch", "guest", "firewall"]),
    default=None,
    help="",
)
@click.option(
    "-df",
    "--default-name-format",
    type=str,
    default=None,
    help="",
)
@click.option(
    "-s",
    "--symbol",
    type=str,
    default=None,
    help="",
)
@click.option(
    "-t",
    "--template-type",
    type=click.Choice(
        [
            "cloud",
            "nat",
            "ethernet_hub",
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
    help="",
)
@click.option(
    "-ci",
    "--compute-id",
    type=str,
    default=None,
    help="",
)
@click.option(
    "-u",
    "--usage",
    type=str,
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
@click.pass_context
def template(
    ctx: click.Context,
    template_id,
    name,
    version,
    category,
    default_name_format,
    symbol,
    template_type,
    compute_id,
    usage,
    use_json,
):
    if (
        not template_id
        and not name
        and not version
        and not category
        and not default_name_format
        and not symbol
        and not template_type
        and not compute_id
        and not usage
        and not use_json
    ):
        raise click.UsageError(
            "For this command the -n and -t options, are required or the -j option on it's own."
        )
    if (name is None or template_type is None) and use_json is None:
        raise click.UsageError(
            "For this command the -n and -t options, are required or the -j option on it's own."
        )

    if compute_id is not None:
        if not is_valid_uuid(compute_id):
            compute_id, ok = resolve_ids(ctx, "compute", compute_id)
            if not ok:
                click.secho(f"{compute_id}", err=True)
                ctx.exit(1)

    if template_id is None:
        template_id = uuid.uuid4()

    if not use_json:
        try:
            data = TemplateCreate(
                template_id=template_id,
                name=name,
                version=version,
                category=category,
                default_name_format=default_name_format,
                symbol=symbol,
                template_type=template_type,
                compute_id=compute_id,
                usage=usage,
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
    execute_and_print(ctx, client, lambda c: c.create_template(data))


@create.command(
    help="Create a Project.",
    epilog="Example: gns3util -s [server] create project -n some_name",
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
    "-caf",
    "--close-after-creation",
    is_flag=True,
    default=True,
    help="Wheater to close the project after it's creation. Default: True",
)
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
    project_id,
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
    close_after_creation,
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
            "For this command the -n option, is required or the -j option on it's own."
        )
    if name is None and use_json is None:
        raise click.UsageError(
            "For this command the -n option, is required or the -j option on it's own."
        )

    if project_id is None:
        project_id = uuid.uuid4()

    if not use_json:
        try:
            supplier = None
            if supplier_logo is None and supplier_url is not None:
                raise click.UsageError(
                    "When adding a supplier to the project a logo has to be set using the -supl option."
                )
            elif supplier_logo is not None:
                supplier = Supplier(
                    logo=supplier_logo,
                    url=supplier_url,
                )

            data = ProjectCreate(
                name=name,
                project_id=project_id,
                path=path,
                auto_close=auto_close,
                auto_open=auto_open,
                auto_start=auto_start,
                scene_height=scene_height,
                scene_width=scene_width,
                zoom=zoom,
                show_layers=show_layers,
                snap_to_grid=snap_to_grid,
                show_grid=show_grid,
                grid_size=grid_size,
                drawing_grid_size=drawing_grid_size,
                show_interface_labels=show_interface_labels,
                supplier=supplier,
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
    execute_and_print(ctx, client, lambda c: c.create_project(data))
    if close_after_creation:
        close_project(ctx, str(project_id))


@create.command(
    help="Create a new node from a template in a given project.",
    epilog="Example: gns3util -s [server] create node-from-template -x 10 -y 10 PROJECT_ID/Name TEMPLATE_ID/Name",
    name="node-from-template",
)
@click.option(
    "-x",
    "--x",
    type=int,
    default=None,
    help="X-Coordinate to place the node at",
)
@click.option(
    "-y",
    "--y",
    type=int,
    default=None,
    help="Y-Coordinate to place the node at",
)
@click.option(
    "-n",
    "--name",
    type=str,
    default=None,
    help="Desired name of the node.",
)
@click.option(
    "-c",
    "--compute-id",
    type=str,
    default="local",
    help='Compute on that the Node get\'s created. Default: "local"',
)
@click.argument("project-id", required=True, type=str)
@click.argument("template-id", required=True, type=str)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def node_from_template(
    ctx: click.Context, x, y, name, compute_id, project_id, template_id, use_json
):
    if not x and not x and not name and not compute_id and not use_json:
        raise click.UsageError(
            "For this command -x and -y options are required or the -j option on it's own."
        )
    if (name or compute_id) and not (x and y):
        raise click.UsageError(
            "For this command -x and -y options are required or the -j option on it's own."
        )

    if project_id and not is_valid_uuid(project_id):
        project_id, ok = resolve_ids(ctx, "project", project_id)
        if not ok:
            click.secho(f"{project_id}", err=True)
            ctx.exit(1)

    if template_id and not is_valid_uuid(template_id):
        template_id, ok = resolve_ids(ctx, "template", template_id)
        if not ok:
            click.secho(f"{template_id}", err=True)
            ctx.exit(1)

    if not use_json:
        try:
            data = TemplateUsage(
                x=x,
                y=y,
                name=name,
                compute_id=compute_id,
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
    execute_and_print(
        ctx,
        client,
        lambda c: c.create_project_node_from_template(project_id, template_id, data),
    )


@create.command(
    help="Create a Node in a Project. To use custom adapters the --use-json option has to be used.",
    epilog="Example: gns3util -s [server] create node -n some_name -nt ethernet_switch PROJECT_ID/Name",
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
    "--node-id",
    is_flag=True,
    default=True,
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
    node_id,
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
    use_json,
):
    if (
        not compute_id
        and not name
        and not node_type
        and not node_id
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
        and not z
        and not locked
        and not port_name_format
        and not port_segment_size
        and not first_port_name
        and not use_json
    ):
        raise click.UsageError(
            "For this command the -n and -nt options are required or the -j option on it's own."
        )
    if name is None or node_type is None and use_json is None:
        raise click.UsageError(
            "For this command the -n and -nt options are required or the -j option on it's own."
        )

    if project_id and not is_valid_uuid(project_id):
        project_id, ok = resolve_ids(ctx, "project", project_id)
        if not ok:
            click.secho(f"{project_id}", err=True)
            ctx.exit(1)

    if compute_id and not is_valid_uuid(compute_id) and compute_id != "local":
        compute_id, ok = resolve_ids(ctx, "compute", compute_id)
        if not ok:
            click.secho(f"{compute_id}", err=True)
            ctx.exit(1)

    if node_id is None:
        node_id = uuid.uuid4()

    if not use_json:
        try:
            label = None
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
                label = Label(
                    text=label_text,
                    style=label_style_attribute,
                    x=label_x_position,
                    y=label_y_position,
                    rotation=label_rotation,
                )

            data = NodeCreate(
                compute_id=compute_id,
                name=name,
                node_type=node_type,
                console=console_port,
                console_type=console_type,
                console_auto_start=console_auto_start,
                aux=aux,
                aux_type=aux_type,
                properties=properties,
                label=label,
                symbol=symbol,
                x=x,
                y=y,
                z=z,
                locked=locked,
                port_name_format=port_name_format,
                port_segment_size=port_segment_size,
                first_port_name=first_port_name,
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
    execute_and_print(ctx, client, lambda c: c.create_node(project_id, data))


@create.command(
    help="Create a new disk for a node in a project.",
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
    if not format and not size and not use_json:
        raise click.UsageError(
            "For this command -f and -s options are required or the -j option on it's own."
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

    if not use_json:
        try:
            data = QemuDiskImageCreate(
                format=format,
                size=size,
                preallocation=preallocation,
                cluster_size=cluster_size,
                refcount_bits=refcount_bits,
                lazy_refcounts=lazy_refcounts,
                subformat=subformat,
                static=static,
                zeroed_grain=zeroed_grain,
                adapter_type=adapter_type,
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
    execute_and_print(
        ctx, client, lambda c: c.create_disk_image(project_id, node_id, disk_name, data)
    )


@create.command(
    epilog="Example: gns3util -s [server] create link PROJECT_ID/Name JSON_DATA",
)
@click.argument("project-id", required=True, type=str)
@click.argument("json-data", required=True, type=str)
@click.pass_context
def link(
    ctx: click.Context,
    project_id,
    json_data,
):
    """
    Create a link between 2 nodes in a project.

    \b
    Required data schema: https://apiv3.gns3.net/#/Links/create_link_v3_projects__project_id__links_post
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
    execute_and_print(ctx, client, lambda c: c.create_link(project_id, data))


@create.command(
    help="Create a new drawing in a project.",
    epilog='Example: gns3util -s [server] create drawing -s "some svg" PROJECT_ID/Name',
)
@click.option(
    "-x",
    "--x",
    type=int,
    default=0,
    help="X-Position of the drawing.",
)
@click.option(
    "-y",
    "--y",
    type=int,
    default=0,
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

    print(data)
    client = get_client(ctx)
    execute_and_print(ctx, client, lambda c: c.create_drawing(project_id, data))


@create.command(
    help="Create a snapshot of a project.",
    epilog="Example: gns3util -s [server] create snapshot -n some-name PROJECT_ID/Name",
)
@click.option(
    "-n",
    "--name",
    type=str,
    default=None,
    help="Desired name for the snapshot.",
)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.argument("project-id", required=True, type=str)
@click.pass_context
def snapshot(ctx: click.Context, name, project_id, use_json):
    if not name and not use_json:
        raise click.UsageError(
            "For this command either the -n option is required or the -j option on it's own."
        )

    if project_id and not is_valid_uuid(project_id):
        project_id, ok = resolve_ids(ctx, "project", project_id)
        if not ok:
            click.secho(f"{project_id}", err=True)
            ctx.exit(1)

    if not use_json:
        try:
            data = SnapshotCreate(
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
    execute_and_print(ctx, client, lambda c: c.create_snapshot(project_id, data))


@create.command(
    help="Create a compute.",
    epilog="Example: gns3util -s [server] create compute -pr https -h 10.0.0.69 -p 3080 -u admin -p admin",
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
    help="Desired id for the compute. (Generated if empty)",
)
@click.option(
    "-c",
    "--connect",
    is_flag=True,
    default=False,
    help="Attempt connection to compute after creation.",
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
    connect,
    use_json,
):
    if not protocol and host and port and not user and not use_json:
        raise click.UsageError(
            "For this command either the -pr, -h, -u and -u options are required or the -j option on it's own."
        )

    if not compute_id:
        compute_id = uuid.uuid4()

    if not use_json:
        try:
            data = ComputeCreate(
                protocol=protocol,
                host=host,
                port=port,
                user=user,
                password=password,
                name=name,
                compute_id=compute_id,
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
    execute_and_print(ctx, client, lambda c: c.create_compute(connect, data))


@create.command(
    help="Create a new resource pool.",
    epilog="Example: gns3util -s [server] create pool -n some-name",
)
@click.option(
    "-n",
    "--name",
    type=str,
    default=None,
    help="Desired name for the pool",
)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def pool(ctx: click.Context, name, use_json):
    if not name and not use_json:
        raise click.UsageError(
            "For this command either the -n option is required or the -j option on it's own."
        )
    if not use_json:
        try:
            data = ResourcePoolCreate(
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
    execute_and_print(ctx, client, lambda c: c.create_pool(data))
