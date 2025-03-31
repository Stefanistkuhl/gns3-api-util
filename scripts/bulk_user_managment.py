import subprocess
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
import re
import uuid
import json
import shlex
import sys
from typing import List, Optional, Dict, Any, Tuple

# --- Configuration ---
DEFAULT_GNS3_SERVER_URL = "http://10.21.34.224:3080"  # dont keep this for commit
DEFAULT_GNS3UTIL_PATH = 'gns3util'
DEFAULT_MAX_DELETE_WORKERS = 20
PROTECTED_USERNAMES = {"admin"}

# --- Core Utility ---


def run_gns3util_command(
    server_url: str,
    subcommand: str,
    action: str,
    args: Optional[List[str]] = None,
    data: Optional[Dict[str, Any]] = None,
    gns3util_path: str = DEFAULT_GNS3UTIL_PATH
) -> Tuple[bool, str, str]:
    """
    Builds and executes a gns3util command-line instruction using subprocess.

    Args:
        server_url: The GNS3 server URL (passed to -s).
        subcommand: The main subcommand ('get', 'post', 'put', 'delete', 'auth').
        action: The specific action within the subcommand.
        args: Optional list of positional string arguments after the action.
        data: Optional Python dictionary converted to JSON, appended last.
        gns3util_path: Path to the gns3util executable.

    Returns:
        Tuple (success: bool, stdout: str, stderr: str).
    """
    if args is None:
        args = []

    command_list = [
        gns3util_path,
        '--server', server_url,
        subcommand,
        action
    ]
    command_list.extend(args)

    json_data_string = None
    if data is not None:
        try:
            json_data_string = json.dumps(data)
            command_list.append(json_data_string)
        except TypeError as e:
            error_msg = f"Error serializing 'data' to JSON: {e}"
            print(error_msg, file=sys.stderr)
            return False, "", error_msg

    debug_command_str = shlex.join(command_list)
    # print(f"DEBUG: Running command: {debug_command_str}", file=sys.stderr) # Uncomment for debug

    try:
        result = subprocess.run(
            command_list,
            capture_output=True,
            text=True,
            check=False  # Don't raise exception on non-zero exit
        )
        success = (result.returncode == 0)
        if not success:
            print(f"Warning: gns3util command exited with code {result.returncode}. "
                  f"Command: '{debug_command_str}'. Stderr: {result.stderr.strip()}", file=sys.stderr)
        return success, result.stdout.strip(), result.stderr.strip()

    except FileNotFoundError:
        error_msg = f"Error: '{
            gns3util_path}' not found. Ensure gns3util is installed and in PATH."
        print(error_msg, file=sys.stderr)
        return False, "", error_msg
    except Exception as e:
        error_msg = f"Error running subprocess command '{
            debug_command_str}': {e}"
        print(error_msg, file=sys.stderr)
        return False, "", error_msg

# --- User/Group Management Functions ---


def get_user_id_map(
    server_url: str,
    gns3util_path: str = DEFAULT_GNS3UTIL_PATH
) -> Optional[Dict[str, str]]:
    """
    Fetches all users from GNS3 and returns a dictionary mapping username to user_id.

    Args:
        server_url: The GNS3 server URL.
        gns3util_path: Path to the gns3util executable.

    Returns:
        A dictionary {username: user_id} or None on failure.
    """
    print(f"\nFetching users from {server_url}...")
    success, stdout, stderr = run_gns3util_command(
        server_url=server_url,
        subcommand='get',
        action='users',
        gns3util_path=gns3util_path
    )

    if not success:
        print(f"Error: Failed to fetch users. STDERR: {
              stderr}", file=sys.stderr)
        return None
    if not stdout:
        print("Error: No output received when fetching users.", file=sys.stderr)
        return None

    try:
        users_list = json.loads(stdout)
        if not isinstance(users_list, list):
            print(f"Error: Expected a list of users, but got {
                  type(users_list)}. Output: {stdout}", file=sys.stderr)
            return None

        user_map = {user['username']: user['user_id']
                    for user in users_list if 'username' in user and 'user_id' in user}
        print(f"Successfully mapped {len(user_map)} users.")
        return user_map
    except json.JSONDecodeError:
        print(f"Error: Could not decode JSON from get users output: {
              stdout}", file=sys.stderr)
        return None
    except Exception as e:
        print(f"Error processing user list: {e}", file=sys.stderr)
        return None


def create_user_and_add_to_group(
    server_url: str,
    group_id: str,
    username: str,
    password: str,
    full_name: str,
    is_superuser: bool = False,
    gns3util_path: str = DEFAULT_GNS3UTIL_PATH
) -> Optional[str]:
    """
    Creates a single user and adds them to a specified group (single-threaded).

    Args:
        server_url: The GNS3 server URL.
        group_id: The ID of the group to add the user to. Can be empty/None.
        username: The username for the new user.
        password: The password for the new user.
        full_name: The full name for the new user.
        is_superuser: Whether the user should be a superuser.
        gns3util_path: Path to the gns3util executable.

    Returns:
        The user_id of the newly created user if successful, otherwise None.
    """
    print(f"Attempting to create user '{username}'...")
    user_payload = {
        "username": username,
        "password": password,
        "name": full_name,
        "is_superuser": is_superuser
    }
    success_create, stdout_create, stderr_create = run_gns3util_command(
        server_url=server_url,
        subcommand='post',
        action='user',
        data=user_payload,
        gns3util_path=gns3util_path
    )

    if not success_create:
        print(f"Error: Failed to create user '{
              username}'. STDERR: {stderr_create}", file=sys.stderr)
        return None

    new_user_id = None
    if success_create:
        # Use regex to find the user_id in the stdout string
        match = re.search(
            r"\"user_id\"\s*:\s*\"([a-f0-9-]+)\"", stdout_create, re.IGNORECASE)
        if match:
            new_user_id = match.group(1)
            print(f"Successfully created user '{
                  username}' and extracted ID via regex: {new_user_id}")
        else:
            print(f"Error: User '{username}' creation reported success, but couldn't extract user_id "
                  f"via regex from output: {stdout_create}", file=sys.stderr)
            return None
    else:
        print(f"Error: Failed to create user '{
              username}'. STDERR: {stderr_create}", file=sys.stderr)
        if stderr_create and "already exists" in stderr_create.lower():
            print(f"Info: User '{
                  username}' might already exist. Attempting to find ID via user map...")
            user_map = get_user_id_map(server_url, gns3util_path)
            if user_map and username in user_map:
                new_user_id = user_map[username]
                print(f"Found existing user ID for '{
                      username}': {new_user_id}")
            else:
                print(f"Error: Could not find user ID for '{
                      username}' even after checking user map.", file=sys.stderr)
                return None
        else:
            return None

    # Add the user to the group if group_id is provided
    if new_user_id and group_id:
        print(f"Adding user {username} ({new_user_id}) to group {group_id}...")
        success_add, stdout_add, stderr_add = run_gns3util_command(
            server_url=server_url,
            subcommand='put',
            action='add_group_member',
            args=[group_id, new_user_id],
            gns3util_path=gns3util_path
        )
        if success_add:
            print(f"Successfully added user '{
                  username}' to group '{group_id}'.")
        else:
            print(f"Error: Failed to add user '{username}' to group '{
                  group_id}'. STDERR: {stderr_add}", file=sys.stderr)
    elif not group_id:
        print(f"Skipping add to group step as no group_id was provided.")

    return new_user_id


def _delete_single_user(
    server_url: str,
    user_id: str,
    username: str,
    gns3util_path: str
) -> Tuple[str, bool]:
    """Helper function to delete one user."""
    # print(f"Deleting user {username} ({user_id})...") # Verbose logging
    success, stdout, stderr = run_gns3util_command(
        server_url=server_url,
        subcommand='delete',
        action='user',
        args=[user_id],
        gns3util_path=gns3util_path
    )
    if not success:
        print(f"Error deleting user {username} ({user_id}). STDERR: {
              stderr}", file=sys.stderr)
    return user_id, success


def delete_users_sequentially(
    server_url: str,
    user_map: Dict[str, str],
    gns3util_path: str = DEFAULT_GNS3UTIL_PATH,
    skip_usernames: set = PROTECTED_USERNAMES
) -> Tuple[int, int]:
    """
    Deletes users from the provided map sequentially (single-threaded).

    Args:
        server_url: The GNS3 server URL.
        user_map: Dictionary {username: user_id} of users to consider for deletion.
        gns3util_path: Path to the gns3util executable.
        skip_usernames: Set of usernames to NOT delete.

    Returns:
        Tuple (deleted_count, failed_count).
    """
    users_to_delete = {
        user_id: username for username, user_id in user_map.items()
        if username not in skip_usernames
    }

    if not users_to_delete:
        print("No users eligible for deletion (after skipping protected users).")
        return 0, 0

    print(f"\nStarting sequential deletion of {len(users_to_delete)} users...")

    deleted_count = 0
    failed_count = 0
    total_tasks = len(users_to_delete)

    for i, (user_id, username) in enumerate(users_to_delete.items(), 1):
        print(f"\rProgress: Deleting user {
              i}/{total_tasks} ({username})...", end="")
        _user_id, success = _delete_single_user(
            server_url,
            user_id,
            username,
            gns3util_path
        )
        if success:
            deleted_count += 1
        else:
            failed_count += 1
            print()

    print()  # Newline after progress indicator
    print(f"\nSequential deletion complete. Successfully deleted: {
          deleted_count}, Failed: {failed_count}")
    return deleted_count, failed_count


def delete_users_concurrently(
    server_url: str,
    user_map: Dict[str, str],
    max_workers: int = DEFAULT_MAX_DELETE_WORKERS,
    gns3util_path: str = DEFAULT_GNS3UTIL_PATH,
    skip_usernames: set = PROTECTED_USERNAMES
) -> Tuple[int, int]:
    """
    Deletes users from the provided map concurrently, skipping protected usernames.

    Args:
        server_url: The GNS3 server URL.
        user_map: Dictionary {username: user_id} of users to consider for deletion.
        max_workers: Maximum number of concurrent deletion threads.
        gns3util_path: Path to the gns3util executable.
        skip_usernames: Set of usernames to NOT delete.

    Returns:
        Tuple (deleted_count, failed_count).
    """
    users_to_delete = {
        user_id: username for username, user_id in user_map.items()
        if username not in skip_usernames
    }

    if not users_to_delete:
        print("No users eligible for deletion (after skipping protected users).")
        return 0, 0

    print(f"\nStarting concurrent deletion of {len(users_to_delete)} users "
          f"(max workers: {max_workers})...")

    deleted_count = 0
    failed_count = 0
    futures = []

    with ThreadPoolExecutor(max_workers=max_workers, thread_name_prefix='DeleteUserWorker') as executor:
        for user_id, username in users_to_delete.items():
            future = executor.submit(
                _delete_single_user,
                server_url,
                user_id,
                username,
                gns3util_path
            )
            futures.append(future)

        total_tasks = len(futures)
        for i, future in enumerate(as_completed(futures), 1):
            try:
                _user_id, success = future.result()
                if success:
                    deleted_count += 1
                else:
                    failed_count += 1
            except Exception as exc:
                print(f"\nError processing deletion task result: {
                      exc}", file=sys.stderr)
                failed_count += 1

            # progress indicator
            print(f"\rProgress: {
                  i}/{total_tasks} deletions processed...", end="")
        print()

    print(f"\nConcurrent deletion complete. Successfully deleted: {
          deleted_count}, Failed: {failed_count}")
    return deleted_count, failed_count


# --- Helper Function for Deletion ---

def _handle_deletion(delete_mode: str, server_url: str, gns3util_path: str):
    """Gets user map, confirms, and calls the appropriate deletion function."""
    user_map = get_user_id_map(server_url, gns3util_path)
    if not user_map:
        print("Could not retrieve user map. Deletion aborted.")
        return

    num_eligible = len([u for u in user_map if u not in PROTECTED_USERNAMES])
    if num_eligible == 0:
        print(f"No users found eligible for deletion (excluding {
              ', '.join(PROTECTED_USERNAMES)}).")
        return

    confirm = input(f"Found {len(user_map)} total users. {num_eligible} are eligible for deletion "
                    f"(excluding {', '.join(PROTECTED_USERNAMES)}). "
                    f"Proceed with {delete_mode} deletion? (yes/no): ").strip().lower()

    if confirm != 'yes':
        print("Deletion cancelled.")
        return

    if delete_mode == 'multi-threaded':
        try:
            max_workers_str = input(f"Enter max concurrent deletions [{
                                    DEFAULT_MAX_DELETE_WORKERS}]: ").strip()
            max_workers = int(
                max_workers_str) if max_workers_str else DEFAULT_MAX_DELETE_WORKERS
            if max_workers <= 0:
                max_workers = 1
        except ValueError:
            print(f"Invalid number, using default {
                  DEFAULT_MAX_DELETE_WORKERS} workers.")
            max_workers = DEFAULT_MAX_DELETE_WORKERS

        delete_users_concurrently(
            server_url=server_url,
            user_map=user_map,
            max_workers=max_workers,
            gns3util_path=gns3util_path,
            skip_usernames=PROTECTED_USERNAMES
        )
    elif delete_mode == 'single-threaded':
        delete_users_sequentially(
            server_url=server_url,
            user_map=user_map,
            gns3util_path=gns3util_path,
            skip_usernames=PROTECTED_USERNAMES
        )
    else:
        print(f"Error: Unknown deletion mode '{delete_mode}'", file=sys.stderr)

# --- Main Execution Logic ---


def main():
    """Main function to handle user interaction and orchestrate tasks."""
    print("--- GNS3 Bulk User Management ---")

    gns3_server_url = input(
        f"Enter GNS3 Server URL [{DEFAULT_GNS3_SERVER_URL}]: ").strip()
    if not gns3_server_url:
        gns3_server_url = DEFAULT_GNS3_SERVER_URL
    print(f"Using GNS3 Server: {gns3_server_url}")

    gns3util_path = input(f"Enter path to gns3util executable [{
                          DEFAULT_GNS3UTIL_PATH}]: ").strip()
    if not gns3util_path:
        gns3util_path = DEFAULT_GNS3UTIL_PATH
    print(f"Using gns3util path: {gns3util_path}")

    while True:
        print("\nChoose an action:")
        print("  1. Create users (single-threaded)")
        print("  2. Delete users (multi-threaded, skips protected)")
        print("  3. Delete users (single-threaded, skips protected)")  # New option
        print("  q. Quit")
        choice = input("Enter your choice (1, 2, 3, or q): ").strip().lower()

        if choice == '1':
            # --- Create Users (Single-threaded) ---
            try:
                num_users = int(
                    input("Enter the number of users to create: ").strip())
                if num_users <= 0:
                    print("Please enter a positive number.")
                    continue
            except ValueError:
                print("Invalid number.")
                continue

            group_id = input("Enter the Group ID to add users to (leave blank to skip): ").strip(
            ) or None
            base_username = input(
                "Enter a base username (e.g., 'student'): ").strip()
            if not base_username:
                print("Base username cannot be empty.")
                continue

            print(f"\nStarting creation of {
                  num_users} users with base '{base_username}'...")
            created_count = 0
            failed_count = 0
            start_time = time.monotonic()
            for i in range(num_users):
                unique_suffix = str(uuid.uuid4()).split('-')[0]
                new_username = f"{base_username}_{unique_suffix}"
                new_password = "ermwhatthesigma"
                new_full_name = f"{base_username.capitalize()} User {
                    i+1} ({unique_suffix})"

                print(f"\n--- Creating User {i+1}/{num_users} ---")
                user_id = create_user_and_add_to_group(
                    server_url=gns3_server_url,
                    group_id=group_id,
                    username=new_username,
                    password=new_password,
                    full_name=new_full_name,
                    gns3util_path=gns3util_path
                )
                if user_id:
                    created_count += 1
                else:
                    failed_count += 1
                # time.sleep(0.1) # Optional delay

            end_time = time.monotonic()
            duration = end_time - start_time
            print(f"\n--- Creation Summary ---")
            print(f"Attempted: {num_users}")
            print(f"Successful: {created_count}")
            print(f"Failed: {failed_count}")
            print(f"Duration: {duration:.2f} seconds")
            print("-" * 30)

        elif choice == '2':
            # --- Delete Users (Multi-threaded) ---
            start_time = time.monotonic()
            _handle_deletion('multi-threaded', gns3_server_url, gns3util_path)
            end_time = time.monotonic()
            duration = end_time - start_time
            print(f"Multi-threaded deletion duration: {duration:.2f} seconds")
            print("-" * 30)

        elif choice == '3':
            # --- Delete Users (Single-threaded) ---
            start_time = time.monotonic()
            _handle_deletion('single-threaded', gns3_server_url, gns3util_path)
            end_time = time.monotonic()
            duration = end_time - start_time
            print(f"Single-threaded deletion duration: {duration:.2f} seconds")
            print("-" * 30)

        elif choice == 'q':
            print("Exiting.")
            break
        else:
            print("Invalid choice. Please enter 1, 2, 3, or q.")


if __name__ == '__main__':
    main()
