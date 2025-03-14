import click
from .auth import auth


@click.group()
def gns3util():
    """A utility for GNS3."""
    pass


gns3util.add_command(auth)

if __name__ == '__main__':
    gns3util()
