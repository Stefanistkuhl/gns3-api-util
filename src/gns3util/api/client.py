import requests
import click
import urllib3
import json
import threading
from dataclasses import dataclass
from typing import Any


@dataclass
class GNS3Error:
    bad_request: bool = False
    not_found: bool = False
    unauthorized: bool = False
    forbidden: bool = False
    conflict: bool = False
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
    resource: str = ""

    @staticmethod
    def has_error(error_instance) -> bool:
        return any(
            getattr(error_instance, field)
            for field in [
                "bad_request",
                "conflict",
                "not_found",
                "unauthorized",
                "forbidden",
                "validation",
                "other_http_code",
                "json_decode",
                "connection",
                "timeout",
                "request",
                "encoding",
                "empty_data",
                "start",
                "unexpected",
            ]
        )

    @staticmethod
    def print_error(error_instance, *args):
        error_types = {
            "bad_request": "Bad Request Error",
            "conflict": "Conflict Error",
            "not_found": "Not Found Error",
            "unauthorized": "Unauthorized Error",
            "forbidden": "Forbidden Error",
            "validation": "Validation Error",
            "other_http_code": "Other HTTP Code Error",
            "json_decode": "JSON Decode Error",
            "connection": "Connection Error",
            "timeout": "Timeout Error",
            "request": "Request Error",
            "encoding": "Encoding Error",
            "empty_data": "Empty Data Error",
            "start": "Start Error",
            "unexpected": "Unexpected Error",
        }
        errors = [
            msg
            for fld, msg in error_types.items()
            if getattr(error_instance, fld, False)
        ]

        display_msg = error_instance.msg
        if isinstance(display_msg, str):
            try:
                payload = json.loads(display_msg)
                if isinstance(payload, dict) and "message" in payload:
                    display_msg = payload["message"]
            except json.JSONDecodeError:
                pass

        if error_instance.not_found:
            click.secho("Error: ", fg="red", err=True, nl=False)
            click.secho(
                "resource not found error: The following resources were not found: ",
                err=True,
                nl=False,
            )
            if error_instance.resource:
                click.secho(f"- {error_instance.resource}", bold=True, err=True)
            elif args:
                for resource in args:
                    click.secho(f"- {resource}", bold=True, err=True)
            else:
                if display_msg:
                    click.secho(display_msg, bold=True, err=True)
        else:
            if errors:
                click.secho(", ".join(errors) + ": ", fg="red", nl=False, err=True)
                click.secho(display_msg, bold=True, err=True)


class GNS3APIClient:
    def __init__(self, server_url, key=None, verify=True):
        self.server_url = server_url.rstrip("/")
        self.key = key
        self.verify = verify
        if not verify:
            urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

    def _get_headers(self):
        headers = {"accept": "application/json"}
        if self.key:
            headers["Authorization"] = f"Bearer {self.key}"
        return headers

    def _handle_request(
        self,
        url,
        headers=None,
        method="GET",
        data=None,
        timeout=10,
        stream=False,
        verify=True,
        params={},
    ) -> tuple[GNS3Error, Any]:
        error = GNS3Error()
        try:
            response = requests.request(
                method,
                url,
                headers=headers,
                json=data,
                timeout=timeout,
                stream=stream,
                verify=verify,
                params=params,
            )
            if response.status_code in (200, 201, 204):
                if stream:
                    return error, response
                else:
                    from .. import utils

                    return error, utils.safe_json(response)
            elif response.status_code == 400:
                error.bad_request = True
                try:
                    error.msg = response.json().get("message")
                except json.JSONDecodeError:
                    error.msg = response.text
                    error.json_decode = True
                return error, None
            elif response.status_code == 404:
                error.not_found = True
                try:
                    from urllib.parse import urlparse

                    path = urlparse(url).path
                    if path.endswith("/"):
                        path = path[:-1]
                    resource = path.split("/")[-1]
                    error.resource = resource
                except Exception:
                    error.resource = ""
                try:
                    error.msg = response.json().get("message")
                except json.JSONDecodeError:
                    error.msg = response.text
                    error.json_decode = True
                return error, None
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
                    error_msg = response.json().get("message")
                    error.msg = error_msg
                except json.JSONDecodeError:
                    error.json_decode = True
                    error.msg = response.text
                return error, None
            elif response.status_code == 409:
                error.conflict = True
                try:
                    error_msg = response.json().get("message")
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
                error.other_http_code = True
                error.other_code = response.status_code
                error.msg = response.text
                return error, None
        except requests.ConnectionError:
            error.connection = True
            base_url = url.split("/v3/")[0] if "/v3/" in url else url
            error.msg = f"Connection error: Could not connect to {base_url}"
            return error, None
        except requests.Timeout:
            error.timeout = True
            error.msg = "Connection timeout: The server took too long to respond."
            return error, None
        except requests.RequestException as e:
            error.request = True
            error.msg = str(e)
            return error, None
        except Exception as e:
            error.unexpected = True
            error.msg = str(e)
            return error, None

    def _api_call(
        self, endpoint, stream=False, method="GET", data=None, verify=True, params={}
    ) -> tuple[GNS3Error, Any]:
        url = f"{self.server_url}/v3/{endpoint}"
        return self._handle_request(
            url,
            headers=self._get_headers(),
            stream=stream,
            method=method,
            data=data,
            verify=verify,
            params=params,
        )

    def _stream_notifications(
        self, endpoint, timeout_seconds=60, verify=True
    ) -> GNS3Error:
        error, response = self._api_call(endpoint, stream=True, verify=verify)
        notification_error = GNS3Error()
        if not GNS3Error.has_error(error):

            def close_stream():
                try:
                    click.secho(f"Closing stream after {timeout_seconds} seconds.")
                    response.close()
                except Exception as e:
                    click.secho(f"Error closing stream: {e}")

            timer = threading.Timer(timeout_seconds, close_stream)
            timer.start()

            try:
                for line in response.iter_lines():
                    if line:
                        decoded_line = line.decode("utf-8")
                        try:
                            notification = json.loads(decoded_line)
                            click.secho(f"Received notification: {notification}")
                        except json.JSONDecodeError:
                            notification_error.json_decode = True
                            notification_error.msg = (
                                f"Received non-JSON line: {decoded_line}"
                            )
                            return notification_error
            except requests.exceptions.ChunkedEncodingError:
                notification_error.encoding = True
                notification_error.msg = "Stream ended unexpectedly."
                return notification_error
            except Exception as e:
                notification_error.unexpected = True
                if not isinstance(e, AttributeError) or "NoneType" not in str(e):
                    notification_error.msg = f"Error processing stream: {e}"
                    return notification_error
                return notification_error
            finally:
                if timer.is_alive():
                    timer.cancel()
                if not response.raw.closed:
                    response.close()
                return notification_error
        else:
            notification_error.start = True
            notification_error.msg = "Failed to start notifications stream."
            return notification_error
