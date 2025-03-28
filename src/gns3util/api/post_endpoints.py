from . import GNS3APIClient


class GNS3PostAPI(GNS3APIClient):

    # controller endpoints

    def version_check(self, data):
        return self._api_call("version", method="POST", data=data)

    def controller_shutdown(self):
        return self._api_call("shutdown", method="POST")

    def controller_login(self):
        return self._api_call("shutdown", method="POST")

    # user endpoints
    # todo add the authenticatie endpoint i am too lazy for rn
    # make in part of the auth file actually

    def user_create(self, data):
        return self._api_call("access/users", method="POST", data=data)

    # group endpoints
    def group_create(self, data):
        return self._api_call("access/groups", method="POST", data=data)

    # role endpoints
    def role_create(self, data):
        return self._api_call("access/roles", method="POST", data=data)

    # acl endpoints
    def acl_create(self, data):
        return self._api_call("access/acl", method="POST", data=data)

    # images endpoints
    def qemu_image_create(self, image_path, data):
        return self._api_call(f"images/qemu/{image_path}", method="POST", data=data)

    def image_upload(self, image_path, install_appliances):
        return self._api_call(f"images/upload/{image_path}?install_appliances={install_appliances}", method="POST")

    def image_install(self):
        return self._api_call("images/install", method="POST")

    # template endpoints

    def template_create(self, data):
        return self._api_call("templates", method="POST", data=data)

    def template_duplicate(self, template_id):
        return self._api_call(f"templates/{template_id}/duplicate", method="POST")

    # project endpoints

    def project_create(self, data):
        return self._api_call("projects", method="POST", data=data)

    def project_close(self, project_id):
        return self._api_call(f"projects/{project_id}/close", method="POST")

    def project_open(self, project_id):
        return self._api_call(f"projects/{project_id}/open", method="POST")

    def project_load(self, data):
        return self._api_call(f"projects/load", method="POST", data=data)

    def project_import(self, project_id, name):
        return self._api_call(f"projects/{project_id}/import?name={name}", method="POST")

    def project_duplicate(self, project_id):
        return self._api_call(f"projects/{project_id}/duplicate", method="POST")

    def project_lock(self, project_id):
        return self._api_call(f"projects/{project_id}/lock", method="POST")

    def project_unlock(self, project_id):
        return self._api_call(f"projects/{project_id}/unlock", method="POST")

    def project_file_write(self, project_id, file_path):
        return self._api_call(f"projects/{project_id}/files/{file_path}", method="POST")

    def project_node_create_from_template(self, project_id, template_id, data):
        return self._api_call(f"projects/{project_id}/templates/{template_id}", method="POST", data=data)

    # node endpoints
    def node_create(self, project_id, data):
        return self._api_call(f"projects/{project_id}/nodes", method="POST", data=data)

    def node_start(self, project_id):
        return self._api_call(f"projects/{project_id}/nodes/start", method="POST")

    def node_stop(self, project_id):
        return self._api_call(f"projects/{project_id}/nodes/stop", method="POST")

    def node_suspend(self, project_id):
        return self._api_call(f"projects/{project_id}/nodes/suspend", method="POST")

    def node_reload(self, project_id):
        return self._api_call(f"projects/{project_id}/nodes/reload", method="POST")

    def node_duplicate(self, project_id, node_id, data):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/duplicate", method="POST", data=data)

    def node_isolate(self, project_id, node_id):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/isolate", method="POST")

    def node_unisolate(self, project_id, node_id):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/unisolate", method="POST")

    def create_disk_img(self, project_id, node_id, disk_name, data):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/qemu/disk_image/{disk_name}", method="POST", data=data)

    def node_create_file(self, project_id, node_id, file_path):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/files/{file_path}", method="POST")

    def nodes_console_reset(self, project_id):
        return self._api_call(f"projects/{project_id}/nodes/console/reset", method="POST")

    def node_console_reset(self, project_id, node_id):
        return self._api_call(f"projects/{project_id}/node/{node_id}/console/reset", method="POST")

    # link endpoints

    def link_create(self, project_id, data):
        return self._api_call(f"projects/{project_id}/links", method="POST", data=data)

    def link_reset(self, project_id, link_id):
        return self._api_call(f"projects/{project_id}/links/{link_id}/reset", method="POST")

    def start_link_capture(self, project_id, link_id, data):
        return self._api_call(f"projects/{project_id}/links/{link_id}/capture/start", method="POST", data=data)

    def stop_link_capture(self, project_id, link_id):
        return self._api_call(f"projects/{project_id}/links/{link_id}/capture/stop", method="POST")

    # drawing endpoint

    # this doenst work lol maybe i just tested wrong
    def drawing_create(self, project_id, data):
        return self._api_call(f"projects/{project_id}/drawings", method="POST", data=data)

    # symbol endpoint

    def symbol_create(self, symbol_id):
        return self._api_call(f"symbols/{symbol_id}/raw", method="POST")

    # snapshot endpoints

    def snapshot_create(self, project_id, data):
        return self._api_call(f"projects/{project_id}/snapshots", method="POST", data=data)

    def snapshot_restore(self, project_id, snapshot_id):
        return self._api_call(f"projects/{project_id}/snapshots/{snapshot_id}/restore", method="POST")

    # compute endpoints

    def create_compute(self, connect, data):
        return self._api_call(f"computes?connect={connect}", method="POST", data=data)

    def connect_compute(self, compute_id):
        return self._api_call(f"computes/{compute_id}/connect", method="POST")

    def auto_idlepc(self, compute_id, data):
        return self._api_call(f"computes/{compute_id}/dynamips/auto_idlepc", method="POST", data=data)

    # applieance endpoints

    def add_applience_version(self, appliance_id, data):
        return self._api_call(f"appliances/{appliance_id}/version", method="POST", data=data)

    def add_applience_version(self, appliance_id, version):
        return self._api_call(f"appliances/{appliance_id}/install?version={version}", method="POST")

    # pool endpoint

    def add_pool(self, data):
        return self._api_call(f"pools", method="POST", data=data)
