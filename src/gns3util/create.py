import click
import json
from . import auth
from .api.post_endpoints import GNS3PostAPI
from .utils import execute_and_print, create_class, create_Exercise, get_command_description, validate_mutually_exclusive_with_json, validate_use_json_mutually_exclusive, is_valid_uuid
from .scripts import resolve_ids
from .server import start_and_get_data
import importlib.resources
import json
from pydantic import ValidationError
from gns3util.schemas import (
    UserCreate,
    UserGroupCreate,
    RoleCreate,
    ACECreate,
)

_zero_arg = {
    "group": "create_group",
    "role": "create_role",
    "template": "create_template",
    "project": "create_project",
    "project_load": "load_project",
    "add_pool": "create_pool",
    "create_compute": "create_compute",
}

_one_arg = {
    "qemu_img": "create_qemu_image",
    "node": "create_node",
    "link_create": "create_link",
    "drawing_create": "create_drawing",
    "snapshot_create": "create_snapshot",
    "add_applience_version": "create_appliance_version"
}

_one_arg_no_data = {
    "create": "create_symbol",
}

_two_arg = {
    "project_node_from_template": "create_project_node_from_template",
}

_three_arg = {
    "disk_img": "create_disk_image"
}

_three_arg_no_data = {
    "node_file": "create_node_file"
}


@click.group()
def create():
    """Creation commands."""
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


# Replace help_path and open with importlib.resources
with importlib.resources.files("gns3util.help_texts").joinpath("help_post.json").open("r", encoding="utf-8") as f:
    help_dict = json.load(f)

# Create click commands with zero arguments
for cmd, func in _zero_arg.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "zero_arg_data")

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.argument('json_data')
        @click.pass_context
        def cmd_func(ctx: click.Context, json_data):
            api_post_client = get_client(ctx)
            try:
                data = json.loads(json_data)
                execute_and_print(
                    ctx, api_post_client, lambda client: getattr(api_post_client, func)(data))
            except json.JSONDecodeError:
                click.secho("Error: ", nl=True, fg="red", err=True)
                click.secho("Invalid JSON input", bold=True, err=True)
                return
        return cmd_func
    create.command(name=cmd, help=current_help_option,
                   epilog=epiloge)(make_cmd())


# Create click commands with one argument plus JSON
for cmd, func in _one_arg.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "one_arg_data")

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.argument('arg')
        @click.argument('json_data')
        @click.pass_context
        def cmd_func(ctx: click.Context, arg, json_data):
            api_post_client = get_client(ctx)
            try:
                data = json.loads(json_data)
                execute_and_print(ctx, api_post_client, lambda client: getattr(
                    api_post_client, func)(arg, data))
            except json.JSONDecodeError:
                click.secho("Error: ", nl=True, fg="red", err=True)
                click.secho("Invalid JSON input", bold=True, err=True)
                return
        return cmd_func
    create.command(name=cmd, help=current_help_option,
                   epilog=epiloge)(make_cmd())

# Create click commands with one argument minus JSON
for cmd, func in _one_arg_no_data.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "one_arg_no_data")

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.argument('arg')
        @click.pass_context
        def cmd_func(ctx: click.Context, arg):
            api_post_client = get_client(ctx)
            execute_and_print(ctx, api_post_client, lambda client: getattr(
                api_post_client, func)(arg))
        return cmd_func
    create.command(name=cmd, help=current_help_option,
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
            api_post_client = get_client(ctx)
            try:
                data = json.loads(json_data)
                execute_and_print(ctx, api_post_client, lambda client: getattr(
                    api_post_client, func)(arg1, arg2, data))
            except json.JSONDecodeError:
                click.secho("Error: ", nl=True, fg="red", err=True)
                click.secho("Invalid JSON input", bold=True, err=True)
                return
        return cmd_func
    create.command(name=cmd, help=current_help_option,
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
            api_post_client = get_client(ctx)
            try:
                data = json.loads(json_data)
                execute_and_print(ctx, api_post_client, lambda client: getattr(
                    api_post_client, func)(arg1, arg2, arg3, data))
            except json.JSONDecodeError:
                click.secho("Error: ", nl=True, fg="red", err=True)
                click.secho("Invalid JSON input", bold=True, err=True)
                return
        return cmd_func
    create.command(name=cmd, help=current_help_option,
                   epilog=epiloge)(make_cmd())

# Create click commands with two arguments minus JSON
for cmd, func in _three_arg_no_data.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "three_arg_no_data")

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.argument('arg1')
        @click.argument('arg2')
        @click.argument('arg3')
        @click.pass_context
        def cmd_func(ctx: click.Context, arg1, arg2, arg3):
            api_post_client = get_client(ctx)
            execute_and_print(ctx, api_post_client, lambda client: getattr(
                api_post_client, func)(arg1, arg2, arg3))
        return cmd_func
    create.command(name=cmd, help=current_help_option,
                   epilog=epiloge)(make_cmd())


@create.command(name="class", help="create everything need to setup a class and it's students")
@click.argument('filename', required=False, type=click.Path(exists=True, readable=True))
@click.option(
    "-c", "--create", is_flag=True, help="Launch a local webpage to enter the info to create a class"
)
@click.pass_context
def make_class(ctx: click.Context, filename, create):

    if filename is None and create is False:
        click.secho(
            "Please either use the -c flag or give a json file as input to use")
        return

    if create:
        data = start_and_get_data(host='localhost', port=8080, debug=True)
        if data:
            class_name, success = create_class(ctx, None, data)
            if success:
                click.secho("Success: ", nl=False, fg="green")
                click.secho("created class ", nl=False)
                click.secho(f"{class_name}", bold=True)
            else:
                click.secho("Error: ", nl=False, fg="red", err=True)
                click.secho(
                    "failed to create class", bold=True, err=True)
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
            click.secho(
                "failed to create class", bold=True, err=True)


@create.command(name="exercise", help="Create a Project for every group in a class with ACL's to lock down access.")
@click.option('-c', '--class_name', type=str)
@click.option('-e', '--exercise_name', type=str)
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
    epilog='Example: gns3util -s [server] create user -u alice -p password',
)
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
              help="Full name to set for the current User")
@click.option("-p", "--password",
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
def user(ctx: click.Context, username, is_active, email, full_name, password, use_json):
    if not username and not password and not use_json:
        raise click.UsageError(
            "For this command -u and -p options are required or the -j option on it's own.")
    if (is_active or email or full_name) and not (username and password):
        raise click.UsageError(
            "For this command -u and -p options are required or the -j option on it's own.")
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
    execute_and_print(
        ctx, client, lambda c: c.create_user(data))


@create.command(
    help="Create a group",
    epilog='Example: gns3util -s [server] create group -n some-name',
)
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
def group(ctx: click.Context, name, use_json):
    if not name and not use_json:
        raise click.UsageError(
            "For this command either the -n option is required or the -j option on it's own.")
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
    execute_and_print(
        ctx, client, lambda c: c.create_group(data))


@create.command(
    help="Creaste a role",
    epilog='Example: gns3util -s [server] create role -n some-name',
)
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
def role(ctx: click.Context, name, description, use_json):
    if not name and not description and not use_json:
        raise click.UsageError(
            "For this command either the -n option is required or the -j option on it's own.")
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
    execute_and_print(
        ctx, client, lambda c: c.create_role(data))


@create.command(
    help="Create an ACE. user, group and role id will try to resolve the input to a valid id on the server if no valid UUIDv4 is given so names can be used instead of ids.",
    epilog='Example: gns3util -s [server] create ace -at user -p /pools/[id] -r',
)
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
def ace(ctx: click.Context, ace_type, path, propagate, allow, user_id, group_id, role_id, use_json):
    if not ace_type and not path and not user_id and not group_id and not role_id and not use_json:
        raise click.UsageError(
            "For this command either the -at, -p, -r options, are required or the -j option on it's own.")
    if ace_type or path or role_id and use_json is None:
        raise click.UsageError(
            "For this command either the -at, -p, -r options, are required or the -j option on it's own.")
    if ace_type == "user" and group_id is not None or ace_type == "group" and user_id is not None:
        raise click.UsageError(
            "If you select user as ACE type you must specify a group id to use and vice versa.")

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
            data = ACECreate(
                ace_type=ace_type,
                path=path,
                propagate=propagate,
                allowed=allow,
                user_id=user_id,
                group_id=group_id,
                role_id=role_id,
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
        ctx, client, lambda c: c.create_acl(data))
