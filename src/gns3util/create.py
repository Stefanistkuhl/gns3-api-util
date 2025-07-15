import click
import json
from . import auth
import os
from .api.post_endpoints import GNS3PostAPI
from .utils import execute_and_print, create_class, create_Exercise, get_command_description
from .server import start_and_get_data
import importlib.resources

_zero_arg = {
    "user": "create_user",
    "group": "create_group",
    "role": "create_role",
    "acl": "create_acl",
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
    if success:
        return GNS3PostAPI(server_url, key['access_token'], verify=verify)
    else:
        os._exit(1)


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


@create.command(name="exercise", help="create everything need to setup a class and it's students")
@click.argument('class_name', type=str)
@click.argument('exercise_name', type=str)
@click.pass_context
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
