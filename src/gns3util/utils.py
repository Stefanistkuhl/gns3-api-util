import requests
import click
from dataclasses import dataclass
import subprocess
import json
from typing import Callable, Any, Optional

GREY = "\033[90m"
CYAN = "\033[96m"
RESET = "\033[0m"


@dataclass
class fuzzy_error:
    network: bool = False
    missing_permissions: bool = False
    empty_data: bool = False


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
def _handle_request(url, headers=None, method="GET", data=None, timeout=10, stream=False) -> request_error:
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
        tuple: (success, response_data), where success is a boolean and response_data is the JSON response or None.
        If stream is true, returns response object.
    """
    error = request_error()
    try:
        response = requests.request(
            method, url, headers=headers, json=data, timeout=timeout, stream=stream
        )
        if response.status_code == 200:
            if stream:
                return True, response
            else:
                return True, response.json()
        elif response.status_code == 404:
            try:
                error_msg = response.json().get('message', 'Resource not found')
                print(f"Not found: {error_msg}")
            except json.JSONDecodeError:
                print(f"Not found: {response.text}")
            return False, None
        elif response.status_code == 401:
            print("Authentication failed: Unauthorized access.")
            try:
                print(f"Response: {response.json()}")
            except json.JSONDecodeError:
                print(f"Response: {response.text}")
            return False, None
        elif response.status_code == 403:
            try:
                error_msg = response.json().get('message', 'Access forbidden')
                print(f"Access forbidden: {error_msg}")
            except json.JSONDecodeError:
                print(f"Access forbidden: {response.text}")
            return False, None
        elif response.status_code == 422:
            print(f"Validation Error: {response.json()}")
            return False, None
        else:
            print(f"Server returned error: {response.status_code}")
            print(f"Response: {response.text}")
            return False, None
    except requests.exceptions.ConnectionError:
        base_url = url.split('/v3/')[0] if '/v3/' in url else url
        print(f"Connection error: Could not connect to {base_url}")
        return False, None
    except requests.exceptions.Timeout:
        print("Connection timeout: The server took too long to respond.")
        return False, None
    except requests.exceptions.RequestException as e:
        print(f"Request error: {str(e)}")
        return False, None
    except Exception as e:
        print(f"Unexpected error during request: {str(e)}")
        return False, None


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


def fuzzy_info(params=fuzzy_info_params) -> fuzzy_error:
    # maybe break this up into more functions
    # add real error handeling with returning error types
    error = fuzzy_error()
    fzf_input_data, api_data, get_fzf_input_error = get_values_for_fuzzy_input(
        params)
    if get_fzf_input_error.network:
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
                    opt_raw = getattr(params.client(
                        params.ctx), params.opt_method)(a[params.opt_key])
                    if not opt_raw[0]:
                        error.request_network_error = True
                        return error
                    opt_data = opt_raw[1]
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


def get_values_for_fuzzy_input(params) -> (list, list, fuzzy_error):
    error = fuzzy_error()
    raw_data = getattr(params.client(params.ctx), params.method)()
    if not raw_data[0]:
        error.network = True
        return None, None, error
    api_data = raw_data[1]
    fzf_input_data = []
    for data in api_data:
        fzf_input_data.append(data[params.key])
    return fzf_input_data, api_data, fuzzy_error


def fuzzy_put():
    pass


def fuzzy_info_wrapper(params):
    error = fuzzy_info(params)
    if error.network:
        click.echo(
            "Failed to fetch data from the API check your Network connection to the server", err=True)
