import click
from .auth import auth
from .get import get

@click.group()
def gns3util():
    """A utility for GNS3."""
    pass

gns3util.add_command(auth)
gns3util.add_command(get)

if __name__ == '__main__':
    gns3util()
