import json
import click
import os
import importlib.resources
from . import auth
from .api.delete_endpoints import GNS3DeleteAPI
from .utils import execute_and_print, get_command_description, fuzzy_delete_class_params, fuzzy_delete_class_wrapper, fuzzy_delete_exercise_params, fuzzy_delete_exercise_wrapper

"""
Number of arguments: 0
Has data: False
"""
_zero_arg_no_data = {
    "prune_images": "prune_images"
}

"""
Number of arguments: 1
Has data: False
"""
_one_arg_no_data = {
    "user": "delete_user",
    "compute": "delete_compute",
    "project": "delete_project",
    "template": "delete_template",
    "image": "delete_image",
    "ace": "delete_ace",
    "role": "delete_role",
    "group": "delete_group",
    "pool": "delete_pool"
}

"""
Number of arguments: 2
Has data: False
"""
_two_arg_no_data = {
    "pool_resource": "delete_pool_resource",
    "link": "delete_link",
    "node": "delete_node",
    "drawing": "delete_drawing",
    "role_priv": "delete_role_priv",
    "user_from_group": "delete_user_from_group",
    "snapshot": "delete_snapshot"
}


@click.group()
def delete():
    """Delete commands."""
    pass


def get_client(ctx):
    """Helper function to create GNS3DeleteAPI instance."""
    server_url = ctx.parent.obj['server']
    verify = ctx.parent.obj['verify']
    success, key = auth.load_and_try_key(ctx)
    if success:
        return GNS3DeleteAPI(server_url, key['access_token'], verify=verify)
    else:
        os._exit(1)


with importlib.resources.files("gns3util.help_texts").joinpath("help_delete.json").open("r", encoding="utf-8") as f:
    help_dict = json.load(f)

# Create click commands with zero arguments and no data
for cmd, func in _zero_arg_no_data.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "zero_arg")

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.pass_context
        def cmd_func(ctx):
            api_delete_client = get_client(ctx)
            execute_and_print(
                ctx, api_delete_client, lambda client: getattr(api_delete_client, func)())
        return cmd_func
    delete.command(name=cmd, help=current_help_option,
                   epilog=epiloge)(make_cmd())

# Create click commands with one argument minus JSON
for cmd, func in _one_arg_no_data.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "one_arg")

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.argument('arg')
        @click.pass_context
        def cmd_func(ctx, arg):
            api_delete_client = get_client(ctx)
            execute_and_print(ctx, api_delete_client, lambda client: getattr(
                api_delete_client, func)(arg))
        return cmd_func
    delete.command(name=cmd, help=current_help_option,
                   epilog=epiloge)(make_cmd())

# Create click commands with two arguments minus JSON
for cmd, func in _two_arg_no_data.items():
    current_help_option, epiloge = get_command_description(
        cmd, help_dict, "two_arg")

    def make_cmd(func=func, help_option=current_help_option, epilog=epiloge):
        @click.argument('arg1')
        @click.argument('arg2')
        @click.pass_context
        def cmd_func(ctx, arg1, arg2):
            api_delete_client = get_client(ctx)
            execute_and_print(ctx, api_delete_client, lambda client: getattr(
                api_delete_client, func)(arg1, arg2))
        return cmd_func
    delete.command(name=cmd, help=current_help_option,
                   epilog=epiloge)(make_cmd())


@delete.command(name="class", help="Delete a class and it's students.")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.option(
    "-y", "--confirm", default=True, is_flag=True, help="Require confirmation for deletion"
)
@click.option(
    "-n", "--non_interactive", type=str, help="Run the command non interactively"
)
@click.option(
    "-a", "--delete_all", is_flag=True, help="Delete all classes"
)
@click.option(
    "-e", "--delete_exercises", is_flag=True, help="Delete all exercies of that class"
)
@click.pass_context
def fuzzy_delete_class(ctx, multi, confirm, non_interactive, delete_all, delete_exercises):
    params = fuzzy_delete_class_params(
        ctx=ctx,
        client=get_client,
        method="groups",
        key="name",
        multi=multi,
        confirm=confirm,
        non_interactive=non_interactive,
        delete_all=delete_all,
        delete_exercises=delete_exercises
    )
    fuzzy_delete_class_wrapper(params)


@delete.command(name="exercise", help="Delete an exercise")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.option(
    "-y", "--confirm", default=True, is_flag=True, help="Require confirmation for deletion"
)
@click.option(
    "-n", "--non_interactive", type=str, help="Run the command non interactively"
)
@click.option(
    "-c", "--set_class", help="Set the class from which to delete the exercise"
)
@click.option(
    "-g", "--set_group", help="Set the group from which to delete the exercise"
)
@click.option(
    "-fc", "--select_class", is_flag=True, help="Set the class from which to delete the exercise"
)
@click.option(
    "-fg", "--select_group", is_flag=True, help="Set the group from which to delete the exercise"
)
@click.option(
    "-a", "--delete_all", is_flag=True, help="Delete all exercises"
)
@click.pass_context
def fuzzy_delete_exercise(
    ctx, multi, confirm, non_interactive, set_class, set_group, delete_all, select_class, select_group
):
    if non_interactive:
        if set_class is None and set_group is not None:
            raise click.UsageError(
                "In non-interactive mode, --set_class and --set_group must both provided with string values."
            )
    if not non_interactive:
        if select_class is False and select_group is True:
            raise click.UsageError(
                "In interactive mode, --select_class and --select_group must both be set."
            )
        if (select_class or select_group == True) and multi == True:
            raise click.UsageError(
                "In interactive mode when either, --select_class or --select_group are set multi mode is not supported."
            )

    class_to_use = set_class
    group_to_use = set_group

    params = fuzzy_delete_exercise_params(
        ctx=ctx,
        client=get_client,
        method="projects",
        key="name",
        multi=multi,
        confirm=confirm,
        non_interactive=non_interactive,
        class_to_use=class_to_use,
        group_to_use=group_to_use,
        select_class=select_class,
        select_group=select_group,
        delete_all=delete_all,
        unattended=False
    )
    fuzzy_delete_exercise_wrapper(params)
