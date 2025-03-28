from ..api import GNS3APIClient
import urllib.parse


class GNS3GetAPI(GNS3APIClient):
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
        self._stream_notifications(
            f"projects/{project_id}/notifications", timeout_seconds)

    def project_locked(self, project_id):
        return self._api_call(f"projects/{project_id}/locked")

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

    # Drawing endpoints
    def drawings(self, project_id):
        return self._api_call(f"projects/{project_id}/drawings")

    def drawing(self, project_id, drawing_id):
        return self._api_call(f"projects/{project_id}/drawings/{drawing_id}")

    # Symbols endpoints
    def symbols(self):
        return self._api_call("symbols")

    def symbol(self, symbol_id):
        encoded_symbol_id = urllib.parse.quote(symbol_id, safe='')
        return self._api_call(f"symbols/{encoded_symbol_id}/raw")

    def symbolDimensions(self, symbol_id):
        return self._api_call(f"symbols/{symbol_id}/dimensions")

    def defaultSymbols(self):
        return self._api_call("symbols/default_symbols")

    # Snapshot endpoints
    def snapshots(self, project_id):
        return self._api_call(f"projects/{project_id}/snapshots")

    # Compute endpoints
    def computes(self):
        return self._api_call("computes")

    def computeByID(self, compute_id):
        return self._api_call(f"computes/{compute_id}")

    def computeByIDDockerImages(self, compute_id):
        return self._api_call(f"computes/{compute_id}/docker/images")

    def computeByIDVirtualvoxVms(self, compute_id):
        return self._api_call(f"computes/{compute_id}/virtualbox/vms")

    def computeByIDVmwareVms(self, compute_id):
        return self._api_call(f"computes/{compute_id}/vmware/vms")

    # Appliance endpoints
    def appliances(self):
        return self._api_call("appliances")

    def appliance(self, appliance_id):
        return self._api_call(f"appliances/{appliance_id}")

    # Resource pools endpoints
    def pools(self):
        return self._api_call("pools")

    # images endpoints

    def images(self, image_type):
        return self._api_call(f"images?image_type={image_type}")

    def imagesByPath(self, image_path):
        return self._api_call(f"images/{image_path}")

    def pool(self, resource_pool_id):
        return self._api_call(f"pools/{resource_pool_id}")

    def poolResources(self, resource_pool_id):
        return self._api_call(f"pools/{resource_pool_id}/resources")

    # Project file and export methods
    def download_project_file(self, project_id, file_path):
        encoded_file_path = urllib.parse.quote(file_path)
        return self._api_call(f"projects/{project_id}/files/{encoded_file_path}")

    def link_capture_stream(self, project_id, link_id, output_file=None, timeout=None):
        """Stream the PCAP capture file from a link."""
        url = f"projects/{project_id}/links/{link_id}/capture/stream"
        success, response = self._api_call(url, stream=True)
        if success and output_file:
            with open(output_file, 'wb') as f:
                for chunk in response.iter_content(chunk_size=8192):
                    f.write(chunk)
            return True, output_file
        return success, response if success else None
