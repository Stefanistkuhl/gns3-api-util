import click
import json
import uuid
import importlib.resources
from . import auth
from .api.put_endpoints import GNS3PutAPI
from .utils import execute_and_print, get_command_description, validate_mutually_exclusive_with_json, validate_use_json_mutually_exclusive, is_valid_uuid
from .scripts import resolve_ids
from gns3util.schemas import (
    IOULicense,
    LoggedInUserUpdate,
    UserUpdate,
    UserGroupUpdate,
    RoleUpdate,
    ACEUpdate,
)
import json
from pydantic import ValidationError

"""
Number of arguments: 0
Has data: True
"""

"""
Number of arguments: 1
Has data: True
"""
_one_arg = {
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


def get_client(ctx: click.Context):
    """Helper function to create GNS3PutAPI instance."""
    server_url = ctx.parent.obj['server']
    verify = ctx.parent.obj['verify']
    success, key = auth.load_and_try_key(ctx)
    if success and key:
        return GNS3PutAPI(server_url, key.access_token, verify)
    else:
        ctx.exit(1)


with importlib.resources.files("gns3util.help_texts").joinpath("help_put.json").open("r", encoding="utf-8") as f:
    help_dict = json.load(f)

# Create click commands with one argument plus JSON
for cmd, func in _one_arg.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "one_arg")

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.argument('arg')
        @click.argument('json_data')
        @click.pass_context
        def cmd_func(ctx: click.Context, arg, json_data):
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
        def cmd_func(ctx: click.Context, arg1, arg2):
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
        def cmd_func(ctx: click.Context, arg1, arg2, json_data):
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
        def cmd_func(ctx: click.Context, arg1, arg2, arg3, json_data):
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


@update.command(
    help="Update the IOULicense",
    epilog='Example: gns3util -s [server] update iou_license --iourc_content "some str"',
)
@click.option("-ic", "--iourc_content",
              type=str,
              default=None,
              callback=validate_mutually_exclusive_with_json,
              help="contents of the license"
              )
@click.option("-lc", "--license_check",
              is_flag=True,
              callback=validate_mutually_exclusive_with_json,
              default=False,
              help="enable license checking")
@click.option("-j", "--use-json",
              type=str,
              default=None,
              callback=validate_use_json_mutually_exclusive,
              help="Provide a string of JSON directly to send.")
@click.pass_context
def iou_license(ctx: click.Context, iourc_content, license_check, use_json):
    if not iou_license and not license_check and not use_json:
        raise click.UsageError(
            "For this command the -ic and -lc options are required or the -j on it's own.")
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
    execute_and_print(
        ctx, client, lambda c: c.iou_license(data))


@update.command(
    help="Update the logged in user",
    epilog='Example: gns3util -s [server] update me -p password -e something@exmaple.com -f someName',
)
@click.option("-p", "--password",
              type=str,
              default=None,
              callback=validate_mutually_exclusive_with_json,
              help="Password to set for the current User"
              )
@click.option("-e", "--email",
              type=str,
              callback=validate_mutually_exclusive_with_json,
              default=None,
              help="Email to set for the current User")
@click.option("-f", "--full-name",
              type=str,
              callback=validate_mutually_exclusive_with_json,
              default=None,
              help="Full name to set for the current User")
@click.option("-j", "--use-json",
              type=str,
              default=None,
              callback=validate_use_json_mutually_exclusive,
              help="Provide a string of JSON directly to send.")
@click.pass_context
def me(ctx: click.Context, password, email, full_name, use_json):
    if not password and not email and not full_name and not use_json:
        raise click.UsageError(
            "For this command the any of the -p, -e and -f options are required or the -j option on it's own.")
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
    execute_and_print(
        ctx, client, lambda c: c.me(data))


@update.command(
    help="Update a  given User with a given ID or name which will be resolved to a ID if a User with a matching name exists.",
    epilog='Example: gns3util -s [server] create user -u alice -p password [user-id]',
)
@click.argument("user-id", required=True, type=str)
@click.option("-u", "--username",
              type=str,
              default=None,
              callback=validate_mutually_exclusive_with_json,
              help="Desired username for the User"
              )
@click.option("-i", "--is-active",
              is_flag=True,
              default=False,
              callback=validate_mutually_exclusive_with_json,
              help="Marking the user as currently active")
@click.option("-e", "--email",
              type=str,
              callback=validate_mutually_exclusive_with_json,
              default=None,
              help="Desired email for the user")
@click.option("-f", "--full-name",
              type=str,
              callback=validate_mutually_exclusive_with_json,
              default=None,
              help="Desired full name for the user")
@click.option("-p", "--password",
              type=str,
              callback=validate_mutually_exclusive_with_json,
              default=None,
              help="Desired password for the user")
@click.option("-j", "--use-json",
              type=str,
              default=None,
              callback=validate_use_json_mutually_exclusive,
              help="Provide a string of JSON directly to send.")
@click.pass_context
def user(ctx: click.Context, user_id, username, is_active, email, full_name, password, use_json):
    if not username and not is_active and not email and not full_name and not password and not use_json:
        raise click.UsageError(
            "For this command any of the options required or the -j option on it's own.")

    if not is_valid_uuid(user_id):
        user_id, ok = resolve_ids(ctx, "user", user_id)
        if not ok:
            click.secho(f"{user_id}", err=True)
            ctx.exit(1)

    if not use_json:
        try:
            data = UserUpdate(
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
    execute_and_print(
        ctx, client, lambda c: c.update_user(user_id, data))


@update.command(
    help="Update a group",
    epilog='Example: gns3util -s [server] update group -n some-name [group-id]',
)
@click.argument("group-id", required=True, type=str)
@click.option("-n", "--name",
              type=str,
              default=None,
              callback=validate_mutually_exclusive_with_json,
              help="Desired name for the group"
              )
@click.option("-j", "--use-json",
              type=str,
              default=None,
              callback=validate_use_json_mutually_exclusive,
              help="Provide a string of JSON directly to send.")
@click.pass_context
def group(ctx: click.Context, role_id, group_id, name, use_json):
    if not name and not use_json:
        raise click.UsageError(
            "For this command either the -n option is required or the -j option on it's own.")

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
    execute_and_print(
        ctx, client, lambda c: c.update_group(group_id, data))


@update.command(
    help="Update a role",
    epilog='Example: gns3util -s [server] update role -n some-name [role-id]',
)
@click.argument("role-id", required=True, type=str)
@click.option("-n", "--name",
              type=str,
              default=None,
              callback=validate_mutually_exclusive_with_json,
              help="Desired name for the role."
              )
@click.option("-d", "--description",
              type=str,
              default=None,
              callback=validate_mutually_exclusive_with_json,
              help="Desired description for the role."
              )
@click.option("-j", "--use-json",
              type=str,
              default=None,
              callback=validate_use_json_mutually_exclusive,
              help="Provide a string of JSON directly to send.")
@click.pass_context
def role(ctx: click.Context, role_id, name, description, use_json):
    if not name and not description and not use_json:
        raise click.UsageError(
            "For this command either any option is required or the -j option on it's own.")

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
    execute_and_print(
        ctx, client, lambda c: c.update_role(role_id, data))


@update.command(
    help="Update an ACE. user, group and role id will try to resolve the input to a valid id on the server if no valid UUIDv4 is given so names can be used instead of ids.",
    epilog='Example: gns3util -s [server] create ace -at user -p /pools/[id] -r',
)
@click.argument("ace-id", required=True, type=str)
@click.option("-at", "--ace-type",
              type=click.Choice(["user", "group"]),
              default=None,
              callback=validate_mutually_exclusive_with_json,
              help="Desired type for the ACE."
              )
@click.option("-p", "--path",
              type=str,
              default=None,
              callback=validate_mutually_exclusive_with_json,
              help="Desired path for the ace to affect."
              )
@click.option("-pr", "--propagate",
              is_flag=True,
              default=True,
              callback=validate_mutually_exclusive_with_json,
              help="Apply ACE rules to all nested endpoints in the path. Default: True"
              )
@click.option("-a", "--allow",
              is_flag=True,
              default=True,
              callback=validate_mutually_exclusive_with_json,
              help="Wheater to allow or deny acces to the set path. Default: True"
              )
@click.option("-u", "--user-id",
              type=str,
              default=None,
              callback=validate_mutually_exclusive_with_json,
              help="Desired user id to use for this ACE."
              )
@click.option("-g", "--group-id",
              type=str,
              default=None,
              callback=validate_mutually_exclusive_with_json,
              help="Desired group id to use for this ACE."
              )
@click.option("-r", "--role-id",
              type=str,
              default=None,
              callback=validate_mutually_exclusive_with_json,
              help="Desired role id to use for this ACE."
              )
@click.option("-j", "--use-json",
              type=str,
              default=None,
              callback=validate_use_json_mutually_exclusive,
              help="Provide a string of JSON directly to send.")
@click.pass_context
def ace(ctx: click.Context, ace_id, ace_type, path, propagate, allow, user_id, group_id, role_id, use_json):
    if not ace_type and not path and not user_id and not group_id and not role_id and not use_json:
        raise click.UsageError(
            "For this command any of the options is required or the -j option on it's own.")

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

    if not is_valid_uuid(role_id):
        role_id, ok = resolve_ids(ctx, "role", role_id)
        if not ok:
            click.secho(f"{role_id}", err=True)
            ctx.exit(1)

    if not use_json:
        try:
            data = ACEUpdate(
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
    execute_and_print(
        ctx, client, lambda c: c.update_acl(ace_id, data))
