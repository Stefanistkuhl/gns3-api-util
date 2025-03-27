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

# Color definitions
GREEN=$(tput setaf 2)
RED=$(tput setaf 1)
NC=$(tput sgr0)
# Retry settings
RETRY_COUNT=2

# Function: call endpoint with retry and display colored output
call_endpoint_with_retry() {
    local desc="$1"
    shift
    local attempt=1
    local exitcode=0

    echo "Calling: ${desc}"

    while [ $attempt -le $((RETRY_COUNT+1)) ]; do
        if gns3util -s "$BASE_URL" get "$@"; then
            echo "${GREEN}${desc}: Success (attempt ${attempt})${NC}"
            return 0
        else
            echo "${RED}${desc}: Failed (attempt ${attempt})${NC}"
            exitcode=1
        fi
        attempt=$((attempt+1))
        sleep 1
    done
    return $exitcode
}

# A list to record endpoints statuses
declare -A endpoint_status

# Helper function to call an endpoint with a description and record status.
call_and_record() {
    local desc="$1"
    shift
    if call_endpoint_with_retry "$desc" "$@"; then
        endpoint_status["$desc"]="Success"
    else
        endpoint_status["$desc"]="Failed"
    fi
}

call_and_record "acl" "acl"
call_and_record "acl-endpoints" "acl-endpoints"
call_and_record "acl-rule (using dummy ID)" "acl-rule" "123"
call_and_record "appliance (using APPLIANCE_ID)" "appliance" "$APPLIANCE_ID"
call_and_record "appliances" "appliances"
call_and_record "compute (using dummy ID)" "compute" "123"
call_and_record "computes" "computes"
call_and_record "default-symbols" "default-symbols"
call_and_record "docker-images (using dummy compute node ID)" "docker-images" "123"
call_and_record "drawing (using PROJECT_ID and DRAWING_ID)" "drawing" "$PROJECT_ID" "$DRAWING_ID"
call_and_record "drawings (using PROJECT_ID)" "drawings" "$PROJECT_ID"
call_and_record "group (using GROUP_ID)" "group" "$GROUP_ID"
call_and_record "group-members (using GROUP_ID)" "group-members" "$GROUP_ID"
call_and_record "groups" "groups"
call_and_record "iou-license" "iou-license"
call_and_record "link (using PROJECT_ID and LINK_ID)" "link" "$PROJECT_ID" "$LINK_ID"
call_and_record "link-filters (using PROJECT_ID and LINK_ID)" "link-filters" "$PROJECT_ID" "$LINK_ID"
call_and_record "links (using PROJECT_ID)" "links" "$PROJECT_ID"
call_and_record "me" "me"
call_and_record "node (using PROJECT_ID and NODE_ID)" "node" "$PROJECT_ID" "$NODE_ID"
call_and_record "node-links (using PROJECT_ID and NODE_ID)" "node-links" "$PROJECT_ID" "$NODE_ID"
call_and_record "nodes (using PROJECT_ID)" "nodes" "$PROJECT_ID"
call_and_record "notifications" "notifications" --timeout "10"
call_and_record "pool (using dummy ID)" "pool" "123"
call_and_record "pool-resources (using dummy pool ID)" "pool-resources" "123"
call_and_record "pools" "pools"
call_and_record "privileges" "privileges"
call_and_record "project (using PROJECT_ID)" "project" "$PROJECT_ID"
call_and_record "project-locked (using PROJECT_ID)" "project-locked" "$PROJECT_ID"
call_and_record "project-notifications (using PROJECT_ID)" "project-notifications" "$PROJECT_ID"
call_and_record "project-stats (using PROJECT_ID)" "project-stats" "$PROJECT_ID"
call_and_record "projects" "projects"
call_and_record "role (using ROLE_ID)" "role" "$ROLE_ID"
call_and_record "role-privileges (using ROLE_ID)" "role-privileges" "$ROLE_ID"
call_and_record "roles" "roles"
call_and_record "statistics" "statistics"
call_and_record "symbol (using SYMBOL_ID)" "symbol" "$SYMBOL_ID"
call_and_record "symbols" "symbols"
call_and_record "template (using TEMPLATE_ID)" "template" "$TEMPLATE_ID"
call_and_record "template (using
call_and_record "user (using USER_ID)" "user" "$USER_ID"
call_and_record "user-groups (using USER_ID)" "user-groups" "$USER_ID"
call_and_record "users" "users"
call_and_record "version" "version"
call_and_record "virtualbox-vms (using dummy compute node ID)" "virtualbox-vms" "123"
call_and_record "vmware-vms (using dummy compute node ID)" "vmware-vms" "123"

# Print summary of endpoint statuses
echo "Summary of endpoint calls:"
for key in "${!endpoint_status[@]}"; do
  echo "${key}: ${endpoint_status[$key]}"
done
