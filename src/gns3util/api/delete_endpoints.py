from . import GNS3APIClient


class GNS3DeleteAPI(GNS3APIClient):

    # user endpoint

    def delete_user(self, user_id):
        return self._api_call(f"access/users/{user_id}", method="DELETE")

    # group endpoints

    def delete_group(self, group_id):
        return self._api_call(f"access/groups/{group_id}", method="DELETE")

    def delete_user_from_group(self, group_id, user_id):
        return self._api_call(f"access/groups/{group_id}/members/{user_id}", method="DELETE")

    # role endpoints

    def delete_role(self, role_id):
        return self._api_call(f"access/roles/{role_id}", method="DELETE")

    def delete_role_priv(self, role_id, priv_id):
        return self._api_call(f"access/roles/{role_id}/privileges/{priv_id}", method="DELETE")

    # acl endpoint

    def delete_acl(self, ace_id):
        return self._api_call(f"access/acl/{ace_id}", method="DELETE")

    # images endpoints

    def prune_images(self):
        return self._api_call(f"images/prune", method="DELETE")

    def delete_image(self, image_path):
        return self._api_call(f"images/{image_path}", method="DELETE")

    # template endpoint

    def delete_template(self, template_id):
        return self._api_call(f"templates/{template_id}", method="DELETE")

    # project endpoint

    def delete_project(self, project_id):
        return self._api_call(f"projects/{project_id}", method="DELETE")

    # nodes endpoints

    def delete_node(self, project_id, node_id):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}", method="DELETE")

    def delete_disk_image(self, project_id, node_id, disk_name):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/qemu/disk_image/{disk_name}", method="DELETE")

    # link endpoint

    def delete_link(self, project_id, link_id):
        return self._api_call(f"projects/{project_id}/links/{link_id}", method="DELETE")

    # drawing endpoint

    def delete_drawing(self, project_id, drawing_id):
        return self._api_call(f"projects/{project_id}/drawings/{drawing_id}", method="DELETE")

    # snapshot endpoint

    def delete_snapshot(self, project_id, snapshot_id):
        return self._api_call(f"projects/{project_id}/snapshots/{snapshot_id}", method="DELETE")

    # compute endpoints

    def delete_compute(self, compute_id):
        return self._api_call(f"computes/{compute_id}", method="DELETE")

    # pool endpoints

    def delete_pool(self, pool_id):
        return self._api_call(f"pools/{pool_id}", method="DELETE")

    def delete_pool_resource(self, pool_id, resource_id):
        return self._api_call(f"pools/{pool_id}/resources/{resource_id}", method="DELETE")
