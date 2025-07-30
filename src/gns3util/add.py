import click
import json
from . import auth
from .api.put_endpoints import GNS3PutAPI
from .api.post_endpoints import GNS3PostAPI
from gns3util.schemas import ApplianceVersionImages, ApplianceVersion
from pydantic import ValidationError
from .utils import (
    execute_and_print,
    resolve_ids,
    is_valid_uuid,
)


@click.group()
def add():
    """Add commands"""
    pass


def get_client(ctx: click.Context):
    """Helper function to create GNS3PutAPI instance."""
    server_url = ctx.parent.obj["server"]
    verify = ctx.parent.obj["verify"]
    success, key = auth.load_and_try_key(ctx)
    if success and key:
        return GNS3PutAPI(server_url, key.access_token, verify=verify)
    else:
        ctx.exit(1)


def get_post_client(ctx: click.Context):
    """Helper function to create GNS3PostAPI instance."""
    server_url = ctx.parent.obj["server"]
    verify = ctx.parent.obj["verify"]
    success, key = auth.load_and_try_key(ctx)
    if success and key:
        return GNS3PostAPI(server_url, key.access_token, verify=verify)
    else:
        ctx.exit(1)


@add.command(
    help="Add member to a user group.",
    epilog="Example: gns3util -s [server] add group-member GROUP_ID/Name USER_ID/Name",
    name="group-member",
)
@click.argument("group-id", required=True, type=str)
@click.argument("user-id", required=True, type=str)
@click.pass_context
def privilege(ctx: click.Context, group_id, user_id):
    if not is_valid_uuid(group_id):
        group_id, ok = resolve_ids(ctx, "group", group_id)
        if not ok:
            click.secho(f"{group_id}", err=True)
            ctx.exit(1)

    if not is_valid_uuid(user_id):
        user_id, ok = resolve_ids(ctx, "user", user_id)
        if not ok:
            click.secho(f"{user_id}", err=True)
            ctx.exit(1)

    client = get_client(ctx)
    execute_and_print(ctx, client, lambda c: c.add_group_member(group_id, user_id))


@add.command(
    help="Add resource to a resource pool. For now only Projects are supported by GNS3v3",
    epilog="Example: gns3util -s [server] add to-pool POOL_ID/Name PROJECT_ID",
    name="to-pool",
)
@click.argument("pool-id", required=True, type=str)
@click.argument("project-id", required=True, type=str)
@click.pass_context
def add_resouce_to_pool(ctx: click.Context, pool_id, project_id):
    if not is_valid_uuid(pool_id):
        pool_id, ok = resolve_ids(ctx, "pool", pool_id)
        if not ok:
            click.secho(f"{pool_id}", err=True)
            ctx.exit(1)

    if not is_valid_uuid(project_id):
        project_id, ok = resolve_ids(ctx, "project", project_id)
        if not ok:
            click.secho(f"{project_id}", err=True)
            ctx.exit(1)
    client = get_client(ctx)
    execute_and_print(
        ctx, client, lambda c: c.add_resource_to_pool(pool_id, project_id)
    )


@add.command(
    help="Add a priviledge to a role.",
    epilog="Example: gns3util -s [server] add privilege ROLE_ID/Name privilege_ID/Name",
)
@click.argument("role-id", required=True, type=str)
@click.argument("privilege-id", required=True, type=str)
@click.pass_context
def privilege(ctx: click.Context, role_id, privilege_id):
    if not is_valid_uuid(role_id):
        role_id, ok = resolve_ids(ctx, "role", role_id)
        if not ok:
            click.secho(f"{role_id}", err=True)
            ctx.exit(1)

    if not is_valid_uuid(privilege_id):
        privilege_id, ok = resolve_ids(ctx, "privilege", privilege_id)
        if not ok:
            click.secho(f"{privilege_id}", err=True)
            ctx.exit(1)

    client = get_client(ctx)
    execute_and_print(ctx, client, lambda c: c.update_role_privs(role_id, privilege_id))


@add.command(
    help="Add a version to an appliance.",
    epilog="Example: gns3util -s [server] add appliance-version -n some_name APPLIANCE_ID/Name",
    name="appliance-version",
)
@click.option(
    "-n",
    "--name",
    type=str,
    default=None,
    help="",
)
@click.option(
    "-id",
    "--idlepc",
    type=str,
    default=None,
    help="Idle-PC for the version in hex eg: 0x60630d08, only needed for Network Devices.",
)
@click.option(
    "-ki",
    "--kernel-image",
    type=str,
    default=None,
    help="",
)
@click.option(
    "-in",
    "--initrd",
    type=str,
    default=None,
    help="",
)
@click.option(
    "-i",
    "--image",
    type=str,
    default=None,
    help="",
)
@click.option(
    "-bi",
    "--bios-image",
    type=str,
    default=None,
    help="",
)
@click.option(
    "-hda",
    "--hda-disk-image",
    type=str,
    default=None,
    help="",
)
@click.option(
    "-hdb",
    "--hdb-disk-image",
    type=str,
    default=None,
    help="",
)
@click.option(
    "-hdc",
    "--hdc-disk-image",
    type=str,
    default=None,
    help="",
)
@click.option(
    "-hdd",
    "--hdd-disk-image",
    type=str,
    default=None,
    help="",
)
@click.option(
    "-cd",
    "--cdrom-image",
    type=str,
    default=None,
    help="",
)
@click.argument("appliance-id", required=True, type=str)
@click.option(
    "-j",
    "--use-json",
    type=str,
    default=None,
    help="Provide a string of JSON directly to send.",
)
@click.pass_context
def appliance_version(
    ctx: click.Context,
    name,
    idlepc,
    kernel_image,
    initrd,
    image,
    bios_image,
    hda_disk_image,
    hdb_disk_image,
    hdc_disk_image,
    hdd_disk_image,
    cdrom_image,
    appliance_id,
    use_json,
):
    if not name and not use_json:
        raise click.UsageError(
            "For this command either the -n option is required or the -j option on it's own."
        )

    if appliance_id and not is_valid_uuid(appliance_id):
        appliance_id, ok = resolve_ids(ctx, "appliance", appliance_id)
        if not ok:
            click.secho(f"{appliance_id}", err=True)
            ctx.exit(1)

    if not use_json:
        images = None
        if (
            kernel_image
            or initrd
            or image
            or bios_image
            or hda_disk_image
            or hdb_disk_image
            or hdc_disk_image
            or hdd_disk_image
            or cdrom_image is not None
        ):
            images = ApplianceVersionImages(
                kernel_image=kernel_image,
                initrd=initrd,
                bios_image=bios_image,
                hda_disk_image=hda_disk_image,
                hdb_disk_image=hdb_disk_image,
                hdc_disk_image=hdc_disk_image,
                hdd_disk_image=hdd_disk_image,
                cdrom_image=cdrom_image,
            )
        try:
            data = ApplianceVersion(
                name=name,
                idlepc=idlepc,
                images=images,
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

    client = get_post_client(ctx)
    execute_and_print(
        ctx, client, lambda c: c.create_appliance_version(appliance_id, data)
    )
