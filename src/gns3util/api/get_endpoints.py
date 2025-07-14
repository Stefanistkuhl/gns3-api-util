from .client import GNS3APIClient, GNS3Error
import urllib.parse
import click


class GNS3GetAPI(GNS3APIClient):
    # Controller endpoints
    def version(self):
        return self._api_call("version", verify=self.verify)

    def iou_license(self):
        return self._api_call("iou_license", verify=self.verify)

    def statistics(self):
        return self._api_call("statistics", verify=self.verify)

    def notifications(self, timeout_seconds=60):
        self._stream_notifications("notifications", timeout_seconds)

    # User endpoints
    def current_user_info(self):
        return self._api_call("access/users/me", verify=self.verify)

    def users(self):
        return self._api_call("access/users", verify=self.verify)

    def user(self, user_id):
        return self._api_call(f"access/users/{user_id}", verify=self.verify)

    def users_groups(self, user_id):
        return self._api_call(f"access/users/{user_id}/groups", verify=self.verify)

    # Project endpoints
    def projects(self):
        return self._api_call("projects", verify=self.verify)

    def project(self, project_id):
        return self._api_call(f"projects/{project_id}", verify=self.verify)

    def project_stats(self, project_id):
        return self._api_call(f"projects/{project_id}/stats", verify=self.verify)

    def project_notifications(self, project_id, timeout_seconds=60):
        self._stream_notifications(
            f"projects/{project_id}/notifications", timeout_seconds)

    def project_locked(self, project_id):
        return self._api_call(f"projects/{project_id}/locked", verify=self.verify)

    # Group endpoints
    def groups(self):
        return self._api_call("access/groups", verify=self.verify)

    def groups_by_id(self, group_id):
        return self._api_call(f"access/groups/{group_id}", verify=self.verify)

    def group_members(self, group_id):
        return self._api_call(f"access/groups/{group_id}/members", verify=self.verify)

    # Role endpoints
    def roles(self):
        return self._api_call("access/roles", verify=self.verify)

    def role_by_id(self, role_id):
        return self._api_call(f"access/roles/{role_id}", verify=self.verify)

    def role_privileges(self, role_id):
        return self._api_call(f"access/roles/{role_id}/privileges", verify=self.verify)

    # Privilege endpoints
    def privileges(self):
        return self._api_call("access/privileges", verify=self.verify)

    # ACL endpoints
    def acl_endpoints(self):
        return self._api_call("access/acl/endpoints", verify=self.verify)

    def acl(self):
        return self._api_call("access/acl", verify=self.verify)

    def acl_by_id(self, ace_id):
        return self._api_call(f"access/acl/{ace_id}", verify=self.verify)

    # Template endpoints
    def templates(self):
        return self._api_call("templates", verify=self.verify)

    def template_by_id(self, template_id):
        return self._api_call(f"templates/{template_id}", verify=self.verify)

    # Nodes endpoints
    def nodes(self, project_id):
        return self._api_call(f"projects/{project_id}/nodes", verify=self.verify)

    def node_by_id(self, project_id, node_id):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}", verify=self.verify)

    def node_links_by_id(self, project_id, node_id):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/links", verify=self.verify)

    def node_dynamips_auto_idlepc(self, project_id, node_id):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/dynamips/auto_idlepc", verify=self.verify)

    def node_dynamips_idlepc_proposals(self, project_id, node_id):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/dynamips/idlepc_proposals", verify=self.verify)

    def node_get_file(self, project_id, node_id, file_path):
        return self._api_call(f"projects/{project_id}/nodes/{node_id}/files/{file_path}", verify=self.verify)

    # Link endpoints
    def links(self, project_id):
        return self._api_call(f"projects/{project_id}/links", verify=self.verify)

    def link_filters(self, project_id, link_id):
        return self._api_call(f"projects/{project_id}/links/{link_id}/available_filters", verify=self.verify)

    def link(self, project_id, link_id):
        return self._api_call(f"projects/{project_id}/links/{link_id}", verify=self.verify)

    def link_interface(self, project_id, link_id):
        return self._api_call(f"projects/{project_id}/links/{link_id}", verify=self.verify)

    # Drawing endpoints
    def drawings(self, project_id):
        return self._api_call(f"projects/{project_id}/drawings", verify=self.verify)

    def drawing(self, project_id, drawing_id):
        return self._api_call(f"projects/{project_id}/drawings/{drawing_id}", verify=self.verify)

    # Symbols endpoints
    def symbols(self):
        return self._api_call("symbols", verify=self.verify)

    def symbol(self, symbol_id):
        encoded_symbol_id = urllib.parse.quote(symbol_id, safe='')
        return self._api_call(f"symbols/{encoded_symbol_id}/raw", verify=self.verify)

    def symbol_dimensions(self, symbol_id):
        return self._api_call(f"symbols/{symbol_id}/dimensions", verify=self.verify)

    def default_symbols(self):
        return self._api_call("symbols/default_symbols", verify=self.verify)

    # Snapshot endpoints
    def snapshots(self, project_id):
        return self._api_call(f"projects/{project_id}/snapshots", verify=self.verify)

    # Compute endpoints
    def computes(self):
        return self._api_call("computes", verify=self.verify)

    def compute_by_id(self, compute_id):
        return self._api_call(f"computes/{compute_id}", verify=self.verify)

    def compute_by_id_docker_images(self, compute_id):
        return self._api_call(f"computes/{compute_id}/docker/images", verify=self.verify)

    def compute_by_id_virtualbox_vms(self, compute_id):
        return self._api_call(f"computes/{compute_id}/virtualbox/vms", verify=self.verify)

    def compute_by_id_vmware_vms(self, compute_id):
        return self._api_call(f"computes/{compute_id}/vmware/vms", verify=self.verify)

    # Appliance endpoints
    def appliances(self):
        return self._api_call("appliances", verify=self.verify)

    def appliance(self, appliance_id):
        return self._api_call(f"appliances/{appliance_id}", verify=self.verify)

    # Resource pools endpoints
    def pools(self):
        return self._api_call("pools", verify=self.verify)

    def pool(self, resource_pool_id):
        return self._api_call(f"pools/{resource_pool_id}", verify=self.verify)

    def pool_resources(self, resource_pool_id):
        return self._api_call(f"pools/{resource_pool_id}/resources", verify=self.verify)

    # Images endpoints
    def images(self, image_type):
        return self._api_call(f"images?image_type={image_type}", verify=self.verify)

    def image_by_path(self, image_path):
        return self._api_call(f"images/{image_path}", verify=self.verify)

    # Project file and export methods

    def download_project_file(self, project_id, file_path):
        encoded_file_path = urllib.parse.quote(file_path)
        return self._api_call(f"projects/{project_id}/files/{encoded_file_path}", verify=self.verify)

    def project_export(self, project_id=str, export_params={}):
        """Export a project with the given parameters."""
        params = {
            "include_snapshots": str(export_params["include_snapshots"]).lower(),
            "include_images": str(export_params["include_images"]).lower(),
            "reset_mac_addresses": str(export_params["reset_mac_addresses"]).lower(),
            "keep_compute_ids": str(export_params["keep_compute_ids"]).lower(),
            "compression": export_params["compression"],
            "compression_level": export_params["compression_level"],
        }
        return self._api_call(f"projects/{project_id}/export", stream=True, params=params)

    def download_exported_project(self, project_id=str, export_params={}):
        """Downloads an exported project and saves it to a file."""
        ok, response = self.project_export(project_id, export_params)
        if GNS3Error.has_error(ok):
            GNS3Error.print_error(ok)
            return

        filename = response.headers.get('content-disposition')

        if filename:
            filename = filename.split("filename=")[1].replace('"', '')
        else:
            filename = "exported_project.gns3project"

        with open(filename, "wb") as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)

        click.secho("Success: ", fg="green", nl=False)
        click.secho("Project exported successfully to ", nl=False)
        click.secho(f"{filename}", bold=True)

    def link_capture_stream(self, project_id, link_id, output_file=None, timeout=None):
        """Stream the PCAP capture file from a link."""
        url = f"projects/{project_id}/links/{link_id}/capture/stream"
        success, response = self._api_call(
            url, stream=True, verify=self.verify)
        if success and output_file:
            with open(output_file, 'wb') as f:
                for chunk in response.iter_content(chunk_size=8192):
                    f.write(chunk)
            return True, output_file
        return success, response if success else None
