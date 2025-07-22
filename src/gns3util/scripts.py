import click
import os
import shutil
from .utils import call_client_method, parse_yml, GNS3Error, replace_vars, call_client_data
from dataclasses import dataclass, field
from typing import List, Dict, Optional, Any
import dacite
import yaml

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


@dataclass
class script_opts:
    name: str = ""
    description: str = ""
    progress_bar: bool = False
    exit_on_fail: bool = True
    iterations_var_name: str = "iteration"


@dataclass
class command_atributes:
    repeat: str = ""
    resolve_id: bool = False


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
                    os.path.join(root, file), script_dir)
                yml_files.append(relative_path)
    return yml_files


def get_script_names_for_completion(ctx: click.Context, param, incomplete):
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

run steps method


  - equals
  - not_equals

  - contains
  - not_contains
  - startswith
  - endswith

  - in
  - not_in

  - greater_than
  - less_than
  - greater_or_equal
  - less_or_equal

"""


@click.group()
def script():
    """script commands."""
    pass


@script.command(name="ls")
@click.pass_context
def list_scripts(ctx: click.Context):
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
@click.argument("filename", type=click.Path(exists=True, readable=True), required=True)
@click.argument("dst", type=str, required=False)
@click.pass_context
def add_script(ctx: click.Context, filename, dst: str):
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
        if click.confirm(
            f"The file {filename} is already present in {
                dst
            } do you want to overwrite it?"
        ):
            os.remove(os.path.join(dst_path, filename))
    shutil.copy(filename, dst_path)

    click.secho(f"Successfully copied {filename} to {dst_path}.")


@script.command(name="rm")
@click.argument(
    "filename", shell_complete=get_script_names_for_completion, required=True
)
@click.pass_context
def remove_script(ctx: click.Context, filename):
    """Remove script from the gns3util scripts directory"""
    if not os.path.exists(GNS3UTIL_SCRIPTS_DIR):
        raise click.ClickException(
            f"The scripts directory does not exist at {
                GNS3UTIL_SCRIPTS_DIR
            }. \n Exiting."
        )
        ctx.exit(1)
    if not os.path.exists(os.path.join(GNS3UTIL_SCRIPTS_DIR, filename)):
        raise click.ClickException(
            f"The script does not exist in the script directory at {
                GNS3UTIL_SCRIPTS_DIR
            }. \n Exiting."
        )

    if click.confirm(f"Do you want to delete script {filename}?"):
        os.remove(os.path.join(GNS3UTIL_SCRIPTS_DIR, filename))

    click.secho(f"Successfully removed {filename}.")


@script.command(name="run")
@click.argument(
    "filename", shell_complete=get_script_names_for_completion, required=True
)
@click.pass_context
def run(ctx: click.Context, filename):
    """Run yml based scripts"""
    if not filename:
        click.secho(
            "Error: No script filename provided. Usage: gns3util script run <filename>",
            err=True,
        )
        ctx.exit(1)

    full_path = os.path.join(GNS3UTIL_SCRIPTS_DIR, filename)

    if not os.path.exists(full_path):
        click.secho(
            f"Error: Script file '{filename}' not found in '{
                GNS3UTIL_SCRIPTS_DIR}'.",
            err=True,
        )
        ctx.exit(1)

    yml, ok = parse_yml(full_path)
    if not ok:
        raise click.ClickException("Invalid yml file: " + str(yml))

    print_script_details(yml)

    # cmds = get_commands(yml)
    # for cmd in cmds:
    #     out, err = run_command(ctx, cmd)
    #     if GNS3Error.has_error(err):
    #         GNS3Error.print_error(err)
    #         return
    #     print(out)


def print_script_details(yml: dict):
    click.secho("Executing script: ", nl=False)
    script_name = "Unnamed Script"
    if yml and "options" in yml and yml["options"]:
        if "name" in yml["options"]:
            script_name = yml["options"]["name"]
    click.secho(f"{script_name}", bold=True)
    if yml and "options" in yml and yml["options"]:
        if "description" in yml["options"]:
            script_description = yml["options"]["description"]
            click.secho(f"{script_description}")


def get_vars(yml: dict) -> list[dict]:
    vars = []
    if yml and "vars" in yml and yml["vars"]:
        for key in yml["vars"].keys():
            vars.append({key: yml["vars"][key]})
    return vars


def get_opts(yml: dict) -> script_opts:
    opts = script_opts()
    if yml and "options" in yml and yml["options"]:
        options = yml["options"]
        for field in script_opts.__dataclass_fields__:
            if field in options:
                setattr(opts, field, options[field])
    return opts


@script.command(name="run-file")
@click.argument("filename", type=click.Path(exists=True, readable=True), required=True)
@click.pass_context
def run_file(ctx: click.Context, filename):
    """Run yml based scripts"""
    yml, ok = parse_yml(filename)
    if not ok:
        raise click.UsageError("Invalid yml file: " + str(yml))

    print_script_details(yml)
    print(yml)

    vars = get_vars(yml)
    print(vars)

    script_obj = load_script(filename)
    import pprint

    pprint.pprint(script_obj)
    cmds = get_commands(ctx, script_obj)
    # thing_with_vars = replace_vars_in_script(script_obj)
    print(cmds)
    # for cmd in cmds:
    #     out, err = run_command(ctx, cmd)
    #     if GNS3Error.has_error(err):
    #         GNS3Error.print_error(err)
    #         return
    #     print(out)


def _run_shell_completion(ctx: click.Context, args, incomplete):
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

id_element_name = {
    "user": ["user_id", "username"],
    "group": ["user_group_id", "name"],
    "role": ["role_id", "name"],
    "privilege": ["privilege_id", "name"],
    "acl-rule": ["ace_id", "path"],
    "template": ["template_id", "name"],
    "project": ["project_id", "name"],
    "compute": ["compute_id", "name"],
    "appliance": ["appliance_id", "name"],
    "pool": ["resource_pool_id", "name"],
}

subcommand_key_map = {
    "user": "users",
    "group": "groups",
    "role": "roles",
    "privilege": "privileges",
    "acl-rule": "acl",
    "template": "templates",
    "project": "projects",
    "compute": "computes",
    "appliance": "appliances",
    "pool": "pools",
}


def resolve_ids(ctx: click.Context, subcommand: str, name: str) -> tuple[str, bool]:
    id = ""
    key = None
    # get method name to get all of the thing like users
    for map_entry in subcommand_key_map.items():
        if map_entry[0] == subcommand:
            key = map_entry[1]
            break
    if not key:
        return "Could not find the method used to resolve this id", False

    cd = call_client_data(ctx=ctx, package="get", method=key)
    get_opts_err, data = call_client_method(cd)
    if GNS3Error.has_error(get_opts_err):
        GNS3Error.print_error(get_opts_err)
        return "", False
    for entry in data:
        for element in id_element_name.items():
            if element[0] == subcommand:
                if entry[element[1][1]] == name:
                    id = entry[element[1][0]]
    if len(id) == 0:
        return f"Failed to resolve the name {name} to a valid id", False
    return id, True


def run_command(ctx: click.Context, input: list) -> tuple[Any, GNS3Error]:
    err = GNS3Error()
    if not input or len(input) < 2:
        err.msg = "Invalid command format in YML."
        return None, err

    pkg = input[0]
    sub_cmd = input[1]

    if pkg in command_group_package:
        run_cmd_err, out = call_client_method(ctx, pkg, sub_cmd)
        if GNS3Error.has_error(run_cmd_err):
            return None, run_cmd_err
        return out, run_cmd_err
    else:
        err.msg = f"Unknown command package: {pkg}"
        return None, err


# --- Dataclasses for YAML structure ---


@dataclass
class Filter:
    field: str
    operator: str
    value: str


@dataclass
class Parameters:
    friendly_name: Optional[str] = None
    username: Optional[str] = None
    email: Optional[str] = None
    id: Optional[str] = None
    description: Optional[str] = None
    filter: Optional[Filter] = None
    delete_all: Optional[bool] = None


@dataclass
class Subcommand:
    name: str
    parameters: Optional[Parameters] = None
    condition: Optional[str] = None


@dataclass
class Command:
    name: str
    repeat: Optional[int] = None
    subcommands: List[Subcommand] = field(default_factory=list)


@dataclass
class JobContent:
    commands: Optional[List[Command]] = None


@dataclass
class Script:
    vars: Dict[str, Any] = field(default_factory=dict)
    options: script_opts = field(default_factory=script_opts)
    jobs: Dict[str, JobContent] = field(default_factory=dict)


def get_commands(ctx: click.Context, script: Script) -> list:
    command_list = []
    for job in script.jobs.items():
        job_content = job[1]
        job_command_list, ok = get_commands_from_job(
            ctx, job_content, script.options)
        command_list.append(job_command_list)
    return command_list


def get_commands_from_job(ctx: click.Context, job: JobContent, opts: script_opts) -> tuple[list, bool]:
    command_list_final = []
    for command in job.commands:
        command_list = []
        for subcommand in command.subcommands:
            if command.repeat is not None and command.repeat > 0:

                # make add thing to list for as often as counter and replaces the {{itoration}} var
                pass

            if subcommand.parameters:
                params = subcommand.parameters
                if (params.id == "" or params.id is None) and params.friendly_name is not None:
                    if command.repeat is not None and command.repeat > 0:
                        for i in range(command.repeat):
                            iterator_var_name_list = []
                            iterator_var_name_list.append(i)
                            name = replace_vars(
                                params.friendly_name, iterator_var_name_list, replace_iterations=True, iteration_var_name=opts.iterations_var_name)
                            id, ok = resolve_ids(
                                ctx, subcommand.name, name)
                            if not ok:
                                if id != "":
                                    click.secho(id, err=True)
                                return command_list_final, False
                            command_list.append(command.name)
                            command_list.append(subcommand.name)
                            command_list.append(id)

                    else:
                        id, ok = resolve_ids(
                            ctx, subcommand.name, params.friendly_name)
                        if not ok:
                            if id != "":
                                click.secho(id, err=True)
                            return command_list_final, False
                    command_list.append(command.name)
                    command_list.append(subcommand.name)
                    command_list.append(id)

            command_list_final.append(command_list)

        return command_list_final, True

# def replace_vars_in_script(script: Script) -> Script:
#     current_script = script
#     for job in current_script.jobs:
#         print(job)
#     return current_script


def load_script(path: str) -> Optional[Script]:
    try:
        with open(path) as f:
            data = yaml.safe_load(f)
        if not isinstance(data, dict):
            raise ValueError("YAML root is not a dict")

        def fix_parameters(d):
            if not d:
                return None
            return dacite.from_dict(Parameters, d)

        def fix_subcommands(lst):
            out = []
            for sc in lst:
                if "name" not in sc:
                    raise ValueError("Subcommand missing 'name' field.")
                params = fix_parameters(sc.get("parameters"))
                out.append(
                    Subcommand(
                        name=sc["name"],
                        parameters=params,
                        condition=sc.get("condition"),
                    )
                )
            return out

        def fix_commands(lst):
            out = []
            for c in lst:
                if "name" not in c:
                    raise ValueError("Command missing 'name' field.")
                subcommands = fix_subcommands(c.get("subcommands", []))
                repeat = c.get("repeat")
                out.append(
                    Command(
                        name=c["name"],
                        repeat=repeat,
                        subcommands=subcommands,
                    )
                )
            return out

        parsed_jobs_dict = {}
        raw_jobs_list = data.get("jobs", [])

        if not isinstance(raw_jobs_list, list):
            raise ValueError("'jobs' section must be a list.")

        for job_item_dict in raw_jobs_list:
            if not isinstance(job_item_dict, dict) or len(job_item_dict) != 1:
                raise ValueError(
                    "Each item in 'jobs' list must be a single-key dictionary representing a job."
                )

            job_name, job_content_raw = list(job_item_dict.items())[0]

            if job_content_raw is None:
                parsed_jobs_dict[job_name] = JobContent(commands=None)
                continue

            if not isinstance(job_content_raw, dict):
                raise ValueError(f"Content for job '{
                                 job_name}' must be a dictionary.")

            commands_data = job_content_raw.get("commands")

            fixed_commands_list = None
            if commands_data is not None:
                if not isinstance(commands_data, list):
                    raise ValueError(f"Commands for job '{
                                     job_name}' must be a list.")
                fixed_commands_list = fix_commands(commands_data)

            parsed_jobs_dict[job_name] = JobContent(
                commands=fixed_commands_list)

        opts = get_opts(data)

        script_data_for_dacite = {
            "vars": data.get("vars", {}),
            "options": opts,
            "jobs": parsed_jobs_dict,
        }

        return dacite.from_dict(Script, script_data_for_dacite)

    except (yaml.YAMLError, KeyError, TypeError, dacite.DaciteError, ValueError) as e:
        print(f"Error parsing YAML script: {e}")
        return None
        raise ValueError(f"Content for job '{
            job_name}' must be a dictionary.")

        commands_data = job_content_raw.get("commands")

        fixed_commands_list = None
        if commands_data is not None:
            if not isinstance(commands_data, list):
                raise ValueError(f"Commands for job '{
                    job_name}' must be a list.")
                fixed_commands_list = fix_commands(commands_data)

            parsed_jobs_dict[job_name] = JobContent(
                commands=fixed_commands_list)

        script_data_for_dacite = {
            "vars": data.get("vars", {}),
            "options": data.get("options", {}),
            "jobs": parsed_jobs_dict,
        }

        return dacite.from_dict(Script, script_data_for_dacite)

    except (yaml.YAMLError, KeyError, TypeError, dacite.DaciteError, ValueError) as e:
        print(f"Error parsing YAML script: {e}")
        return None
