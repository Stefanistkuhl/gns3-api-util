import requests
import rich
import click
from dataclasses import dataclass, field
import subprocess
import json
from typing import Callable, Any, Optional

GREY = "\033[90m"
CYAN = "\033[96m"
RESET = "\033[0m"


@dataclass
class request_error:
    not_found: bool = False
    unauthorized: bool = False
    forbidden: bool = False
    validation: bool = False
    other_http_code: bool = False
    json_decode: bool = False
    connection: bool = False
    timeout: bool = False
    request: bool = False
    unexpected: bool = False
    encoding: bool = False
    start: bool = False
    empty_data: bool = False
    other_code: int = 0
    msg: str = ""


@dataclass
class fuzzy_info_params:
    ctx: Any
    client: Callable[[Any], Any]
    opt_method: Optional[str] = None
    opt_key: Optional[str] = None
    method: str = "str"
    key: str = "str"
    multi: bool = False
    opt_data: bool = False


# change this to reutnr a error type to have more flexibility with printing
def _handle_request(url, headers=None, method="GET", data=None, timeout=10, stream=False) -> (request_error, Any):
    """
    Handles HTTP requests with standardized error handling and response processing.

    Args:
        url (str): The URL to make the request to.
        headers (dict, optional): HTTP headers. Defaults to None.
        method (str, optional): HTTP method (GET, POST, etc.). Defaults to 'GET'.
        data (dict, optional): Request payload. Defaults to None.
        timeout (int, optional): Request timeout in seconds. Defaults to 10.
        stream (bool, optional): whether to stream the response. Defaults to False.

    Returns:
        tuple: (error, response_data), where success is a boolean and response_data is the JSON response or None.
        If stream is true, returns response object.
    """
    error = request_error()
    try:
        response = requests.request(
            method, url, headers=headers, json=data, timeout=timeout, stream=stream
        )
        if response.status_code == 200:
            if stream:
                return error, response
            else:
                return error, response.json()
        elif response.status_code == 404:
            error.not_found = True
            try:
                error.msg = response.json().get('message', 'Resource not found')
                return error, None
            except json.JSONDecodeError:
                error.msg = response.text
                error.json_decode = True
                return error, None
            return False, None
        elif response.status_code == 401:
            error.unauthorized = True
            try:
                error.msg = response.json()
            except json.JSONDecodeError:
                error.json_decode = True
                error.msg = response.text
            return error, None
        elif response.status_code == 403:
            error.forbidden = True
            try:
                error_msg = response.json().get('message', 'Access forbidden')
                error.msg = error_msg
            except json.JSONDecodeError:
                error.json_decode = True
                error.msg = response.text
            return error, None
        elif response.status_code == 422:
            error.validation = True
            error.msg = response.json()
            return error, None
        else:
            error.other = True
            error.other_http_code = response.status_code
            error.msg = response.text
            return error, None
    except requests.exceptions.ConnectionError:
        error.connection = True
        base_url = url.split('/v3/')[0] if '/v3/' in url else url
        error.msg = f"Connection error: Could not connect to {base_url}"
        return error, None
    except requests.exceptions.Timeout:
        error.timeout = True
        error.msg = "Connection timeout: The server took too long to respond."
        return error, None
    except requests.exceptions.RequestException as e:
        error.request = True
        error.msg = str(e)
        return error, None
    except Exception as e:
        error.unexpected = True
        error.msg = str(e)
        return error, None


def fzf_select(options, multi=False):
    """
    Opens an fzf window with the given options and returns the selected option(s).

    Args:
        options: A list of strings representing the options to choose from.

    Returns:
        A list of strings containing the selected option(s), or an empty list if none were selected or if fzf is not found.
    """
    try:
        if multi:
            fzf_process = subprocess.Popen(
                ['fzf', '--multi'],
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )
        else:
            fzf_process = subprocess.Popen(
                ['fzf'],
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )

        output, error = fzf_process.communicate('\n'.join(options))

        if error:
            if "fzf: command not found" in error:
                print(
                    "Error: fzf is not installed. Please install it to use this feature.")
                return []
            else:
                print(f"Error running fzf: {error}")
                return []

        if output:
            return [line.strip() for line in output.strip().split('\n')]
        else:
            return []

    except FileNotFoundError:
        print("Error: fzf executable not found in PATH. Please ensure it's installed and accessible.")
        return []


def fuzzy_info(params=fuzzy_info_params) -> request_error:
    # maybe break this up into more functions
    # add real error handeling with returning error types
    error = request_error()
    fzf_input_data, api_data, get_fzf_input_error = get_values_for_fuzzy_input(
        params)
    if has_error(get_fzf_input_error):
        return get_fzf_input_error
    selected = fzf_select(fzf_input_data, multi=params.multi)
    matched = set()
    for selected_item in selected:
        for a in api_data:
            if a[params.key] == selected_item and a[params.key] not in matched:
                for k, v in a.items():
                    print(f"{CYAN}{k}{RESET}: {v}")
                print(f"{GREY}---{RESET}")
                if params.opt_data:
                    opt_data_error, opt_data = getattr(params.client(
                        params.ctx), params.opt_method)(a[params.opt_key])
                    if has_error(opt_data_error):
                        error.request_network_error = True
                        return error
                    if opt_data == []:
                        # either add callbacks for this in the future or print
                        # something better or use ifs to detierme it
                        print(f"Empty data returned from method {
                            params.opt_method} for the {a[params.key]} value")
                    else:
                        for d in opt_data:
                            print(f"{GREY}---{RESET}")
                            for k2, v2 in d.items():
                                print(f"{CYAN}{k2}{RESET}: {v2}")
                            print(f"{GREY}---{RESET}")
                break
    return error


def get_values_for_fuzzy_input(params) -> (list, list, request_error):
    fuzzy_error = request_error()
    get_data_error, api_data = getattr(
        params.client(params.ctx), params.method)()
    if has_error(get_data_error):
        fuzzy_error.connection = True
        return None, None, fuzzy_error
    fzf_input_data = []
    for data in api_data:
        fzf_input_data.append(data[params.key])
    return fzf_input_data, api_data, fuzzy_error


def fuzzy_put():
    pass


def fuzzy_info_wrapper(params):
    error = fuzzy_info(params)
    if error.connection:
        click.echo(
            "Failed to fetch data from the API check your Network connection to the server", err=True)


def has_error(error_instance: request_error) -> bool:
    return any(getattr(error_instance, field) for field in [
        'not_found',
        'unauthorized',
        'forbidden',
        'validation',
        'other_http_code',
        'json_decode',
        'connection',
        'timeout',
        'request',
        'encoding',
        'empty_data',
        'start',
        'unexpected'
    ])


def execute_and_print(ctx, client, func):
    client = client(ctx)
    error, data = func(client)
    if not has_error(error):
        rich.print_json(json.dumps(data, indent=2))
