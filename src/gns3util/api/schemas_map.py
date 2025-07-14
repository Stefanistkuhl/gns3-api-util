from typing import Any
from gns3util.schemas import (
    Version,
    IOULicense,
    Token,
    Credentials,
    User,
    LoggedInUserUpdate,
    UserGroup,
    Role,
    Privilege,
    ACE,
    Image,
    Template,
    Project,
    Node,
    Link,
    UDPPortInfo,
    EthernetPortInfo,
    Drawing,
    Snapshot,
    Compute,
    ComputeDockerImage,
    ComputeVirtualBoxVM,
    ComputeVMwareVM,
    AutoIdlePC,
    Appliance,
    ResourcePool,
    Resource,
)

RESPONSE_SCHEMA_MAP = {
    # get
    # controller
    "version": Version,
    "iou_license": IOULicense,
    "statistics": list[dict],
    "notifications": Any,

    # users
    "user_authenticate": Token,
    "curret_user_info": User,
    "me": User,
    "users": list[User],
    "create_user": User,
    "user": User,
    "update_user": User,
    "delete_user": None,

    # groups
    "users_groups": list[UserGroup],
    "groups": list[UserGroup],
    "create_group": UserGroup,
    "group_by_id": UserGroup,
    "update_group": UserGroup,
    "delete_group": None,
    "group_members": list[User],
    "add_group_member": None,
    "delete_user_from_group": None,

    # roles
    "roles": list[Role],
    "create_role": Role,
    "role_by_id": Role,
    "update_role": Role,
    "delete_role": None,
    "role_privileges": list[Privilege],
    "update_role_privs": None,
    "delete_role_priv": None,

    # privs
    "privileges": list[Privilege],

    # acl
    # this endpoint should have a custom class with the schema and for code that creates acls it should call this first to see if the endpoint is availiable
    "acl_endpoints": list[dict],
    "acl": list[ACE],
    "create_acl": ACE,
    "acl_by_id": ACE,
    "update_ace": ACE,
    "delete_ace": None,

    # images
    "create_qemu_image": Image,
    "images": list[Image],
    "upload_image": Image,
    "prune_images": None,
    "install_image": None,
    "image_by_path": Image,
    "delete_image": None,

    # Templates
    "templates": list[Template],
    "create_template": Template,
    "template_by_id": Template,
    "update_template": Template,
    "delete_endpoint": None,
    "duplicate_template": Template,

    # Projects
    "projects": Project,
    "create_project": Project,
    "project": Project,
    "update_project": Project,
    "delete_project": None,
    "project_stats": dict,
    "close_project": None,
    "open_project": Project,
    "load_project": Project,
    "project_notifications": Any,
    "download_exported_project": Any,
    "import_project": Project,
    "duplicate_project": Project,
    "project_locked": bool,
    "lock_project": None,
    "unlock_project": None,
    "download_project_file": Any,
    "write_project_file": None,
    "create_project_node_from_template": Node,

    # nodes
    "create_node": Node,
    "nodes": list[Node],
    "start_nodes": None,
    "stop_nodes": None,
    "suspend_nodes": None,
    "reload_nodes": None,
    "node_by_id": Node,
    "update_node": Node,
    "delete_node": None,
    "duplicate_node": Node,
    "start_node": None,
    "stop_node": None,
    "suspend_node": None,
    "reload_node": None,
    "isolate_node": None,
    "unisolate_node": None,
    "node_links_by_id": list[Link],
    # figure out what to do with this
    "node_dynamips_audo_idlepc": Any,
    "node_dynamips_audo_idlepc_proposals": list[str],
    "create_disk_image": None,
    "update_disk_image": None,
    "delete_disk_image": None,
    "node_get_file": Any,
    "create_node_file": Any,
    "reset_nodes_console": None,
    "reset_node_console": None,

    # links
    "links": Link,
    "create_link": Link,
    "link_filters": list[dict],
    "link": Link,
    "update_link": Link,
    "delete_link": None,
    "reset_link": Link,
    "start_link_capture": Link,
    "stop_link_capture": None,
    "link_capture_stream": Any,
    "link_interface": UDPPortInfo | EthernetPortInfo,

    # drawings
    "drawings": list[Drawing],
    "create_drawing": Drawing,
    "drawing": Drawing,
    "update_drawing": Drawing,
    "delete_drawing": None,

    # symbols
    "symbols": list[dict],
    "symbol": Any,
    "create_symbol": None,
    "symbol_dimensions": dict,
    "default_symbols": dict,

    # snapshots
    "create_snapshot": Snapshot,
    "snapshots": list[Snapshot],
    "delete_snapshot": None,
    "restore_snapshot": Project,

    # compute
    "create_compute": Compute,
    "computes": list[Compute],
    "connect_compute": None,
    "compute_by_id": Compute,
    "update_compute": Compute,
    "delete_compute": None,
    "compute_by_id_docker_images": list[ComputeDockerImage],
    "compute_by_id_virtualbox_vms": list[ComputeVirtualBoxVM],
    "compute_by_id_vmware_vms": list[ComputeVirtualBoxVM],
    "set_auto_idlepc": AutoIdlePC,

    # appliances
    "appliances": list[Appliance],
    "appliance": Appliance,
    "create_appliance_version": dict,
    "install_appliance_version": None,

    # ressource pools
    "pools": ResourcePool,
    "create_pool": ResourcePool,
    "pool": ResourcePool,
    "update_pool": ResourcePool,
    "delete_pool": None,
    "add_resource_to_pool": None,
    "delete_pool_resource": None,
}
