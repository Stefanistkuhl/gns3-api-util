import click
import time
import urllib.parse
import threading
from . import auth
import requests
import json
import os


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
        return _handle_request(url, headers=self._get_headers(), stream=stream, method=method, data=data)

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

    def link_capture_stream(self, project_id, link_id, output_file=None, timeout=None):
        """
        Stream the PCAP capture file from a link.
        Args:
            project_id (str): Project ID
            link_id (str): Link ID
            output_file (str, optional): Path to save the PCAP file. If None, returns the stream data.
            timeout (int, optional): Maximum time in seconds to run the capture. If None, runs until closed.
        Returns:
            tuple: (success, response_data) where success is a boolean and response_data contains
                   the stream data or file path depending on output_file parameter.
        """
        url = f"{self.server_url}/v3/projects/{project_id}/links/{link_id}/capture/stream"

        try:
            response = requests.get(
                url,
                headers=self._get_headers(),
                stream=True,
                timeout=timeout
            )

            if response.status_code == 200:
                if output_file:
                    start_time = time.time()
                    with open(output_file, 'wb') as f:
                        for chunk in response.iter_content(chunk_size=8192):
                            f.write(chunk)
                            # Check if timeout has been reached
                            if timeout and (time.time() - start_time > timeout):
                                print(
                                    f"Capture stopped after {timeout} seconds timeout")
                                break
                    print(f"PCAP file saved to {output_file}")
                    response.close()
                    return True, output_file
                else:
                    # For raw data return, we'll need to accumulate chunks
                    data = bytearray()
                    start_time = time.time()
                    for chunk in response.iter_content(chunk_size=8192):
                        data.extend(chunk)
                        # Check if timeout has been reached
                        if timeout and (time.time() - start_time > timeout):
                            print(
                                f"Capture stopped after {timeout} seconds timeout")
                            break
                    response.close()
                    return True, bytes(data)
            elif response.status_code == 409:
                print(f"Capture error: {response.text}")
                return False, None
            elif response.status_code == 404:
                print(
                    f"Not found: Project or link not found - {response.text}")
                return False, None
            else:
                print(f"Server returned error: {response.status_code}")
                print(f"Response: {response.text}")
                return False, None
        except requests.exceptions.ConnectionError:
            print(f"Connection error: Could not connect to {self.server_url}")
            return False, None
        except requests.exceptions.Timeout:
            print("Connection timeout: The server took too long to respond.")
            return False, None
        except requests.exceptions.RequestException as e:
            print(f"Request error: {str(e)}")
            return False, None
        except Exception as e:
            print(f"Unexpected error during capture stream: {str(e)}")
            return False, None

    # Controller endpoints

    def version(self):
        return self._api_call("version")

    def iou_license(self):
        return self._api_call("iou_license")

    def statistics(self):
        return self._api_call("statistics")

    def notifications(self, timeout_seconds=60):
        self._stream_notifications("notifications", timeout_seconds)

    # User endpoints
    def current_user_info(self):
        return self._api_call("access/users/me")

    def users(self):
        return self._api_call("access/users")

    def user(self, user_id):
        return self._api_call(f"access/users/{user_id}")

    def users_groups(self, user_id):
        return self._api_call(f"access/users/{user_id}/groups")

    # Project endpoints
    def projects(self):
        return self._api_call("projects")

    def project(self, project_id):
        return self._api_call(f"projects/{project_id}")

    def project_stats(self, project_id):
        return self._api_call(f"projects/{project_id}/stats")

    def project_notifications(self, project_id, timeout_seconds=60):
        self._stream_notifications(f"projects/{project_id}/notifications", timeout_seconds)

    def project_locked(self, project_id):
        """Check if a project is locked."""
        return self._api_call(f"projects/{project_id}/locked")

    def download_project_file(self, project_id, file_path):
        encoded_file_path = urllib.parse.quote(file_path)
        response = requests.get(
            f"{self.server_url}/v3/projects/{project_id}/files/{encoded_file_path}",
            headers=self._get_headers(),
            stream=True
        )

        if response.status_code == 200:
            filename = file_path.split("/")[-1]
            with open(filename, 'wb') as f:
                for chunk in response.iter_content(chunk_size=8192):
                    f.write(chunk)
            print(f"File downloaded successfully to {filename}")
        else:
            print(
                f"Failed to download file. Status code: {response.status_code}")
            print(response.text)

    def project_export(self, project_id, export_params):
        """Export a project with the given parameters."""
        params = {
            "include_snapshots": str(export_params["include_snapshots"]).lower(),
            "include_images": str(export_params["include_images"]).lower(),
            "reset_mac_addresses": str(export_params["reset_mac_addresses"]).lower(),
            "keep_compute_ids": str(export_params["keep_compute_ids"]).lower(),
            "compression": export_params["compression"],
            "compression_level": export_params["compression_level"],
        }
        url = f"{self.server_url}/v3/projects/{project_id}/export"
        try:
            response = requests.get(
                url, params=params, headers=self._get_headers(), stream=True)
            response.raise_for_status()
            return response
        except requests.exceptions.RequestException as e:
            print(f"Error during request: {e}")
            return None

    def download_exported_project(self, project_id, export_params):
        """Downloads an exported project and saves it to a file."""
        response = self.project_export(project_id, export_params)

        if response and response.status_code == 200:
            filename = response.headers.get('content-disposition')
            if filename:
                filename = filename.split("filename=")[1].replace('"', '')
            else:
                filename = "exported_project.gns3project"

            with open(filename, 'wb') as f:
                for chunk in response.iter_content(chunk_size=8192):
                    f.write(chunk)
            print(f"Project exported successfully to {filename}")
        else:
            print(
                f"Failed to export project. Status code: {response.status_code if response else 'N/A'}")
            if response:
                print(response.text)

    # Group endpoints
    def groups(self):
        return self._api_call("access/groups")

    def groupsById(self, group_id):
        return self._api_call(f"access/groups/{group_id}")

    def groupMembers(self, group_id):
        return self._api_call(f"access/groups/{group_id}/members")

    # Role endpoints
    def roles(self):
        return self._api_call("access/roles")

    def roleById(self, role_id):
        return self._api_call(f"access/roles/{role_id}")

    def rolePrivileges(self, role_id):
        return self._api_call(f"access/roles/{role_id}/privileges")

    # Privilege endpoints
    def privileges(self):
        return self._api_call("access/privileges")

    # ACL endpoints
    def aclEndpoints(self):
        return self._api_call("access/acl/endpoints")

    def acl(self):
        return self._api_call("access/acl")

    def aclById(self, ace_id):
        return self._api_call(f"access/acl/{ace_id}")

    # Template endpoints
    def templates(self):
        return self._api_call("templates")

    def templateByID(self, template_id):
        return self._api_call(f"templates/{template_id}")

    # Nodes endpoints
    def nodes(self, project_id):
        return self._api_call(f"projects/{project_id}/nodes")

    def nodeByID(self, project_id, node_id):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}")

    def nodeLinksByID(self, project_id, node_id):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/links")

    def nodeDynamipsAutoIdlepc(self, project_id, node_id):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/dynamips/auto_idlepc")

    def nodeDynaimipsIdlecpcProposals(self, project_id, node_id):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/dynamips/idlepc_proposals")

    def nodeGetFile(self, project_id, node_id, file_path):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/files/{file_path}")

    # Link endpoints

    def links(self, project_id):
        return self._api_call(f"projects/{project_id}/links")

    def linkFilters(self, project_id, link_id):
        return self._api_call(f"projects/{project_id}/links/{link_id}/available_filters")

    def link(self, project_id, link_id):
        return self._api_call(f"projects/{project_id}/links/{link_id}")

    # Drawing endpoionts

    def drawings(self, project_id):
        return self._api_call(f"projects/{project_id}/drawings")

    def drawing(self, project_id, drawing_id):
        return self._api_call(f"projects/{project_id}/drawings/{drawing_id}")

    # Symbols endpoints

    def symbols(self):
        return self._api_call(f"symbols")

    # this does not work simply fix later
    def symbol(self, symbol_id):
        encoded_symbol_id = urllib.parse.quote(symbol_id, safe='')
        print(encoded_symbol_id)
        return self._api_call(f"symbols/{encoded_symbol_id}/raw")

    def symbolDimensions(self, symbol_id):
        return self._api_call(f"symbols/{symbol_id}/dimensions")

    def defaultSymbols(self):
        return self._api_call(f"symbols/default_symbols")

    # Snapshot endpoints

    def snapshots(self, project_id):
        return self._api_call(f"projects/{project_id}/snapshots")

    # Computes endpoints

    def computes(self):
        return self._api_call(f"computes")

    def computeByID(self, compute_id):
        return self._api_call(f"computes/{compute_id}")

    def computeByIDDockerImages(self, compute_id):
        return self._api_call(f"computes/{compute_id}/docker/images")

    def computeByIDVirtualvoxVms(self, compute_id):
        return self._api_call(f"computes/{compute_id}/virtualbox/vms")

    def computeByIDVmwareVms(self, compute_id):
        return self._api_call(f"computes/{compute_id}/vmware/vms")

    # Applicances endpoints

    def appliances(self):
        return self._api_call(f"appliances")

    def appliance(self, appliance_id):
        return self._api_call(f"appliances/{appliance_id}")

    # Ressource pools endoint

    def pools(self):
        return self._api_call(f"pools")

    def pool(self, resource_pool_id):
        return self._api_call(f"pools/{resource_pool_id}")

    def poolResources(self, resource_pool_id):
        return self._api_call(f"pools/{resource_pool_id}/resources")

get = click.Group('get')

def get_client(ctx):
    """Helper function to create GNS3APIClient instance."""
    key_file = os.path.expanduser("~/.gns3key")
    # Get server from parent context
    server_url = ctx.parent.obj['server']
    key = auth.loadKey(key_file)
    return GNS3APIClient(server_url, key)

# Add helper to reduce duplication
def execute_and_print(ctx, func):
    client = get_client(ctx)
    success, data = func(client)
    if success:
        print(json.dumps(data, indent=2))

# CLI command registrations

# Commands with no arguments
_zero_arg = {
    "version": "version",
    "iou-license": "iou_license",
    "statistics": "statistics",
    "me": "current_user_info",
    "users": "users",
    "projects": "projects",
    "groups": "groups",
    "roles": "roles",
    "privileges": "privileges",
    "acl-endpoints": "aclEndpoints",
    "acl": "acl",
    "templates": "templates",
    "symbols": "symbols",
    "default-symbols": "defaultSymbols",
    "computes": "computes",
    "appliances": "appliances",
    "pools": "pools"
}
# Create click commands with zero arguments
for cmd, func in _zero_arg.items():
    def make_cmd(func=func):
        @click.pass_context
        def cmd_func(ctx):
            execute_and_print(ctx, lambda client: getattr(client, func)())
        return cmd_func
    get.command(name=cmd)(make_cmd())

# Commands with one argument
_one_arg = {
    "user": "user",
    "user-groups": "users_groups",
    "project": "project",
    "project-stats": "project_stats",
    "project-locked": "project_locked",
    "group": "groupsById",
    "group-members": "groupMembers",
    "role": "roleById",
    "role-privileges": "rolePrivileges",
    "template": "templateByID",
    "compute": "computeByID",
    "docker-images": "computeByIDDockerImages",
    "virtualbox-vms": "computeByIDVirtualvoxVms",
    "vmware-vms": "computeByIDVmwareVms",
    "appliance": "appliance",
    "pool": "pool",
    "pool-resources": "poolResources",
    "drawings": "drawings",
    "symbol": "symbol",
    "acl-rule": "aclById",
    "links": "links",
    "nodes": "nodes"
}
# Create click commands with one argument
for cmd, func in _one_arg.items():
    def make_cmd(func=func):
        @click.argument('arg')
        @click.pass_context
        def cmd_func(ctx, arg):
            execute_and_print(ctx, lambda client: getattr(client, func)(arg))
        return cmd_func
    get.command(name=cmd)(make_cmd())

# Commands with two arguments (assumed: project_id and id)
_two_arg = {
    "node": "nodeByID",
    "node-links": "nodeLinksByID",
    "link": "link",
    "link-filters": "linkFilters",
    "drawing": "drawing"
}
# Create click commands with two arguments
for cmd, func in _two_arg.items():
    def make_cmd(func=func):
        @click.argument('project_id')
        @click.argument('id')
        @click.pass_context
        def cmd_func(ctx, project_id, id):
            execute_and_print(ctx, lambda client: getattr(client, func)(project_id, id))
        return cmd_func
    get.command(name=cmd)(make_cmd())

# Special commands with timeout options
@get.command()
@click.option('--timeout', '-t', 'timeout_seconds', default=60, help='Notification stream timeout in seconds')
@click.pass_context
def notifications(ctx, timeout_seconds):
    get_client(ctx).notifications(timeout_seconds)

@get.command()
@click.argument('project_id')
@click.option('--timeout', '-t', 'timeout_seconds', default=60, help='Notification stream timeout in seconds')
@click.pass_context
def project_notifications(ctx, project_id, timeout_seconds):
    get_client(ctx).project_notifications(project_id, timeout_seconds)