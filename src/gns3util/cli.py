import click
from .auth import auth
from .get import get, stream, export
from .fuzzy import fuzzy
from .post import post, controller, compute, project, node, image, snapshot
from .add import add
from .create import create
from .update import update
from .delete import delete
from .ssl import install
from . scripts import script
from . utils import install_completion


@click.group()
@click.option('--server', '-s', envvar="GNS3_SERVER", required=True, type=str, help="GNS3 server URL")
@click.option('--insecure', '-i', required=False, is_flag=True, default=True, flag_value=False, help="Ignore unsigned SSL-Certificates")
@click.option('--key_file', '-k', required=False, type=click.Path(exists=True, readable=True), help="Set a location for a keyfile to use")
@click.pass_context
def gns3util(ctx: click.Context, server, insecure, key_file):
    """A utility for GNS3."""
    ctx.ensure_object(dict)
    ctx.obj = {'server': server, 'verify': insecure, 'key_file': key_file}


gns3util.add_command(auth)
gns3util.add_command(get)
gns3util.add_command(stream)
gns3util.add_command(export)
gns3util.add_command(fuzzy)
gns3util.add_command(post)
gns3util.add_command(controller)
gns3util.add_command(compute)
gns3util.add_command(project)
gns3util.add_command(node)
gns3util.add_command(image)
gns3util.add_command(snapshot)
gns3util.add_command(create)
gns3util.add_command(update)
gns3util.add_command(add)
gns3util.add_command(delete)
gns3util.add_command(install)
gns3util.add_command(script)
gns3util.add_command(install_completion)


if __name__ == '__main__':
    gns3util()
