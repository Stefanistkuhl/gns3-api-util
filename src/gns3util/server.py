from bottle import route, run, static_file
from typing import Optional
import os

STATIC_DIR = os.path.join(os.getcwd(), "src", "gns3util", "static")


@route('/')
def index():
    return static_file("index.html", root=STATIC_DIR)


def start_server(host: str = 'localhost', port: int = 8080, debug: bool = True) -> None:
    run(host=host, port=port, debug=debug)
