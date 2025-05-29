# from . import GNS3APIClient
from .client import GNS3APIClient


class GNS3PutAPI(GNS3APIClient):

    # controller endpoint

    def iou_license(self, data):
        return self._api_call("iou_license", method="PUT", data=data, verify=self.verify)

    # user endpoints

    def me(self, data):
        return self._api_call("access/users/me", method="PUT", data=data, verify=self.verify)

    def update_user(self, user_id, data):
        return self._api_call(f"access/users/{user_id}", method="PUT", data=data, verify=self.verify)

    # user group endpoints

    def update_group(self, group_id, data):
        return self._api_call(f"access/groups/{group_id}", method="PUT", data=data, verify=self.verify)

    def add_group_member(self, group_id, user_id):
        return self._api_call(f"access/groups/{group_id}/members/{user_id}", method="PUT", verify=self.verify)

    # role endpoints

    def update_role(self, role_id, data):
        return self._api_call(f"access/roles/{role_id}", method="PUT", data=data, verify=self.verify)

    def update_role_privs(self, role_id, priv_id):
        return self._api_call(f"access/roles/{role_id}/privileges/{priv_id}", method="PUT", verify=self.verify)

    # acl endpoint

    def update_ace(self, ace_id, data):
        return self._api_call(f"access/acl/{ace_id}", method="PUT", data=data, verify=self.verify)

    # template endpoint

    def update_template(self, template_id, data):
        return self._api_call(f"templates/{template_id}", method="PUT", data=data, verify=self.verify)

    # project endpoint

    def update_project(self, project_id, data):
        return self._api_call(f"projects/{project_id}", method="PUT", data=data, verify=self.verify)

    # node enpoint

    def update_node(self, project_id, node_id, data):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}", method="PUT", data=data, verify=self.verify)

    def update_disk_image(self, project_id, node_id, disk_name, data):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/qemu/disk_image/{disk_name}", method="PUT", data=data, verify=self.verify)

    # link endpoint

    def update_link(self, project_id, link_id, data):
        return self._api_call(f"projects/{project_id}/links/{link_id}", method="PUT", data=data, verify=self.verify)

    # drawing endpoint

    def update_drawing(self, project_id, drawing_id, data):
        return self._api_call(f"projects/{project_id}/drawings/{drawing_id}", method="PUT", data=data, verify=self.verify)

    # compute endpoint

    def update_compute(self, compute_id, data):
        return self._api_call(f"computes/{compute_id}", method="PUT", data=data, verify=self.verify)

    # ressouce pool endpoints

    def update_pool(self, pool_id, data):
        return self._api_call(f"pools/{pool_id}", method="PUT", data=data, verify=self.verify)

    def add_resource_to_pool(self, pool_id, resource_id):
        return self._api_call(f"pools/{pool_id}/resources/{resource_id}", method="PUT", verify=self.verify)
