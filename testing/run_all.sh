#!/bin/bash
set -e

# Define IDs and parameters
USER_ID="49570cf3-e9a7-4bc2-a1cc-3738937bf232"
GROUP_ID="3e671ebb-3f53-4548-9c34-cf30619544bc"
ROLE_ID="db9c15bf-871d-4526-ba22-2a7e5ea8c5af"
TEMPLATE_ID="e231aa90-d1cf-421b-8b3b-2358ff9d7cc2"
PROJECT_ID="c7acc43b-63fe-4d7c-a149-2e687cb73098"
NODE_ID="bf5e7bcb-e068-4b06-8efa-9611307a12cc"
LINK_ID="a9888151-e1e4-49cf-bbbd-b3d1e76b780c"
DRAWING_ID="e9a246bd-6c74-4602-bfe5-6116f7c9a53a"
SYMBOL_ID="/symbols/affinity/square/green/rj45.svg"
APPLIANCE_ID="3b65c68f-cdde-4dde-a0e7-5ef8c9b7ec2c"
FILE_PATH="12-FUS-TRJ-Übung-6-DHCPv4.gns3"
NODE_FILE="some_file"
CAPTURE_FILE="capture.pcap"

BASE_URL="http://10.21.34.222:3080"

echo "Calling: acl"
gns3util -s "$BASE_URL" get acl

echo "Calling: acl-endpoints"
gns3util -s "$BASE_URL" get acl-endpoints

echo "Calling: acl-rule (using dummy ID)"
gns3util -s "$BASE_URL" get acl-rule 123

echo "Calling: appliance (using APPLIANCE_ID)"
gns3util -s "$BASE_URL" get appliance $APPLIANCE_ID

echo "Calling: appliances"
gns3util -s "$BASE_URL" get appliances

echo "Calling: compute (using dummy ID)"
gns3util -s "$BASE_URL" get compute 123

echo "Calling: computes"
gns3util -s "$BASE_URL" get computes

echo "Calling: default-symbols"
gns3util -s "$BASE_URL" get default-symbols

echo "Calling: docker-images (using dummy compute node ID)"
gns3util -s "$BASE_URL" get docker-images 123

echo "Calling: drawing (using PROJECT_ID and DRAWING_ID)"
gns3util -s "$BASE_URL" get drawing $PROJECT_ID $DRAWING_ID

echo "Calling: drawings (using PROJECT_ID)"
gns3util -s "$BASE_URL" get drawings $PROJECT_ID

echo "Calling: group (using GROUP_ID)"
gns3util -s "$BASE_URL" get group $GROUP_ID

echo "Calling: group-members (using GROUP_ID)"
gns3util -s "$BASE_URL" get group-members $GROUP_ID

echo "Calling: groups"
gns3util -s "$BASE_URL" get groups

echo "Calling: iou-license"
gns3util -s "$BASE_URL" get iou-license

echo "Calling: link (using PROJECT_ID and LINK_ID)"
gns3util -s "$BASE_URL" get link $PROJECT_ID $LINK_ID

echo "Calling: link-filters (using PROJECT_ID and LINK_ID)"
gns3util -s "$BASE_URL" get link-filters $PROJECT_ID $LINK_ID

echo "Calling: links (using PROJECT_ID)"
gns3util -s "$BASE_URL" get links $PROJECT_ID

echo "Calling: me"
gns3util -s "$BASE_URL" get me

echo "Calling: node (using PROJECT_ID and NODE_ID)"
gns3util -s "$BASE_URL" get node $PROJECT_ID $NODE_ID

echo "Calling: node-links (using PROJECT_ID and NODE_ID)"
gns3util -s "$BASE_URL" get node-links $PROJECT_ID $NODE_ID

echo "Calling: nodes (using PROJECT_ID)"
gns3util -s "$BASE_URL" get nodes $PROJECT_ID

echo "Calling: notifications"
gns3util -s "$BASE_URL" get notifications --timeout 10

echo "Calling: pool (using dummy ID)"
gns3util -s "$BASE_URL" get pool 123

echo "Calling: pool-resources (using dummy pool ID)"
gns3util -s "$BASE_URL" get pool-resources 123

echo "Calling: pools"
gns3util -s "$BASE_URL" get pools

echo "Calling: privileges"
gns3util -s "$BASE_URL" get privileges

echo "Calling: project (using PROJECT_ID)"
gns3util -s "$BASE_URL" get project $PROJECT_ID

echo "Calling: project-locked (using PROJECT_ID)"
gns3util -s "$BASE_URL" get project-locked $PROJECT_ID

echo "Calling: project-notifications (using PROJECT_ID)"
gns3util -s "$BASE_URL" get project-notifications $PROJECT_ID

echo "Calling: project-stats (using PROJECT_ID)"
gns3util -s "$BASE_URL" get project-stats $PROJECT_ID

echo "Calling: projects"
gns3util -s "$BASE_URL" get projects

echo "Calling: role (using ROLE_ID)"
gns3util -s "$BASE_URL" get role $ROLE_ID

echo "Calling: role-privileges (using ROLE_ID)"
gns3util -s "$BASE_URL" get role-privileges $ROLE_ID

echo "Calling: roles"
gns3util -s "$BASE_URL" get roles

echo "Calling: statistics"
gns3util -s "$BASE_URL" get statistics

echo "Calling: symbol (using SYMBOL_ID)"
gns3util -s "$BASE_URL" get symbol $SYMBOL_ID

echo "Calling: symbols"
gns3util -s "$BASE_URL" get symbols

echo "Calling: template (using TEMPLATE_ID)"
gns3util -s "$BASE_URL" get template $TEMPLATE_ID

echo "Calling: templates"
gns3util -s "$BASE_URL" get templates

echo "Calling: user (using USER_ID)"
gns3util -s "$BASE_URL" get user $USER_ID

echo "Calling: user-groups (using USER_ID)"
gns3util -s "$BASE_URL" get user-groups $USER_ID

echo "Calling: users"
gns3util -s "$BASE_URL" get users

echo "Calling: version"
gns3util -s "$BASE_URL" get version

echo "Calling: virtualbox-vms (using dummy compute node ID)"
gns3util -s "$BASE_URL" get virtualbox-vms 123

echo "Calling: vmware-vms (using dummy compute node ID)"
gns3util -s "$BASE_URL" get vmware-vms 123

echo "All endpoints have been called."
