import click
import os
import shutil
import sys
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

GNS3UTIL_SCRIPTS_DIR = os.path.expanduser("~/.gns3util_scripts")


def get_yml_files(script_dir):
    """
    Recursively finds .yml files in script_dir and returns their
    paths relative to script_dir.
    """
    yml_files = []
    if not os.path.exists(script_dir):
        return []

    for root, dirs, files in os.walk(script_dir):
        for file in files:
            if file.endswith(".yml"):
                relative_path = os.path.relpath(
                    os.path.join(root, file), script_dir
                )
                yml_files.append(relative_path)
    return yml_files


def get_script_names_for_completion(ctx, param, incomplete):
    """
    Provides the list of available YML script names for shell completion.
    """
    yml = get_yml_files(GNS3UTIL_SCRIPTS_DIR)
    return [script for script in yml if script.startswith(incomplete)]


"""
TODOS:
script ls
script run (--dry) (shell completion mit avaliable Scripts)
script test (check for errs)
script rm
script add

show commands/steps
loading bars option
"""


@click.group()
def script():
    """script commands."""
    pass


@script.command(name="ls")
@click.pass_context
def list_scripts(ctx):
    """List all avaliable scripts"""
    if not os.path.exists(GNS3UTIL_SCRIPTS_DIR):
        os.mkdir(GNS3UTIL_SCRIPTS_DIR)
        click.secho(f"Created script directory: {GNS3UTIL_SCRIPTS_DIR}")
        return

    files = get_yml_files(GNS3UTIL_SCRIPTS_DIR)
    if files:
        click.secho("Available GNS3 scripts:")
        for f in files:
            click.secho(f"- {f}")
    else:
        click.secho(f"No GNS3 scripts found in {GNS3UTIL_SCRIPTS_DIR}.")


@script.command(name="add")
@click.argument(
    "filename",  type=click.Path(exists=True, readable=True), required=True)
@click.argument(
    "dst",  type=str, required=False)
@click.pass_context
def add_script(ctx, filename, dst: str):
    """Add script to the gns3util scripts directory"""
    if not os.path.exists(GNS3UTIL_SCRIPTS_DIR):
        os.mkdir(GNS3UTIL_SCRIPTS_DIR)
        click.secho(f"Created script directory: {GNS3UTIL_SCRIPTS_DIR}")

    if dst:
        dst_split = dst.split("/")
        dst_path = os.path.join(GNS3UTIL_SCRIPTS_DIR, *dst_split)
        if not os.path.exists(dst_path):
            os.makedirs(dst_path)
    else:
        dst_path = GNS3UTIL_SCRIPTS_DIR

    if os.path.exists(os.path.join(dst_path, filename)):
        if click.confirm(f"The file {filename} is already present in {dst} do you want to overwrite it?"):
            os.remove(os.path.join(dst_path, filename))
    shutil.copy(filename, dst_path)

    click.secho(f"Successfully copied {filename} to {dst_path}.")


@script.command(name="rm")
@click.argument(
    "filename", shell_complete=get_script_names_for_completion, required=True)
@click.pass_context
def remove_script(ctx, filename):
    """Remove script from the gns3util scripts directory"""
    if not os.path.exists(GNS3UTIL_SCRIPTS_DIR):
        raise click.ClickException(f"The scripts directory does not exist at {
                                   GNS3UTIL_SCRIPTS_DIR}. \n Exiting.")
        ctx.exit(1)
    if not os.path.exists(os.path.join(GNS3UTIL_SCRIPTS_DIR, filename)):
        raise click.ClickException(f"The script does not exist in the script directory at {
                                   GNS3UTIL_SCRIPTS_DIR}. \n Exiting.")

    if click.confirm(f"Do you want to delete script {filename}?"):
        os.remove(os.path.join(GNS3UTIL_SCRIPTS_DIR, filename))

    click.secho(f"Successfully removed {filename}.")


@script.command(name="run")
@click.argument(
    "filename", shell_complete=get_script_names_for_completion,  required=True)
@click.pass_context
def run(ctx, filename):
    """Run yml based scripts"""
    if not filename:
        click.secho("Error: No script filename provided. "
                    "Usage: gns3util script run <filename>", err=True)
        ctx.exit(1)

    full_path = os.path.join(GNS3UTIL_SCRIPTS_DIR, filename)

    if not os.path.exists(full_path):
        click.secho(
            f"Error: Script file '{filename}' not found in "
            f"'{GNS3UTIL_SCRIPTS_DIR}'.",
            err=True,
        )
        ctx.exit(1)

    yml, ok = parse_yml(full_path)
    if not ok:
        raise click.ClickException("Invalid yml file: " + str(yml))

    click.secho("Executing script: ", nl=False)
    script_name = "Unnamed Script"
    if yml and "options" in yml and yml["options"]:
        if "name" in yml["options"][0]:
            script_name = yml["options"][0]["name"]
    click.secho(f"{script_name}", bold=True)

    cmds = get_commands(yml)
    for cmd in cmds:
        out, err = run_command(ctx, cmd)
        if GNS3Error.has_error(err):
            GNS3Error.print_error(err)
            return
        print(out)


@script.command(name="run-file")
@click.argument(
    "filename",  type=click.Path(exists=True, readable=True), required=True)
@click.pass_context
def run_file(ctx, filename):
    """Run yml based scripts"""
    yml, ok = parse_yml(filename)
    if not ok:
        raise click.Exception("Invalid yml file: " + str(yml))

    click.secho("Executing script: ", nl=False)
    script_name = "Unnamed Script"
    if yml and "options" in yml and yml["options"]:
        if "name" in yml["options"][0]:
            script_name = yml["options"][0]["name"]
    click.secho(f"{script_name}", bold=True)

    cmds = get_commands(yml)
    for cmd in cmds:
        out, err = run_command(ctx, cmd)
        if GNS3Error.has_error(err):
            GNS3Error.print_error(err)
            return
        print(out)


def _run_shell_completion(ctx, args, incomplete):
    """
    Provides dynamic completion for the 'run' subcommand's 'filename' argument.
    `args` is a list of arguments already provided on the command line.
    `incomplete` is the current string being completed.
    """
    if len(args) > 0:
        return []

    script_names = get_script_names_for_completion()

    suggestions = [s for s in script_names if s.startswith(incomplete)]

    return suggestions


run.shell_completion = _run_shell_completion


def get_commands(yml: dict) -> list:
    commands_list = []
    if "commands" in yml and isinstance(yml["commands"], list):
        for commands_dict in yml["commands"]:
            cmd = []
            cmd_name = commands_dict.get("name")
            if cmd_name and "subcommands" in commands_dict and isinstance(commands_dict["subcommands"], list):
                for sub_command_dict in commands_dict["subcommands"]:
                    sub_cmd_name = sub_command_dict.get("name")
                    if sub_cmd_name:
                        commands_list.append([cmd_name, sub_cmd_name])
    return commands_list


def run_command(ctx, input: list) -> tuple[Any, GNS3Error]:
    if not input or len(input) < 2:
        return None, GNS3Error("Invalid command format in YML.")

    pkg = input[0]
    sub_cmd = input[1]

    if pkg in command_group_package:
        run_cmd_err, out = call_client_method(ctx, pkg, sub_cmd)
        if GNS3Error.has_error(run_cmd_err):
            return None, run_cmd_err
        return out, run_cmd_err
    else:
        return None, GNS3Error(f"Unknown command package: {pkg}")
