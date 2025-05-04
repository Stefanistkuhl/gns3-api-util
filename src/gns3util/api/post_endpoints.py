# from . import GNS3APIClient
from .client import GNS3APIClient


class GNS3PostAPI(GNS3APIClient):

    # controller endpoints

    def check_version(self, data):
        return self._api_call("version", method="POST", data=data)

    def shutdown_controller(self):
        return self._api_call("shutdown", method="POST")

    def login_controller(self):
        return self._api_call("shutdown", method="POST")

    # user endpoints

    def user_authenticate(self, data):
        return self._api_call("access/users/authenticate", method="POST", data=data)

    def create_user(self, data):
        return self._api_call("access/users", method="POST", data=data)

    # group endpoints
    def create_group(self, data):
        return self._api_call("access/groups", method="POST", data=data)

    # role endpoints
    def create_role(self, data):
        return self._api_call("access/roles", method="POST", data=data)

    # acl endpoints
    def create_acl(self, data):
        return self._api_call("access/acl", method="POST", data=data)

    # images endpoints
    def create_qemu_image(self, image_path, data):
        return self._api_call(f"images/qemu/{image_path}", method="POST", data=data)

    def upload_image(self, image_path, install_appliances):
        return self._api_call(f"images/upload/{image_path}?install_appliances={install_appliances}", method="POST")

    def install_image(self):
        return self._api_call("images/install", method="POST")

    # template endpoints

    def create_template(self, data):
        return self._api_call("templates", method="POST", data=data)

    def duplicate_template(self, template_id):
        return self._api_call(f"templates/{template_id}/duplicate", method="POST")

    # project endpoints

    def create_project(self, data):
        return self._api_call("projects", method="POST", data=data)

    def close_project(self, project_id):
        return self._api_call(f"projects/{project_id}/close", method="POST")

    def open_project(self, project_id):
        return self._api_call(f"projects/{project_id}/open", method="POST")

    def load_project(self, data):
        return self._api_call(f"projects/load", method="POST", data=data)

    def import_project(self, project_id, name):
        return self._api_call(f"projects/{project_id}/import?name={name}", method="POST")

    def duplicate_project(self, project_id):
        return self._api_call(f"projects/{project_id}/duplicate", method="POST")

    def lock_project(self, project_id):
        return self._api_call(f"projects/{project_id}/lock", method="POST")

    def unlock_project(self, project_id):
        return self._api_call(f"projects/{project_id}/unlock", method="POST")

    def write_project_file(self, project_id, file_path):
        return self._api_call(f"projects/{project_id}/files/{file_path}", method="POST")

    def create_project_node_from_template(self, project_id, template_id, data):
        return self._api_call(f"projects/{project_id}/templates/{template_id}", method="POST", data=data)

    # node endpoints
    def create_node(self, project_id, data):
        return self._api_call(f"projects/{project_id}/nodes", method="POST", data=data)

    def start_node(self, project_id):
        return self._api_call(f"projects/{project_id}/nodes/start", method="POST")

    def stop_node(self, project_id):
        return self._api_call(f"projects/{project_id}/nodes/stop", method="POST")

    def suspend_node(self, project_id):
        return self._api_call(f"projects/{project_id}/nodes/suspend", method="POST")

    def reload_node(self, project_id):
        return self._api_call(f"projects/{project_id}/nodes/reload", method="POST")

    def duplicate_node(self, project_id, node_id, data):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/duplicate", method="POST", data=data)

    def isolate_node(self, project_id, node_id):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/isolate", method="POST")

    def unisolate_node(self, project_id, node_id):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/unisolate", method="POST")

    def create_disk_image(self, project_id, node_id, disk_name, data):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/qemu/disk_image/{disk_name}", method="POST", data=data)

    def create_node_file(self, project_id, node_id, file_path):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/files/{file_path}", method="POST")

    def reset_nodes_console(self, project_id):
        return self._api_call(f"projects/{project_id}/nodes/console/reset", method="POST")

    def reset_node_console(self, project_id, node_id):
        return self._api_call(f"projects/{project_id}/node/{node_id}/console/reset", method="POST")

    # link endpoints

    def create_link(self, project_id, data):
        return self._api_call(f"projects/{project_id}/links", method="POST", data=data)

    def reset_link(self, project_id, link_id):
        return self._api_call(f"projects/{project_id}/links/{link_id}/reset", method="POST")

    def start_link_capture(self, project_id, link_id, data):
        return self._api_call(f"projects/{project_id}/links/{link_id}/capture/start", method="POST", data=data)

    def stop_link_capture(self, project_id, link_id):
        return self._api_call(f"projects/{project_id}/links/{link_id}/capture/stop", method="POST")

    # drawing endpoint

    # this doenst work lol maybe i just tested wrong
    def create_drawing(self, project_id, data):
        return self._api_call(f"projects/{project_id}/drawings", method="POST", data=data)

    # symbol endpoint

    def create_symbol(self, symbol_id):
        return self._api_call(f"symbols/{symbol_id}/raw", method="POST")

    # snapshot endpoints

    def create_snapshot(self, project_id, data):
        return self._api_call(f"projects/{project_id}/snapshots", method="POST", data=data)

    def restore_snapshot(self, project_id, snapshot_id):
        return self._api_call(f"projects/{project_id}/snapshots/{snapshot_id}/restore", method="POST")

    # compute endpoints

    def create_compute(self, connect, data):
        return self._api_call(f"computes?connect={connect}", method="POST", data=data)

    def connect_compute(self, compute_id):
        return self._api_call(f"computes/{compute_id}/connect", method="POST")

    def set_auto_idlepc(self, compute_id, data):
        return self._api_call(f"computes/{compute_id}/dynamips/auto_idlepc", method="POST", data=data)

    # applieance endpoints

    def create_appliance_version(self, appliance_id, data):
        return self._api_call(f"appliances/{appliance_id}/version", method="POST", data=data)

    def install_appliance_version(self, appliance_id, version):
        return self._api_call(f"appliances/{appliance_id}/install?version={version}", method="POST")

    # pool endpoint

    def create_pool(self, data):
        return self._api_call(f"pools", method="POST", data=data)
