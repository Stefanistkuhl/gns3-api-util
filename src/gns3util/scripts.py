import click
from .utils import call_client_method, parse_yml, GNS3Error
from typing import Callable, Any, Optional
import itertools

command_group_package = {
    "add": "add",
    "auth": "auth",
    "compute": "post",
    "controller": "post",
    "project": "post",
    "node": "post",
    "image": "post",
    "snapshot": "post",
    "create": "create",
    "update": "update",
    "delete": "delete",
    "get": "get",
    "fuzzy": "fuzzy",
}


@click.command(name="script")
@click.argument('filename', required=False, type=click.Path(exists=True, readable=True))
@click.pass_context
def script(ctx, filename):
    """Run yml based scripts"""
    yml, ok = parse_yml(filename)
    if not ok:
        raise Exception("Invalid yml file: " + str(yml))
    # have func to validate name of things and stuff aswell and raise err if it has wrong names and stuff
    click.secho("Executing script: ", nl=False)
    click.secho(f"{yml["options"][0]["name"]}", bold=True)
    cmds = get_commands(yml)
    for cmd in cmds:
        out, err = run_command(ctx, cmd)
        if GNS3Error.has_error(err):
            GNS3Error.print_error(err)
            return
        print(out)


def get_commands(yml: dict) -> list:
    commands_list = []
    for commands in yml["commands"]:
        cmd = []
        cmd_name = commands["name"]
        for command in commands["subcommands"]:
            sub_cmd_name = command["name"]
            cmd.append(cmd_name)
            cmd.append(sub_cmd_name)
            commands_list.append(cmd)
    return commands_list


def run_command(ctx, input: list) -> tuple[Any, GNS3Error]:
    for pkg in command_group_package:
        if pkg == input[0]:
            run_cmd_err, out = call_client_method(ctx, pkg, input[1])
            if GNS3Error.has_error(run_cmd_err):
                return None, run_cmd_err
            return out, run_cmd_err
