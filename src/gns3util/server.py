from bottle import route, run, response, post, request, get
import time
import multiprocessing
import webbrowser
import importlib.resources

RECEIVED_DATA = None
SHUTDOWN_EVENT = multiprocessing.Event()


class StopServerException(Exception):
    pass


@route("/")
def index():
    # Serve index.html from package resources
    with (
        importlib.resources.files("gns3util.static")
        .joinpath("index.html")
        .open("rb") as f
    ):
        data = f.read()
    response.content_type = "text/html"
    return data


@route("/style.css")
def serve_css():
    with (
        importlib.resources.files("gns3util.static")
        .joinpath("style.css")
        .open("rb") as f
    ):
        data = f.read()
    response.content_type = "text/css"
    return data


@route("/script.js")
def serve_js():
    with (
        importlib.resources.files("gns3util.static")
        .joinpath("script.js")
        .open("rb") as f
    ):
        data = f.read()
    response.content_type = "application/javascript"
    return data


@get("/favicon.ico")
def serve_ico():
    with (
        importlib.resources.files("gns3util.static")
        .joinpath("favicon.ico")
        .open("rb") as f
    ):
        data = f.read()
    response.content_type = "image/x-icon"
    return data


@post("/data")
def process_submission():
    global RECEIVED_DATA
    try:
        json_data = request.json
        if json_data:
            RECEIVED_DATA["data"] = json_data
            SHUTDOWN_EVENT.set()
            return {"status": "success", "received_data": json_data}
        else:
            return {"status": "error", "message": "No JSON data received"}
    except Exception as e:
        print(f"Error processing JSON: {e}")
        return {"status": "error", "message": "Invalid JSON data"}


def run_server(host: str, port: int, debug: bool):
    try:
        run(host=host, port=port, debug=debug)
    finally:
        print("Server stopped gracefully.")


def start_and_get_data(host: str = "localhost", port: int = 8080, debug: bool = True):
    global RECEIVED_DATA
    manager = multiprocessing.Manager()
    RECEIVED_DATA = manager.dict()
    RECEIVED_DATA["data"] = None
    webbrowser.open(f"http://{host}:{port}")
    proc = multiprocessing.Process(target=run_server, args=(host, port, debug))
    proc.start()
    SHUTDOWN_EVENT.wait()
    time.sleep(1)
    proc.terminate()
    proc.join()
    return RECEIVED_DATA["data"]
