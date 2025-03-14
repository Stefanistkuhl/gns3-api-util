import click
import urllib.parse
import threading
import auth
import requests
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
        elif response.status_code == 401:
            print("Authentication failed: Unauthorized access.")
            try:
                print(f"Response: {response.json()}")
            except json.JSONDecodeError:
                print(f"Response: {response.text}")
            return False, None
        elif response.status_code == 422:
            print(f"Validation Error: {response.json()}")
            return False, None
        else:
            print(f"Server returned error: {response.status_code}")
            print(f"Response: {response.text}")
            return False, None
    except requests.exceptions.ConnectionError:
        print(
            f"Connection error: Could not connect to {
                url.split('/v3/')[0] if '/v3/' in url else url
            }"
        )
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


""" Contoller """


def version(server_url):
    url = f"{server_url}/v3/version"
    headers = {
        "accept": "application/json",
    }
    return _handle_request(url, headers=headers)


def iou_license(key, server_url):
    url = f"{server_url}/v3/iou_license"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


def statistics(key, server_url):
    url = f"{server_url}/v3/statistics"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


def notifications(key, server_url, timeout_seconds=60):
    """
    Handles the notifications stream from the server with a timeout.

    Args:
        key (dict): Authentication key.
        server_url (str): The base URL of the server.
        timeout_seconds (int, optional): Timeout in seconds. Defaults to 60.
    """
    url = f"{server_url}/v3/notifications"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    success, response = _handle_request(url, headers=headers, stream=True)

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


""" Users """


def currentUserInfo(key, server_url):
    url = f"{server_url}/v3/access/users/me"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


def users(key, server_url):
    url = f"{server_url}/v3/access/users"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


def user(key, server_url, user_id):
    url = f"{server_url}/v3/access/users/{user_id}"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


def usersGroups(key, server_url, user_id):
    url = f"{server_url}/v3/access/users/{user_id}/groups"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


""" Users groups """


def groups(key, server_url):
    url = f"{server_url}/v3/access/groups"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


def groupsById(key, server_url, group_id):
    url = f"{server_url}/v3/access/groups/{group_id}"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


def groupMembers(key, server_url, group_id):
    url = f"{server_url}/v3/access/groups/{group_id}/members"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


""" Roles """


def roles(key, server_url):
    url = f"{server_url}/v3/access/roles"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


def roleById(key, server_url, role_id):
    url = f"{server_url}/v3/access/roles/{role_id}"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


def rolePriviledges(key, server_url, role_id):
    url = f"{server_url}/v3/access/roles/{role_id}/privileges"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


""" Privileges """


def priviledges(key, server_url):
    url = f"{server_url}/v3/access/privileges"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


""" ACLS """


def aclEndpoints(key, server_url):
    url = f"{server_url}/v3/access/acl/endpoints"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


def acl(key, server_url):
    url = f"{server_url}/v3/access/acl"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


def aclById(key, server_url, ace_id):
    url = f"{server_url}/v3/access/acl/{ace_id}"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


"""Templates"""


def templates(key, server_url):
    url = f"{server_url}/v3/templates"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


def templateByID(key, server_url, template_id):
    url = f"{server_url}/v3/templates/{template_id}"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


""" Projects """


def projects(key, server_url):
    url = f"{server_url}/v3/projects"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


def projectByID(key, server_url, project_id):
    url = f"{server_url}/v3/projects/{project_id}"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


def projectStatsByID(key, server_url, project_id):
    url = f"{server_url}/v3/projects/{project_id}/stats"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


def projectNotificationsByID(key, server_url, project_id, timeout_seconds=60):
    url = f"{server_url}/v3/projects/{project_id}/notifications"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    success, response = _handle_request(url, headers=headers, stream=True)

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


def projectLockedByID(key, server_url, project_id):
    url = f"{server_url}/v3/projects/{project_id}/locked"
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    return _handle_request(url, headers=headers)


def projectExportByID(key, server_url, project_id, include_snapshots, include_images, reset_mac_addresses, keep_compute_ids, compression, compression_level):
    # ?include_snapshots=false&include_images=false&reset_mac_addresses=false&keep_compute_ids=false&compression=zstd&compression_level=3
    """
    Avalaible Compression types:
    zip,
    bzip2,
    lzma,
    zstd
    """
    url = f"{server_url}/v3/projects/{project_id}/export"

    params = {
        "include_snapshots": str(include_snapshots).lower(),
        "include_images": str(include_images).lower(),
        "reset_mac_addresses": str(reset_mac_addresses).lower(),
        "keep_compute_ids": str(keep_compute_ids).lower(),
        "compression": compression,
        "compression_level": compression_level,
    }
    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }
    try:
        response = requests.get(
            url, params=params, headers=headers, stream=True)
        response.raise_for_status()
        return response
    except requests.exceptions.RequestException as e:
        print(f"Error during request: {e}")
        return None


def download_exported_project(key, server_url, project_id, export_params):
    """
    Downloads the exported project and handles the response.

    Args:
        key (dict): A dictionary containing the 'access_token'.
        server_url (str): The base URL of the server.
        project_id (str): The UUID of the project.
        export_params (dict): A dictionary containing the export parameters.
    """

    response = projectExportByID(key, server_url, project_id, **export_params)

    if response:
        if response.status_code == 200:
            filename = response.headers.get('content-disposition')
            if filename:
                filename = filename.split("filename=")[1].replace('"', '')
            else:
                # default name if content-disposition is missing.
                filename = "exported_project.gns3project"

            with open(filename, 'wb') as f:
                # 8kb chunks
                for chunk in response.iter_content(chunk_size=8192):
                    f.write(chunk)
            print(f"Project exported successfully to {filename}")

        else:
            print(f"Failed to export project. Status code: {
                  response.status_code}")
            print(response.text)


def ProjectFile(key, server_url, project_id, file_path):
    """
    Returns a file from a project.

    Args:
        key (dict): A dictionary containing the 'access_token'.
        server_url (str): The base URL of the server (e.g., "http://10.21.34.222:3080").
        project_id (str): The UUID of the project.
        file_path (str): The path to the file within the project.

    Returns:
        requests.Response: The response from the server.
    """

    # URL encode the file path to handle special characters
    encoded_file_path = urllib.parse.quote(file_path)

    url = f"{server_url}/v3/projects/{project_id}/files/{encoded_file_path}"

    access_token = key["access_token"]
    headers = {
        'accept': 'application/json',
        'Authorization': f'Bearer {access_token}'
    }

    try:
        response = requests.get(url, headers=headers, stream=True)
        response.raise_for_status()
        return response
    except requests.exceptions.RequestException as e:
        print(f"Error during request: {e}")
        return None


def download_project_file(key, server_url, project_id, file_path):
    """
    Downloads a file from a project and handles the response.

    Args:
        key (dict): A dictionary containing the 'access_token'.
        server_url (str): The base URL of the server (e.g., "http://10.21.34.222:3080").
        project_id (str): The UUID of the project.
        file_path (str): The path to the file within the project.
    """

    response = ProjectFile(key, server_url, project_id, file_path)

    if response:
        if response.status_code == 200:
            # Extract filename from Content-Disposition header (if available)
            # default to the passed in filename.
            filename = file_path.split("/")[-1]
            content_disposition = response.headers.get('content-disposition')
            if content_disposition:
                filename = content_disposition.split(
                    "filename=")[1].replace('"', '')

            with open(filename, 'wb') as f:
                for chunk in response.iter_content(chunk_size=8192):
                    f.write(chunk)
            print(f"File downloaded successfully to {filename}")
        else:
            print(f"Failed to download file. Status code: {
                  response.status_code}")
            print(response.text)


url = "http://10.21.34.222:3080"
key = auth.loadKey("/home/stefiii/.gns3key")
userID = "49570cf3-e9a7-4bc2-a1cc-3738937bf232"
groupID = "3e671ebb-3f53-4548-9c34-cf30619544bc"
roleID = "db9c15bf-871d-4526-ba22-2a7e5ea8c5af"
templateID = "e231aa90-d1cf-421b-8b3b-2358ff9d7cc2"
projectID = "c7acc43b-63fe-4d7c-a149-2e687cb73098"
# projectID = "c4d51906-70cc-41f7-9a8a-9602a5ed8577"
export_params = {
    "include_snapshots": False,
    "include_images": False,
    "reset_mac_addresses": False,
    "keep_compute_ids": False,
    "compression": "zstd",
    "compression_level": 3
}
filePath = "12-FUS-TRJ-Übung-6-DHCPv4.gns3"
# aceID = "no acls set yet so untested"
print("-------Controller-------")
print(version(url))
print(iou_license(key, url))
print(statistics(key, url))
# notifications(key, url, timeout_seconds=3)
print("-------Users-------")
print(currentUserInfo(key, url))
print(users(key, url))
print(user(key, url, userID))
print(usersGroups(key, url, userID))
print("-------Users Groups-------")
print(groups(key, url))
print(groupsById(key, url, groupID))
print(groupMembers(key, url, groupID))
print("-------Roles-------")
# print(roles(key, url))
# print(roleById(key, url, roleID))
# print(roleById(key, url, roleID))
print("too much output so commeted out lol")
print("-------Priviledges-------")
# print(priviledges(key, url))
print("too much output so commeted out lol")
print("-------ACLS-------")
# print(aclEndpoints(key, url))
print("too much output so commeted out lol")
print(acl(key, url))
# print(aclById(key, url, aceID))
print("-------Templates-------")
# print(templates(key, url))
print("too much output so commeted out lol")
print(templateByID(key, url, templateID))
print("-------Projects-------")
print(projects(key, url))
print(projectByID(key, url, projectID))
print(projectLockedByID(key, url, projectID))
download_project_file(key, url, projectID, filePath)
download_exported_project(key, url, projectID, export_params)
