import requests
import subprocess
import json


def _handle_request(url, headers=None, method="GET", data=None, timeout=10, stream=False):
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


def fzf_select(options):
    """
    Opens an fzf window with the given options and returns the selected option(s).

    Args:
        options: A list of strings representing the options to choose from.

    Returns:
        A list of strings containing the selected option(s), or an empty list if none were selected or if fzf is not found.
    """
    try:
        fzf_process = subprocess.Popen(
            ['fzf', '--multi'],
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
