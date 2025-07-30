import json
import re
import yaml
import sys
from enum import Enum
import os
import importlib
import uuid
import getpass
import rich
import click
from dataclasses import dataclass, field
import subprocess
from typing import Callable, Any, Optional
from .api.client import GNS3Error
from InquirerPy import inquirer
from .api.schemas_map import RESPONSE_SCHEMA_MAP
from pydantic import ValidationError, TypeAdapter
from dacite import from_dict
from gns3util.schemas import (
    User,
    UserCreate,
    UserGroup,
    UserGroupCreate,
    Role,
    Project,
    ProjectCreate,
    ResourcePool,
    ResourcePoolCreate,
    ACE,
    ACECreate,
    UserUpdate,
    Version,
)


@dataclass
class fuzzy_info_params:
    ctx: click.Context
    client: Callable[[Any], Any]
    opt_method: Optional[str] = None
    opt_key: Optional[str] = None
    is_class: Optional[bool] = False
    method: str = "str"
    key: str = "str"
    multi: bool = False
    opt_data: bool = False


@dataclass
class fuzzy_password_params:
    ctx: click.Context
    client: Callable[[Any], Any]
    method: str = "str"
    key: str = "str"
    multi: bool = False


@dataclass
class fuzzy_delete_class_params:
    ctx: click.Context
    client: Callable[[Any], Any]
    method: str = "groups"
    key: str = "name"
    multi: bool = False
    confirm: bool = True
    non_interactive: str = None
    delete_all: bool = False
    delete_exercises: bool = False


@dataclass
class fuzzy_delete_exercise_params:
    ctx: click.Context
    client: Callable[[Any], Any]
    method: str = "projects"
    key: str = "name"
    multi: bool = False
    confirm: bool = True
    non_interactive: str = None
    unattended: bool = False
    class_to_use: str = None
    group_to_use: str = None
    select_class: bool = False
    select_group: bool = False
    delete_all: bool = False


class fuzzy_params_type(Enum):
    user_info = 1
    group_info = 2
    group_info_with_usernames = 3
    user_info_and_group_membership = 4


@dataclass
class call_client_data:
    ctx: click.Context
    package: str
    method: str
    args: Optional[list[Any]] = field(default_factory=list)
    has_body_data: bool = False


@dataclass
class Student:
    fullName: Optional[str]
    userName: str
    password: str
    email: Optional[str]


@dataclass
class Group:
    name: str
    students: list[Student]


@dataclass
class Class:
    name: str
    groups: list[Group]


@dataclass
class group_list_element:
    id: uuid.UUID
    number: str
    name: str


@dataclass
class Exercise:
    name: str
    id: str
    class_name: str
    group_number: str


@dataclass
class Selected_exercise:
    exercise_name: str
    class_name: str | None
    group_number: str | None


@dataclass
class Exercise_pool:
    pool_id: str
    class_name: str
    group_number: str


# todo make setting what fuzzy finder to use like skim instead
def fzf_select(options: list[str], multi=False):
    """
    Opens an fzf window with the given options and returns the selected option(s). Falls back to `inquirer` if fzf is not installed.

    Args:
        options: A list of strings representing the options to choose from.

    Returns:
        A list of strings containing the selected option(s), or an empty list if none were selected.
    """
    try:
        if multi:
            fzf_process = subprocess.Popen(
                ["fzf", "--multi"],
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
            )
        else:
            fzf_process = subprocess.Popen(
                ["fzf"],
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
            )

        output, error = fzf_process.communicate("\n".join(options))
        return_code = fzf_process.returncode

        if return_code != 0:
            click.secho("Aborted!")
            sys.exit(1)

        if error:
            return get_selection_inquirerpy(options, multi)

        if output:
            return [line.strip() for line in output.strip().split("\n")]
        else:
            return []

    except FileNotFoundError:
        return get_selection_inquirerpy(options, multi)


def get_selection_inquirerpy(options, multi=False):
    if multi:
        result = inquirer.checkbox(
            message="Select options:", choices=options, cycle=True
        ).execute()
    else:
        result = inquirer.select(
            message="Select an option:", choices=options, cycle=True
        ).execute()
    return result if isinstance(result, list) else [result]


def call_client_method(call_data: call_client_data) -> tuple[GNS3Error, Any]:
    module = importlib.import_module(f".{call_data.package}", package=__package__)
    client = module.get_client(call_data.ctx)
    method = getattr(client, call_data.method)
    return method(*call_data.args)


def print_key_value_with_secho(key, value, color="cyan", reset="reset"):
    click.secho(f"{key}: ", fg=color, nl=False)
    click.secho(value)


def print_separator_with_secho(color="white"):
    click.secho("---", fg=color)


def fuzzy_info(params=fuzzy_info_params) -> GNS3Error:
    error = GNS3Error()
    fzf_input_data, api_data, get_fzf_input_error = get_values_for_fuzzy_input(params)
    if GNS3Error.has_error(get_fzf_input_error):
        GNS3Error.print_error(get_fzf_input_error)
        return get_fzf_input_error
    selected = fzf_select(fzf_input_data, multi=params.multi)
    matched = set()
    for selected_item in selected:
        for data in api_data:
            if (
                getattr(data, params.key) == selected_item
                and getattr(data, params.key) not in matched
            ):
                print_separator_with_secho()
                for k, v in data.dict().items():
                    print_key_value_with_secho(k, v)
                print_separator_with_secho()
                if params.opt_data:
                    opt_data_error, opt_data = getattr(
                        params.client(params.ctx), params.opt_method
                    )(getattr(data, params.opt_key))
                    if GNS3Error.has_error(opt_data_error):
                        GNS3Error.print_error(opt_data_error)
                        error.request = True
                        return error
                    if opt_data == []:
                        click.secho(
                            f"Empty data returned from method {
                                params.opt_method
                            } for the {getattr(data, params.opt_key)} value",
                            bold=True,
                            err=True,
                        )
                    else:
                        for d in opt_data:
                            print_separator_with_secho()
                            for k2, v2 in d.items():
                                print_key_value_with_secho(k2, v2)
                            print_separator_with_secho()
                break
    return error


def get_values_for_fuzzy_input(params) -> (list[str], list, GNS3Error):
    cd = call_client_data(ctx=params.ctx, package="get", method=params.method)
    get_data_error, api_data_raw = call_client_method(cd)
    fuzzy_error = GNS3Error()
    if GNS3Error.has_error(get_data_error):
        fuzzy_error.connection = True
        return [""], None, fuzzy_error
    api_data = validate_response(params.method, api_data_raw)
    fzf_input_data = []
    for data in api_data:
        fzf_input_data.append(getattr(data, params.key))
    return fzf_input_data, api_data, fuzzy_error


def get_password():
    while True:
        password = getpass.getpass(prompt="Enter your password: ")

        if 8 <= len(password) <= 100:
            if any(char.isdigit() for char in password):
                return password
            else:
                click.echo("Password must contain at least one number.")
        else:
            click.echo("Password must be between 8 and 100 characters long.")


def fuzzy_change_password(params=fuzzy_password_params) -> GNS3Error:
    fzf_input_data, api_data_raw, get_fzf_input_error = get_values_for_fuzzy_input(
        params
    )
    if GNS3Error.has_error(get_fzf_input_error):
        return get_fzf_input_error
    selected = fzf_select(fzf_input_data, multi=params.multi)
    matched = set()
    api_data: list[User] = validate_response(params.method, api_data_raw)
    for selected_item in selected:
        for data in api_data:
            if data.username == selected_item and data.username not in matched:
                click.secho("Changing the password for user ", nl=False)
                click.secho(f"{data.username}", bold=True)
                pw = get_password()
                user_update_data = {"password": pw}
                cd = call_client_data(
                    ctx=params.ctx,
                    package="update",
                    method="update_user",
                    args=[data.user_id, user_update_data],
                )
                change_password_error, result = call_client_method(cd)
                if GNS3Error.has_error(change_password_error):
                    return change_password_error
                click.secho("Success: ", nl=False, fg="green")
                click.secho("changed the password for user ", nl=False)
                click.secho(f"{data.username}", bold=True)
                break
    return change_password_error


def fuzzy_delete_class(params=fuzzy_delete_class_params) -> GNS3Error:
    error = GNS3Error()
    class_names: list[str] = []
    class_ids: list[str] = []
    selected = []

    fzf_input_data, groups_raw, get_fzf_input_error = get_values_for_fuzzy_input(params)
    if GNS3Error.has_error(get_fzf_input_error) and groups_raw is None:
        return get_fzf_input_error

    groups: list[UserGroup] = validate_response(params.method, groups_raw)

    class_names, class_ids, get_classes_error = get_classes(groups)
    if GNS3Error.has_error(get_classes_error):
        return get_classes_error

    if not class_names:
        click.secho("No classes available to delete", err=True)
        return error

    if params.non_interactive is None and params.delete_all is False:
        # Interactive mode
        selected = fzf_select(class_names, multi=params.multi)
    elif params.delete_all:
        selected.extend(class_names)
    else:
        # Non-interactive mode
        if params.non_interactive in class_names:
            selected = [params.non_interactive]
        else:
            click.secho("Error: ", fg="red", nl=False, err=True)
            click.secho(f"Class ", nl=False, err=True)
            click.secho(f"{params.non_interactive} ", bold=True, nl=False, err=True)
            click.secho("not found.", err=True)
            return error

    for selected_item in selected:
        if params.confirm:
            if not click.confirm(f"Do you want to delete the class {selected_item}?"):
                click.secho("Deletion aborted.")
                continue

        error = delete_class(params, selected_item, class_names, class_ids)
        if GNS3Error.has_error(error):
            return error
        if params.delete_exercises:
            params_del = fuzzy_delete_exercise_params(
                ctx=params.ctx,
                client=params.client,
                method="projects",
                key="name",
                multi=False,
                confirm=False,
                non_interactive=None,
                unattended=True,
                class_to_use=selected_item,
                group_to_use=None,
                select_class=False,
                select_group=False,
                delete_all=False,
            )
            fuzzy_delete_exercise(params_del)
            if GNS3Error.has_error(error):
                return error

    return error


def get_classes(input: list[UserGroup]) -> tuple[list[str], list[str], GNS3Error]:
    error = GNS3Error()
    classes = []
    ids = []
    seen_classes = set()

    for data in input:
        split = data.name.split("-")
        if len(split) == 3:
            class_name = split[0]
            for classes_data in input:
                if class_name == classes_data.name:
                    if class_name not in seen_classes:
                        id = classes_data.user_group_id
                        classes.append(class_name)
                        ids.append(str(id))
                        seen_classes.add(class_name)

    return classes, ids, error


def get_exercises(input: list[Project]) -> tuple[list[Exercise], GNS3Error]:
    error = GNS3Error()
    exercises: list[Exercise] = []

    for data in input:
        split = data.name.split("-")
        # Only process if we have at least 4 parts (class-exercise-group-uuid)
        if len(split) >= 4:
            exercise_name = split[1]
            id = data.project_id
            exercise_data = Exercise(
                name=exercise_name,
                id=id,
                class_name=split[0],
                group_number=split[2],
            )

            exercises.append(exercise_data)

    return exercises, error


def delete_class(
    params: fuzzy_delete_class_params,
    selected_item: str,
    class_names: list,
    class_ids: list,
) -> GNS3Error:
    groups_to_delete = []
    students_to_delete = []
    groups, get_groups_in_class_error = get_groups_in_class(params.ctx, selected_item)
    if GNS3Error.has_error(get_groups_in_class_error):
        return get_groups_in_class_error
    for group in groups:
        groups_to_delete.append(group.id)
    for i, class_name in enumerate(class_names):
        if selected_item == class_name:
            groups_to_delete.append(class_ids[i])

    for group in groups_to_delete:
        members_id, get_members_error = get_group_members(
            params.ctx, group, id_only=True
        )
        if GNS3Error.has_error(get_members_error):
            return get_members_error
        students_to_delete.extend(members_id)

    student_ids = list(set(students_to_delete))
    for student_id in student_ids:
        delete_user_error = delete_from_id(params.ctx, "delete_user", student_id)
        if GNS3Error.has_error(delete_user_error):
            return delete_user_error

    group_ids = list(set(groups_to_delete))
    for group_id in group_ids:
        delete_groups_error = delete_from_id(params.ctx, "delete_group", group_id)
        if GNS3Error.has_error(delete_groups_error):
            return delete_groups_error

    click.secho("Success: ", nl=False, fg="green")
    click.secho("deleted the class ", nl=False)
    click.secho(f"{selected_item}", bold=True)
    return GNS3Error()


def delete_exercise(
    params: fuzzy_delete_exercise_params,
    selected_item: Selected_exercise,
    exercises: list[Exercise],
) -> GNS3Error:
    projects_to_delete = []
    pools_to_delete = []
    acls_to_delete = []
    for exercise in exercises:
        projects_to_delete.append(exercise.id)

    pools, get_pools_error = get_pools_for_exercise(
        params.ctx, selected_item.exercise_name
    )
    if GNS3Error.has_error(get_pools_error):
        return get_pools_error
    for pool in pools:
        if selected_item.class_name:
            if pool.class_name != selected_item.class_name:
                continue
            if selected_item.group_number:
                if pool.group_number != selected_item.group_number:
                    continue
        pools_to_delete.append(pool.pool_id)

    acls_to_delete, get_acls_error = get_acls_for_exercise(params.ctx, pools_to_delete)
    if GNS3Error.has_error(get_acls_error):
        return get_acls_error

    project_ids = list(set(projects_to_delete))
    for project_id in project_ids:
        close_project_error = close_project(params.ctx, project_id)
        if GNS3Error.has_error(close_project_error):
            return close_project_error
        delete_project_error = delete_from_id(params.ctx, "delete_project", project_id)
        if GNS3Error.has_error(delete_project_error):
            return delete_project_error

    pools_ids = list(set(pools_to_delete))
    for pool_id in pools_ids:
        delete_pool_error = delete_from_id(params.ctx, "delete_pool", pool_id)
        if GNS3Error.has_error(delete_pool_error):
            return delete_pool_error

    acl_ids = list(set(acls_to_delete))
    for ace_id in acl_ids:
        delete_ace_error = delete_from_id(params.ctx, "delete_ace", ace_id)
        if GNS3Error.has_error(delete_ace_error):
            return delete_ace_error

    click.secho("Success: ", nl=False, fg="green")
    click.secho("deleted the exercise ", nl=False)
    click.secho(f"{selected_item.exercise_name}", bold=True)
    return GNS3Error()


def fuzzy_delete_exercise(params=fuzzy_delete_exercise_params) -> GNS3Error:
    error = GNS3Error()
    selected: list[Selected_exercise] = []

    fzf_input_data, api_data_raw, get_fzf_input_error = get_values_for_fuzzy_input(
        params
    )
    if GNS3Error.has_error(get_fzf_input_error):
        return get_fzf_input_error
    api_data: list[Project] = validate_response(params.method, api_data_raw)

    exercises, get_exercies_error = get_exercises(api_data)
    if GNS3Error.has_error(get_exercies_error):
        return get_exercies_error

    if not exercises:
        click.secho("No exercises available to delete", err=True)
        return GNS3Error()

    if (
        params.non_interactive is None
        and params.delete_all is False
        and params.unattended is False
    ):
        # Interactive mode
        exercise_names = []
        selected_exercise_names = []
        selected_exercise_names
        for exercise in exercises:
            exercise_names.append(exercise.name)
        selected_exercise_names = fzf_select(
            list(set(exercise_names)), multi=params.multi
        )

        # Interactively selecting the class and group
        if params.select_class:
            exercise_class_names = []
            selected_class = []
            selected_group = []
            for exercise in exercises:
                if exercise.name == selected_exercise_names[0]:
                    exercise_class_names.append(exercise.class_name)
            selected_class = fzf_select(
                list(set(exercise_class_names)), multi=params.multi
            )
            if params.select_group:
                exercise_group_numbers = []
                for exercise in exercises:
                    if (
                        exercise.name == selected_exercise_names[0]
                        and exercise.class_name == selected_class[0]
                    ):
                        exercise_group_numbers.append(exercise.group_number)
                selected_group = fzf_select(
                    list(set(exercise_group_numbers)), multi=params.multi
                )

            if not selected_group:
                selected_group.append(None)
            selected_exercise_data = Selected_exercise(
                exercise_name=selected_exercise_names[0],
                class_name=selected_class[0],
                group_number=selected_group[0],
            )
            selected.append(selected_exercise_data)
        else:
            for selected_exercise in selected_exercise_names:
                selected_exercise_data = Selected_exercise(
                    exercise_name=selected_exercise,
                    class_name=None,
                    group_number=None,
                )
                selected.append(selected_exercise_data)

    elif params.delete_all:
        exercise_name_list = []
        for exercise in exercises:
            exercise_name_list.append(exercise.name)
        exercise_name_list = set(exercise_name_list)
        for exercise in exercise_name_list:
            selected_exercise_data = Selected_exercise(
                exercise_name=exercise,
                class_name=None,
                group_number=None,
            )
            selected.append(selected_exercise_data)
    elif params.unattended:
        found = False
        exercise_name_list = []
        for exercise in exercises:
            if exercise.class_name == params.class_to_use:
                exercise_name_list.append(exercise.name)
        exercise_name_list = set(exercise_name_list)
        for exercise in exercise_name_list:
            found = True
            selected_exercise_data = Selected_exercise(
                exercise_name=exercise,
                class_name=params.class_to_use,
                group_number=None,
            )
            selected.append(selected_exercise_data)

        if not found:
            click.secho("Error: ", fg="red", nl=False, err=True)
            click.secho(f"No exercises for class ", nl=False, err=True)
            click.secho(f"{params.class_to_use} ", bold=True, nl=False, err=True)
            click.secho("found.", err=True)
            return error
    else:
        # Non-interactive mode
        found = False
        for exercise in exercises:
            if params.non_interactive != exercise.name:
                continue

            if params.class_to_use:
                if params.class_to_use != exercise.class_name:
                    continue

            if params.group_to_use:
                if params.group_to_use != exercise.group_number:
                    continue

            found = True

        if not found:
            click.secho("Error: ", fg="red", nl=False, err=True)
            click.secho(f"Exercise ", nl=False, err=True)
            click.secho(f"{params.non_interactive} ", bold=True, nl=False, err=True)
            click.secho("not found.", err=True)
            return error

        selected_exercise_data = Selected_exercise(
            exercise_name=params.non_interactive,
            class_name=params.class_to_use,
            group_number=params.group_to_use,
        )
        selected.append(selected_exercise_data)

    for selected_item in selected:
        if params.confirm:
            if not click.confirm(
                f"Do you want to delete the exercise {selected_item.exercise_name}?"
            ):
                click.secho("Deletion aborted.")
                continue
        exercises_to_delete: list[Exercise] = []
        for exercise in exercises:
            if exercise.name != selected_item.exercise_name:
                continue

            # If a class filter is provided then it must match.
            if selected_item.class_name:
                if exercise.class_name != selected_item.class_name:
                    continue

                # If a group filter is also provided then it must match.
                if selected_item.group_number:
                    if exercise.group_number != selected_item.group_number:
                        continue

            exercises_to_delete.append(exercise)

        error = delete_exercise(params, selected_item, exercises_to_delete)
        if GNS3Error.has_error(error):
            return error

    return error


def get_group_members(
    ctx: click.Context, group_id: str, id_only=False
) -> tuple[list[User] | list[str], GNS3Error]:
    member_ids: list[str] = []
    cd = call_client_data(
        ctx=ctx, package="get", method="group_members", args=[group_id]
    )
    get_members_error, members_raw = call_client_method(cd)
    if GNS3Error.has_error(get_members_error):
        return [], get_members_error
    members: list[User] = validate_response(cd.method, members_raw)
    if id_only:
        for member in members:
            member_ids.append(str(member.user_id))
        return member_ids, get_members_error
    else:
        return members, get_members_error


def delete_from_id(ctx: click.Context, method: str, id: str) -> GNS3Error:
    cd = call_client_data(ctx=ctx, package="delete", method=method, args=[id])
    delete_error, _ = call_client_method(cd)
    return delete_error


def parse_yml(filepath: str) -> tuple[Any, bool]:
    try:
        with open(filepath, "r") as f:
            file_content = f.read().strip()
        if not file_content:
            return "YAML file is empty.", False
        data = yaml.safe_load(file_content)
        if data is not None and file_content:
            return data, True
    except FileNotFoundError:
        return f"File not found: {filepath}", False
    except yaml.YAMLError as e:
        return f"Invalid yml in {filepath}: {e}", False
    except Exception as e:
        return f"An unexpected error occurred: {e}", False


def parse_json(filepath: str) -> tuple[bool, Any]:
    try:
        with open(filepath, "r") as f:
            data = json.load(f)
        return False, data
    except FileNotFoundError:
        return True, f"File not found: {filepath}"
    except json.JSONDecodeError as e:
        return True, f"Invalid JSON in {filepath}: {e}"
    except Exception as e:
        return True, f"An unexpected error occurred: {e}"


def add_user_to_group(ctx: click.Context, user_id: str, group_id: str) -> GNS3Error:
    cd = call_client_data(
        ctx=ctx, package="add", method="add_group_member", args=[group_id, user_id]
    )
    add_user_to_group_error, result = call_client_method(cd)
    if GNS3Error.has_error(add_user_to_group_error):
        return add_user_to_group_error
    return add_user_to_group_error


def create_class(
    ctx: click.Context, filename: str = None, data_input: dict = None
) -> tuple[str, bool]:
    if filename == None:
        data_raw = data_input
        data = from_dict(data_class=Class, data=data_raw)
    else:
        error_load, data_raw = parse_json(filename)
        data = from_dict(data_class=Class, data=data_raw)
        if error_load:
            click.secho("Error: ", nl=False, fg="red", err=True)
            click.secho("Failed to load file: ", nl=False, err=True)
            click.secho(f"{data}", bold=True, err=True)
            return "", False

    class_obj = data.groups
    class_creation_data = UserGroupCreate(name=data.name)
    class_raw, create_group_error = create_user_group(ctx, class_creation_data)
    if GNS3Error.has_error(create_group_error) and class_raw == None:
        GNS3Error.print_error(create_group_error)
        return data.name, False
    class_id = class_raw.user_group_id
    click.secho("Success: ", nl=False, fg="green")
    click.secho(f"created the group {data.name}")
    for group in class_obj:
        group_creation_data = UserGroupCreate(name=group.name)
        group_raw, create_group_error = create_user_group(ctx, group_creation_data)
        if GNS3Error.has_error(create_group_error) and group_raw == None:
            GNS3Error.print_error(create_group_error)
            return data.name, False
        group_id = group_raw.user_group_id
        click.secho("Success: ", nl=False, fg="green")
        click.secho(f"created the group {group.name}")
        for student in group.students:
            if student.fullName != "":
                user_creation_data = UserCreate(
                    username=student.userName,
                    password=student.password,
                    email=student.email,
                    full_name=student.fullName,
                )
            else:
                user_creation_data = UserCreate(
                    username=student.userName,
                    password=student.password,
                    email=student.email,
                )
            user_data, create_user_error = create_user(ctx, user_creation_data)
            if GNS3Error.has_error(create_user_error) and user_data == None:
                GNS3Error.print_error(create_user_error)
                return data.name, False
            user_id = user_data.user_id
            click.secho("Success: ", nl=False, fg="green")
            click.secho(f"created the user {student.userName}")
            add_user_to_class_error = add_user_to_group(ctx, user_id, class_id)
            if GNS3Error.has_error(add_user_to_class_error):
                GNS3Error.print_error(add_user_to_class_error)
                return data.name, False
            add_user_to_group_error = add_user_to_group(ctx, user_id, group_id)
            if GNS3Error.has_error(add_user_to_group_error):
                GNS3Error.print_error(add_user_to_group_error)
                return data.name, False
    return data.name, True


def create_user_group(
    ctx: click.Context, group_creation_data: UserGroupCreate
) -> tuple[UserGroup | None, GNS3Error]:
    cd = call_client_data(
        ctx=ctx,
        package="create",
        method="create_group",
        has_body_data=True,
        args=[group_creation_data.dict()],
    )
    create_group_error, result_raw = call_client_method(cd)
    if GNS3Error.has_error(create_group_error):
        return None, create_group_error
    result: UserGroup = validate_response("create_group", result_raw)
    return result, create_group_error


def create_user(
    ctx: click.Context, user_data: UserCreate
) -> tuple[User | None, GNS3Error]:
    cd = call_client_data(
        ctx=ctx,
        package="create",
        method="create_user",
        has_body_data=True,
        # for stuff with secret use things to dump
        # https://docs.pydantic.dev/2.0/usage/types/secrets/
        args=[json.loads(user_data.model_dump_json())],
    )
    create_user_error, result_raw = call_client_method(cd)

    if GNS3Error.has_error(create_user_error):
        return None, create_user_error
    result: User = validate_response(cd.method, result_raw)
    return result, create_user_error


def get_fuzzy_info_params(
    input: fuzzy_params_type, ctx: click.Context, get_client, multi: bool
) -> fuzzy_info_params:
    if input == fuzzy_params_type.user_info:
        return fuzzy_info_params(
            ctx=ctx,
            client=get_client,
            method="users",
            key="username",
            multi=multi,
            opt_data=False,
        )
    elif input == fuzzy_params_type.group_info:
        return fuzzy_info_params(
            ctx=ctx,
            client=get_client,
            method="groups",
            key="name",
            multi=multi,
            opt_data=False,
        )
    elif input == fuzzy_params_type.group_info_with_usernames:
        return fuzzy_info_params(
            ctx=ctx,
            client=get_client,
            opt_method="group_members",
            opt_key="user_group_id",
            method="groups",
            key="name",
            multi=multi,
            opt_data=True,
        )
    elif input == fuzzy_params_type.user_info_and_group_membership:
        return fuzzy_info_params(
            ctx=ctx,
            client=get_client,
            opt_method="users_groups",
            opt_key="user_id",
            method="users",
            key="username",
            multi=multi,
            opt_data=True,
        )


def fuzzy_info_wrapper(params: fuzzy_info_params):
    error = fuzzy_info(params)
    if GNS3Error.has_error(error):
        if error.connection:
            click.secho("Error: ", fg="red", nl=False, err=True)
            click.secho(
                "Failed to fetch data from the API check your Network connection to the server",
                bold=True,
                err=True,
            )
            return
        GNS3Error.print_error(error)


def fuzzy_delete_class_wrapper(params: fuzzy_delete_class_params):
    error = fuzzy_delete_class(params)
    if GNS3Error.has_error(error):
        if error.connection:
            click.secho("Error: ", fg="red", nl=False, err=True)
            click.secho(
                "Failed to fetch data from the API check your Network connection to the server",
                bold=True,
                err=True,
            )
            return
        GNS3Error.print_error(error)


def fuzzy_delete_exercise_wrapper(params: fuzzy_delete_exercise_params):
    error = fuzzy_delete_exercise(params)
    if GNS3Error.has_error(error):
        if error.connection:
            click.secho("Error: ", fg="red", nl=False, err=True)
            click.secho(
                "Failed to fetch data from the API check your Network connection to the server",
                bold=True,
                err=True,
            )
            return
        GNS3Error.print_error(error)


def fuzzy_put_wrapper(params):
    error = fuzzy_change_password(params)
    if GNS3Error.has_error(error):
        if error.connection:
            click.secho("Error: ", fg="red", nl=False, err=True)
            click.secho(
                "Failed to fetch data from the API check your Network connection to the server",
                bold=True,
                err=True,
            )
            return
        GNS3Error.print_error(error)


def execute_and_print(ctx: click.Context, client, func):
    error, data = func(client)
    if GNS3Error.has_error(error):
        GNS3Error.print_error(error)
    else:
        rich.print_json(json.dumps(data, indent=2))


def get_role_id(ctx: click.Context, name: str) -> tuple[uuid.UUID, GNS3Error]:
    cd = call_client_data(ctx=ctx, package="get", method="roles")
    get_roles_error, roles_raw = call_client_method(cd)
    if GNS3Error.has_error(get_roles_error):
        return None, get_roles_error
    roles: list[Role] = validate_response("roles", roles_raw)
    for role in roles:
        if role.name == name:
            return role.role_id, get_roles_error
    get_roles_error.not_found = True
    get_roles_error.msg = f"No Matching role with the name {name} found"
    return None, get_roles_error


def create_project(
    ctx: click.Context, data: ProjectCreate
) -> (Project | None, GNS3Error):
    cd = call_client_data(
        ctx=ctx, package="create", method="create_project", args=[data.dict()]
    )
    create_project_error, result_raw = call_client_method(cd)
    if GNS3Error.has_error(create_project_error):
        return None, create_project_error
    result: Project = validate_response(cd.method, result_raw)
    close_project_error = close_project(ctx, str(result.project_id))
    if GNS3Error.has_error(close_project_error):
        return result, close_project_error
    return result, create_project_error


def close_project(ctx: click.Context, project_id: str) -> GNS3Error:
    cd = call_client_data(
        ctx=ctx, package="post", method="close_project", args=[project_id]
    )
    close_project_error, _ = call_client_method(cd)
    if GNS3Error.has_error(close_project_error):
        return close_project_error
    return close_project_error


def get_groups_in_class(
    ctx: click.Context, class_name: str
) -> tuple[list[group_list_element], GNS3Error]:
    group_list: list[group_list_element] = []
    cd = call_client_data(ctx=ctx, package="get", method="groups")
    get_groups_error, groups_raw = call_client_method(cd)
    if GNS3Error.has_error(get_groups_error):
        return group_list, get_groups_error
    groups: list[UserGroup] = validate_response("groups", groups_raw)
    for group in groups:
        if class_name in group.name and class_name != group.name:
            group_number = group.name.split("-")[-1]
            group_list_entry = group_list_element(
                id=group.user_group_id,
                number=group_number,
                name=group.name,
            )
            group_list.append(group_list_entry)

    return group_list, get_groups_error


def get_pools_for_exercise(
    ctx: click.Context, exercise_name: str
) -> tuple[list[Exercise_pool], GNS3Error]:
    pool_list: list[Exercise_pool] = []
    cd = call_client_data(ctx=ctx, package="get", method="pools")
    get_pools_error, pools_raw = call_client_method(cd)
    if GNS3Error.has_error(get_pools_error) or len(pools_raw) == 0:
        if len(pools_raw) == 0:
            click.secho(
                f"No pools were found for the exercise {
                    exercise_name
                } meaning it was probaly half manually deleted. Please clear the reaming projects and ACL's manually.",
                err=True,
            )
            ctx.exit(1)

        return pool_list, get_pools_error
    pools: list[ResourcePool] = validate_response(cd.method, pools_raw)
    for pool in pools:
        split = pool.name.split("-")
        # Match pools with at least 5 parts (class-exercise-group-pool-uuid)
        if exercise_name in pool.name and len(split) >= 5:
            group_number = split[2]
            classname = split[0]
            pool_data = Exercise_pool(
                pool_id=str(pool.resource_pool_id),
                class_name=classname,
                group_number=group_number,
            )
            pool_list.append(pool_data)
    if len(pool_list) == 0:
        return pool_list, get_pools_error
    return pool_list, get_pools_error


def get_acls_for_exercise(
    ctx: click.Context, pools: list[str]
) -> (list[str], GNS3Error):
    acls_list = []
    cd = call_client_data(ctx=ctx, package="get", method="acl")
    get_alcs_error, acls_raw = call_client_method(cd)
    if GNS3Error.has_error(get_alcs_error) or len(acls_raw) == 0:
        if len(acls_raw) == 0:
            click.secho(
                "No ACL's were found for this exercise meaning it was probaly half manually deleted. Please clear the reaming projects and pools manually.",
                err=True,
            )
            ctx.exit(1)

        return acls_list, get_alcs_error
    acls: list[ACE] = validate_response(cd.method, acls_raw)
    for acl in acls:
        first_slash_index = acl.path.find("/")
        second_slash_index = acl.path.find("/", first_slash_index + 1)
        path_ressouce_id = acl.path[second_slash_index + 1 :]

        for pool in pools:
            if path_ressouce_id == pool:
                acls_list.append(acl.ace_id)

    return acls_list, get_alcs_error


def create_acl(ctx: click.Context, data: ACECreate) -> tuple[ACE | None, GNS3Error]:
    cd = call_client_data(
        ctx=ctx,
        package="create",
        method="create_acl",
        args=[json.loads(data.model_dump_json())],
    )
    create_acl_error, result_raw = call_client_method(cd)
    if GNS3Error.has_error(create_acl_error):
        return None, create_acl_error
    result: ACE = validate_response(cd.method, result_raw)
    return result, create_acl_error


def create_pool(
    ctx: click.Context, data: ResourcePoolCreate
) -> tuple[ResourcePool | None, GNS3Error]:
    cd = call_client_data(
        ctx=ctx, package="create", method="create_pool", args=[data.dict()]
    )
    create_pool_error, result_raw = call_client_method(cd)
    if GNS3Error.has_error(create_pool_error):
        return None, create_pool_error
    result: ResourcePool = validate_response(cd.method, result_raw)
    return result, create_pool_error


def add_resource_to_pool(
    ctx: click.Context, pool_id: str, resource_id: str
) -> GNS3Error:
    cd = call_client_data(
        ctx=ctx,
        package="add",
        method="add_resource_to_pool",
        args=[pool_id, resource_id],
    )
    add_to_pool_error, result = call_client_method(cd)
    return add_to_pool_error


def create_Exercise(ctx: click.Context, class_name: str, exercise_name: str) -> bool:
    role_id, get_role_id_error = get_role_id(ctx, "User")
    if GNS3Error.has_error(get_role_id_error):
        GNS3Error.print_error(get_role_id_error)
        return False

    groups, get_groups_error = get_groups_in_class(ctx, class_name)
    for group in groups:
        split = group.name.split("-")
        if len(split) != 3:
            continue
        uuid_suffix = str(uuid.uuid4())[:8]
        project_name = f"{class_name}-{exercise_name}-{group.number}-{uuid_suffix}"
        pool_name = f"{class_name}-{exercise_name}-{group.number}-pool-{uuid_suffix}"
        project_creation_data = ProjectCreate(name=project_name)
        project_raw, create_project_error = create_project(ctx, project_creation_data)
        if GNS3Error.has_error(create_project_error) and project_raw == None:
            GNS3Error.print_error(create_project_error)
            return False
        project_id = project_raw.project_id
        click.secho("Success: ", nl=False, fg="green")
        click.secho(f"created the project {project_name}")

        pool_creation_data = ResourcePoolCreate(name=pool_name)
        pool_raw, create_pool_error = create_pool(ctx, pool_creation_data)
        if GNS3Error.has_error(create_pool_error):
            GNS3Error.print_error(create_pool_error)
            return False
        pool_id = pool_raw.resource_pool_id

        add_to_pool_error = add_resource_to_pool(ctx, str(pool_id), project_id)
        if GNS3Error.has_error(add_to_pool_error):
            if add_to_pool_error.not_found:
                GNS3Error.print_error(
                    add_to_pool_error, "project id: " + project_id, "pool_id: missing"
                )
            else:
                GNS3Error.print_error(add_to_pool_error)
            return False

        create_ace_data = ACECreate(
            ace_type="group",
            allowed=True,
            propagate=True,
            role_id=role_id,
            group_id=group.id,
            path=f"/pools/{pool_id}",
        )
        _, create_acl_error = create_acl(ctx, create_ace_data)
        if GNS3Error.has_error(create_acl_error):
            GNS3Error.print_error(create_acl_error)
            return False
        click.secho("Success: ", nl=False, fg="green")
        click.secho(f"created the acl for resource {create_ace_data.path}")
    return True


def safe_json(resp):
    if resp.headers.get("Content-Length") == "0" or not resp.text:
        return None
    return resp.json()


def print_usernames_and_ids(ctx: click.Context):
    call_data = call_client_data(package="get", method="users")
    error, users_raw = call_client_method(ctx, call_data.package, call_data.method)
    if GNS3Error.has_error(error):
        GNS3Error.print_error(error)
        return
    users: list[User] = validate_response(call_data.method, users_raw)
    click.secho("List of all users and their id:", fg="green")
    for user in users:
        print_separator_with_secho()
        username = user.username
        user_id = user.user_id
        click.secho("Username: ", fg="cyan", nl=False)
        click.secho(f"{username}")
        click.secho("ID: ", fg="cyan", nl=False)
        click.secho(f"{user_id}")
        print_separator_with_secho()


def get_command_description(
    cmd: str, help_dict: dict, arg_type: str
) -> tuple[str, str]:
    """
    Retrieves the description of a command from the help dictionary.

    Args:
        cmd (str): The command name to look up.
        help_dict (dict): The dictionary containing help metadata.
        arg_type (str): The key in help_dict corresponding to the argument count category
                        (e.g., "zero_arg", "one_arg", etc.).

    Returns:
        tuple[str,str]: The description and example of the command, or an empty string if not found.
    """
    current_help_option = ""
    epiloge = ""

    for key, value in help_dict[arg_type].items():
        if key == cmd:
            current_help_option = str(value["description"])
            epiloge = str(value["example"])
            break

    return current_help_option, epiloge


@click.command(name="install-completion")
@click.option(
    "--shell",
    type=click.Choice(["bash", "zsh", "fish"]),
    required=True,
    help="Shell to act on (install or uninstall completion).",
)
@click.option(
    "--install",
    is_flag=True,
    help="Automatically install completion (POSIX only).",
)
@click.option(
    "--uninstall",
    is_flag=True,
    help="Automatically remove completion (POSIX only).",
)
@click.pass_context
def install_completion(ctx: click.Context, shell, install, uninstall):
    """
    Install or uninstall shell completion for bash, zsh, or fish.

    If neither --install nor --uninstall is given, prints instructions.
    """
    if install and uninstall:
        raise click.UsageError(
            "Error: --install and --uninstall are mutually exclusive."
        )
        ctx.exit(1)

    install_cmds = {
        "bash": [
            "_GNS3UTIL_COMPLETE=bash_source gns3util > ~/.gns3util-complete.bash",
            "echo '. ~/.gns3util-complete.bash' >> ~/.bashrc",
        ],
        "zsh": [
            "_GNS3UTIL_COMPLETE=zsh_source gns3util > ~/.gns3util-complete.zsh",
            "echo '. ~/.gns3util-complete.zsh' >> ~/.zshrc",
        ],
        "fish": [
            "_GNS3UTIL_COMPLETE=fish_source gns3util "
            "> ~/.config/fish/completions/gns3util.fish",
        ],
    }
    uninstall_cmds = {
        "bash": [
            "rm -f ~/.gns3util-complete.bash",
            "sed -i '/gns3util-complete.bash/d' ~/.bashrc",
        ],
        "zsh": [
            "rm -f ~/.gns3util-complete.zsh",
            "sed -i '/gns3util-complete.zsh/d' ~/.zshrc",
        ],
        "fish": [
            "rm -f ~/.config/fish/completions/gns3util.fish",
        ],
    }

    cmds_to_run = None
    action = None

    if install:
        action = "install"
        cmds_to_run = install_cmds[shell]
    elif uninstall:
        action = "uninstall"
        cmds_to_run = uninstall_cmds[shell]

    if action:
        if os.name == "posix":
            os.system("\n".join(cmds_to_run))
            click.secho(f"Completion {action}ed. Please reopen your terminal.")
            return
        else:
            click.secho(
                "Automatic install/uninstall is supported only on POSIX systems."
            )
            return

    click.secho("To install shell completion, run:")
    for c in install_cmds[shell]:
        click.secho(f"  $ {c}", bold=True)
    click.secho("")
    click.secho("To uninstall shell completion, run:")
    for c in uninstall_cmds[shell]:
        click.secho(f"  $ {c}", bold=True)
    click.secho("")
    click.secho("When done, please reopen your terminal.")


def replace_vars(
    input_string: str,
    replacements: list,
    replace_iterations=False,
    iteration_var_name="iteration",
) -> str:
    current_string = input_string
    if replace_iterations:
        pattern = re.compile(r"{{(" + re.escape(iteration_var_name) + r")}}")
        for i, replacement_value in enumerate(replacements):
            m = re.search(pattern, current_string)
            if m:
                current_string = re.sub(
                    pattern, str(replacement_value), current_string, count=1
                )
            else:
                click.secho(
                    f"Warning: No more placeholders found to replace remaining values from index {
                        i
                    }."
                )
                break
        return current_string
    current_string = input_string
    pattern = r"{{(\w*)}}"
    for i, replacement_value in enumerate(replacements):
        m = re.search(pattern, current_string)
        if m:
            current_string = re.sub(
                pattern, str(replacement_value), current_string, count=1
            )
        else:
            click.secho(
                f"Warning: No more placeholders found to replace remaining values from index {
                    i
                }."
            )
            break
    return current_string


def validate_response[T](method: str, data: Any) -> T:
    expected_schema_type = RESPONSE_SCHEMA_MAP.get(method)
    if expected_schema_type is None:
        raise ValueError(
            f"Schema for method '{method}' not found in RESPONSE_SCHEMA_MAP."
        )

    try:
        if expected_schema_type is Any:
            return data

        elif expected_schema_type is type(None):
            if data is not None:
                raise TypeError(f"Expected None for '{method}' but got {type(data)}.")
            return None

        adapter = TypeAdapter(expected_schema_type)
        validated_data = adapter.validate_python(data)
        return validated_data

    except ValidationError as e:
        # redo those prints
        print(f"Validation Error for '{method}' (response):")
        print(e.errors())
        raise e
    except TypeError as e:
        print(f"Type Error for '{method}' (response):")
        raise e
    except Exception as e:
        print(
            f"An unexpected error occurred during validation for '{
                method
            }' (response): {e}"
        )
        raise e


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
    "node": ["node_id", "name"],
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
    "node": "nodes",
}


def resolve_ids(
    ctx: click.Context, subcommand: str, name: str, args: list = []
) -> tuple[str, bool]:
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
    if key == "nodes":
        cd.args = args
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


def get_data_for_update[T](cd: call_client_data, id: str) -> T:
    err, data_raw = call_client_method(cd)
    if GNS3Error.has_error(err):
        GNS3Error.print_error(err)
        cd.ctx.exit(1)
    validated_data = validate_response(cd.method, data_raw)

    key = None

    for map_entry in subcommand_key_map.items():
        if map_entry[1] == cd.method:
            key = map_entry[0]
            break

    for data in validated_data:
        for element in id_element_name.items():
            if element[0] == key:
                id_obj = getattr(data, element[1][0])
                if str(id_obj) == id:
                    return data
    return None


def is_valid_uuid(uuid_to_test: str, version: int = 4) -> bool:
    """
    Check if uuid_to_test is a valid UUID.

     Parameters
    ----------
    uuid_to_test : str
    version : {1, 2, 3, 4}

     Returns
    -------
    `True` if uuid_to_test is a valid UUID, otherwise `False`.
    """

    try:
        uuid_obj = uuid.UUID(uuid_to_test, version=version)
    except ValueError:
        return False
    except TypeError:
        return False
    return str(uuid_obj) == uuid_to_test
