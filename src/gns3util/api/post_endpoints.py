from .client import GNS3APIClient


class GNS3PostAPI(GNS3APIClient):
    # controller endpoints

    def check_version(self, data):
        return self._api_call("version", method="POST", data=data, verify=self.verify)

    def shutdown_controller(self):
        return self._api_call("shutdown", method="POST", verify=self.verify)

    def login_controller(self):
        return self._api_call("shutdown", method="POST", verify=self.verify)

    # user endpoints

    def user_authenticate(self, data):
        return self._api_call(
            "access/users/authenticate", method="POST", data=data, verify=self.verify
        )

    def create_user(self, data):
        return self._api_call(
            "access/users", method="POST", data=data, verify=self.verify
        )

    # group endpoints
    def create_group(self, data):
        return self._api_call(
            "access/groups", method="POST", data=data, verify=self.verify
        )

    # role endpoints
    def create_role(self, data):
        return self._api_call(
            "access/roles", method="POST", data=data, verify=self.verify
        )

    # acl endpoints
    def create_acl(self, data):
        return self._api_call(
            "access/acl", method="POST", data=data, verify=self.verify
        )

    # images endpoints
    def create_qemu_image(self, image_path, data):
        return self._api_call(
            f"images/qemu/{image_path}", method="POST", data=data, verify=self.verify
        )

    def upload_image(self, image_path, install_appliances):
        return self._api_call(
            f"images/upload/{image_path}?install_appliances={install_appliances}",
            method="POST",
            verify=self.verify,
        )

    def install_image(self):
        return self._api_call("images/install", method="POST", verify=self.verify)

    # template endpoints

    def create_template(self, data):
        return self._api_call("templates", method="POST", data=data, verify=self.verify)

    def duplicate_template(self, template_id):
        return self._api_call(
            f"templates/{template_id}/duplicate", method="POST", verify=self.verify
        )

    # project endpoints

    def create_project(self, data):
        return self._api_call("projects", method="POST", data=data, verify=self.verify)

    def close_project(self, project_id):
        return self._api_call(
            f"projects/{project_id}/close", method="POST", verify=self.verify
        )

    def open_project(self, project_id):
        return self._api_call(
            f"projects/{project_id}/open", method="POST", verify=self.verify
        )

    def load_project(self, data):
        return self._api_call(
            f"projects/load", method="POST", data=data, verify=self.verify
        )

    def import_project(self, project_id, name):
        return self._api_call(
            f"projects/{project_id}/import?name={name}",
            method="POST",
            verify=self.verify,
        )

    def duplicate_project(self, project_id, data):
        return self._api_call(
            f"projects/{project_id}/duplicate",
            method="POST",
            data=data,
            verify=self.verify,
        )

    def lock_project(self, project_id):
        return self._api_call(
            f"projects/{project_id}/lock", method="POST", verify=self.verify
        )

    def unlock_project(self, project_id):
        return self._api_call(
            f"projects/{project_id}/unlock", method="POST", verify=self.verify
        )

    def write_project_file(self, project_id, file_path):
        return self._api_call(
            f"projects/{project_id}/files/{file_path}",
            method="POST",
            verify=self.verify,
        )

    def create_project_node_from_template(self, project_id, template_id, data):
        return self._api_call(
            f"projects/{project_id}/templates/{template_id}",
            method="POST",
            data=data,
            verify=self.verify,
        )

    # node endpoints
    def create_node(self, project_id, data):
        return self._api_call(
            f"projects/{project_id}/nodes", method="POST", data=data, verify=self.verify
        )

    def start_nodes(self, project_id):
        return self._api_call(
            f"projects/{project_id}/nodes/start", method="POST", verify=self.verify
        )

    def start_node(self, project_id, node_id):
        return self._api_call(
            f"projects/{project_id}/nodes/{node_id}/start",
            method="POST",
            verify=self.verify,
        )

    def stop_nodes(self, project_id):
        return self._api_call(
            f"projects/{project_id}/nodes/stop", method="POST", verify=self.verify
        )

    def stop_node(self, project_id, node_id):
        return self._api_call(
            f"projects/{project_id}/nodes/{node_id}/stop",
            method="POST",
            verify=self.verify,
        )

    def suspend_nodes(self, project_id):
        return self._api_call(
            f"projects/{project_id}/nodes/suspend", method="POST", verify=self.verify
        )

    def reload_nodes(self, project_id):
        return self._api_call(
            f"projects/{project_id}/nodes/reload", method="POST", verify=self.verify
        )

    def suspend_node(self, project_id, node_id):
        return self._api_call(
            f"projects/{project_id}/nodes/{node_id}/suspend",
            method="POST",
            verify=self.verify,
        )

    def reload_node(self, project_id, node_id):
        return self._api_call(
            f"projects/{project_id}/nodes/{node_id}/reload",
            method="POST",
            verify=self.verify,
        )

    def duplicate_node(self, project_id, node_id, data):
        return self._api_call(
            f"projects/{project_id}/nodes/{node_id}/duplicate",
            method="POST",
            data=data,
            verify=self.verify,
        )

    def isolate_node(self, project_id, node_id):
        return self._api_call(
            f"projects/{project_id}/nodes/{node_id}/isolate",
            method="POST",
            verify=self.verify,
        )

    def unisolate_node(self, project_id, node_id):
        return self._api_call(
            f"projects/{project_id}/nodes/{node_id}/unisolate",
            method="POST",
            verify=self.verify,
        )

    def create_disk_image(self, project_id, node_id, disk_name, data):
        return self._api_call(
            f"projects/{project_id}/nodes/{node_id}/qemu/disk_image/{disk_name}",
            method="POST",
            data=data,
            verify=self.verify,
        )

    def create_node_file(self, project_id, node_id, file_path):
        return self._api_call(
            f"projects/{project_id}/nodes/{node_id}/files/{file_path}",
            method="POST",
            verify=self.verify,
        )

    def reset_nodes_console(self, project_id):
        return self._api_call(
            f"projects/{project_id}/nodes/console/reset",
            method="POST",
            verify=self.verify,
        )

    def reset_node_console(self, project_id, node_id):
        return self._api_call(
            f"projects/{project_id}/node/{node_id}/console/reset",
            method="POST",
            verify=self.verify,
        )

    # link endpoints

    def create_link(self, project_id, data):
        return self._api_call(
            f"projects/{project_id}/links", method="POST", data=data, verify=self.verify
        )

    def reset_link(self, project_id, link_id):
        return self._api_call(
            f"projects/{project_id}/links/{link_id}/reset",
            method="POST",
            verify=self.verify,
        )

    def start_link_capture(self, project_id, link_id, data):
        return self._api_call(
            f"projects/{project_id}/links/{link_id}/capture/start",
            method="POST",
            data=data,
            verify=self.verify,
        )

    def stop_link_capture(self, project_id, link_id):
        return self._api_call(
            f"projects/{project_id}/links/{link_id}/capture/stop",
            method="POST",
            verify=self.verify,
        )

    # drawing endpoint

    def create_drawing(self, project_id, data):
        return self._api_call(
            f"projects/{project_id}/drawings",
            method="POST",
            data=data,
            verify=self.verify,
        )

    # symbol endpoint

    def create_symbol(self, symbol_id):
        return self._api_call(
            f"symbols/{symbol_id}/raw", method="POST", verify=self.verify
        )

    # snapshot endpoints

    def create_snapshot(self, project_id, data):
        return self._api_call(
            f"projects/{project_id}/snapshots",
            method="POST",
            data=data,
            verify=self.verify,
        )

    def restore_snapshot(self, project_id, snapshot_id):
        return self._api_call(
            f"projects/{project_id}/snapshots/{snapshot_id}/restore",
            method="POST",
            verify=self.verify,
        )

    # compute endpoints

    def create_compute(self, connect, data):
        return self._api_call(
            f"computes?connect={connect}", method="POST", data=data, verify=self.verify
        )

    def connect_compute(self, compute_id):
        return self._api_call(
            f"computes/{compute_id}/connect", method="POST", verify=self.verify
        )

    def set_auto_idlepc(self, compute_id, data):
        return self._api_call(
            f"computes/{compute_id}/dynamips/auto_idlepc",
            method="POST",
            data=data,
            verify=self.verify,
        )

    # applieance endpoints

    def create_appliance_version(self, appliance_id, data):
        return self._api_call(
            f"appliances/{appliance_id}/version",
            method="POST",
            data=data,
            verify=self.verify,
        )

    def install_appliance_version(self, appliance_id, version):
        return self._api_call(
            f"appliances/{appliance_id}/install?version={version}",
            method="POST",
            verify=self.verify,
        )

    # pool endpoint

    def create_pool(self, data):
        return self._api_call(f"pools", method="POST", data=data, verify=self.verify)
