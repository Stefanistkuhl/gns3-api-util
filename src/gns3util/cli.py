import click
from .auth import auth
from .get import get, stream
from .fuzzy import fuzzy
from .post import post, controller, compute, project, node, image, snapshot
from .add import add
from .create import create
from .update import update
from .delete import delete


@click.group()
@click.option('--server', '-s', required=True, type=str, help="GNS3 server URL")
@click.option('--insecure', '-i', required=False, is_flag=True, default=True, flag_value=False, help="Ignore unsigned SSL-Certificates")
@click.pass_context
def gns3util(ctx, server, insecure):
    """A utility for GNS3."""
    ctx.ensure_object(dict)
    ctx.obj = {'server': server, 'verify': insecure}


gns3util.add_command(auth)
gns3util.add_command(get)
gns3util.add_command(stream)
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

if __name__ == '__main__':
    gns3util()
