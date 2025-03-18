# Package init file
import time
import urllib.parse
import threading
import requests
import json
from .. import utils


class GNS3APIClient:
    def __init__(self, server_url, key=None):
        self.server_url = server_url.rstrip('/')
        self.key = key

    def _get_headers(self):
        headers = {'accept': 'application/json'}
        if self.key:
            headers['Authorization'] = f'Bearer {self.key["access_token"]}'
        return headers

    def _api_call(self, endpoint, stream=False, method="GET", data=None):
        url = f"{self.server_url}/v3/{endpoint}"
        return utils._handle_request(url, headers=self._get_headers(), stream=stream, method=method, data=data)

    def _stream_notifications(self, endpoint, timeout_seconds=60):
        success, response = self._api_call(endpoint, stream=True)

        if success:
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
                            print(f"Received non-JSON line: {decoded_line}")
            except requests.exceptions.ChunkedEncodingError:
                print("Stream ended unexpectedly.")
            except Exception as e:
                if not isinstance(e, AttributeError) or "NoneType" not in str(e):
                    print(f"Error processing stream: {e}")
            finally:
                if timer.is_alive():
                    timer.cancel()
                if not response.raw.closed:
                    response.close()
        else:
            print("Failed to start notifications stream.")


__all__ = ['GNS3APIClient']
