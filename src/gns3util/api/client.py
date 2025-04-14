# Package init file
import requests
import json
import urllib.parse
import threading
from dataclasses import dataclass


@dataclass
class GNS3Error:
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

    @staticmethod
    def has_error(error_instance) -> bool:
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


class GNS3APIClient:
    def __init__(self, server_url, key=None):
        self.server_url = server_url.rstrip('/')
        self.key = key

    def _get_headers(self):
        headers = {'accept': 'application/json'}
        if self.key:
            headers['Authorization'] = f'Bearer {self.key["access_token"]}'
        return headers

    def _handle_request(self, url, headers=None, method="GET", data=None, timeout=10, stream=False) -> (GNS3Error, any):
        error = GNS3Error()
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
                error.other_http_code = True
                error.other_code = response.status_code
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

    def _api_call(self, endpoint, stream=False, method="GET", data=None):
        url = f"{self.server_url}/v3/{endpoint}"
        return self._handle_request(url, headers=self._get_headers(), stream=stream, method=method, data=data)

    def _stream_notifications(self, endpoint, timeout_seconds=60) -> GNS3Error:
        error, response = self._api_call(endpoint, stream=True)
        notification_error = GNS3Error()
        if not GNS3Error.has_error(error):
            def close_stream():
                try:
                    print(f"Closing stream after {timeout_seconds} seconds.")
                    response.close()
                except Exception as e:
                    print(f"Error closing stream: {e}")

            timer = threading.Timer(timeout_seconds, close_stream)
            timer.start()

            try:
                for line in response.iter_lines():
                    if line:
                        decoded_line = line.decode('utf-8')
                        try:
                            notification = json.loads(decoded_line)
                            print(f"Received notification: {notification}")
                        except json.JSONDecodeError:
                            notification_error.json_decode = True
                            notification_error.msg = f"Received non-JSON line: {decoded_line}"
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
