import click
import os
from . import auth
from .api.get_endpoints import GNS3GetAPI
from .utils import fuzzy_info_wrapper, get_fuzzy_info_params, fuzzy_params_type, fuzzy_put_wrapper, fuzzy_password_params


def get_client(ctx):
    """Helper function to create GNS3GetAPI instance."""
    server_url = ctx.parent.obj['server']
    verify = ctx.parent.obj['verify']
    success, key = auth.load_and_try_key(ctx)
    if success:
        return GNS3GetAPI(server_url, key['access_token'], verify=verify)
    else:
        os._exit(1)


@click.group()
def fuzzy():
    """Interactive get commands using fzf."""
    pass


@fuzzy.command(name="user-info", help="find user info using fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_user_info(ctx, multi):
    params = get_fuzzy_info_params(
        fuzzy_params_type.user_info, ctx, get_client, multi)
    fuzzy_info_wrapper(params)


@fuzzy.command(name="ui", help="find user info using fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_user_info_command_short(ctx, multi):
    params = get_fuzzy_info_params(
        fuzzy_params_type.user_info, ctx, get_client, multi)
    fuzzy_info_wrapper(params)


@fuzzy.command(name="group-info", help="find group info using fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_group_info(ctx, multi):
    params = get_fuzzy_info_params(
        fuzzy_params_type.group_info, ctx, get_client, multi)
    fuzzy_info_wrapper(params)


@fuzzy.command(name="gi", help="find group info using fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_group_info_command_short(ctx, multi):
    params = get_fuzzy_info_params(
        fuzzy_params_type.group_info, ctx, get_client, multi)
    fuzzy_info_wrapper(params)


@fuzzy.command(name="group-info-with-usernames", help="find group info with members using fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_group_info_with_members(ctx, multi):
    params = get_fuzzy_info_params(
        fuzzy_params_type.group_info_with_usernames, ctx, get_client, multi)
    fuzzy_info_wrapper(params)


@fuzzy.command(name="gim", help="find group info with members using fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_group_info_with_members_command_short(ctx, multi):
    params = get_fuzzy_info_params(
        fuzzy_params_type.group_info_with_usernames, ctx, get_client, multi)
    fuzzy_info_wrapper(params)


@fuzzy.command(name="user-info-and-group-membership", help="find user info and group membership using fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_user_info_and_groups(ctx, multi):
    params = get_fuzzy_info_params(
        fuzzy_params_type.user_info_and_group_membership, ctx, get_client, multi)
    fuzzy_info_wrapper(params)


@fuzzy.command(name="uig", help="find user info and group membership using fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def find_user_info_and_groups_short(ctx, multi):
    params = get_fuzzy_info_params(
        fuzzy_params_type.user_info_and_group_membership, ctx, get_client, multi)
    fuzzy_info_wrapper(params)


@fuzzy.command(name="chpw", help="Find a user and change their passwor dusing fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def change_password_short(ctx, multi):
    params = fuzzy_password_params(
        ctx=ctx,
        client=get_client,
        method="users",
        key="username",
        multi=multi,
    )
    fuzzy_put_wrapper(params)


@fuzzy.command(name="change-password", help="Find a user and change their passwor dusing fzf")
@click.option(
    "-m", "--multi", is_flag=True, help="Enable multi-select mode."
)
@click.pass_context
def change_password(ctx, multi):
    params = fuzzy_password_params(
        ctx=ctx,
        client=get_client,
        method="users",
        key="username",
        multi=multi,
    )
    fuzzy_put_wrapper(params)
