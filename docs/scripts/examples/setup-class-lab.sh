#!/usr/bin/env bash
# setup-class-lab.sh
# Usage: ./setup-class-lab.sh <server_url> [student_count]

if [ $# -lt 1 ]; then
    echo "Usage: $0 <server_url> [student_count]"
    echo "Example: $0 http://gns3-server:3080 5"
    exit 1
fi

GNS3_SERVER=$1
STUDENT_COUNT=${2:-5}

# Authenticate (uncomment if needed)
# gns3util -s $GNS3_SERVER auth login

# Create class projects
for i in $(seq 1 $STUDENT_COUNT); do
    gns3util -s $GNS3_SERVER project new --name "Student-$i-Lab" --auto-close true
done

# Create basic lab nodes for each student
for i in $(seq 1 $STUDENT_COUNT); do
    PROJECT_NAME="Student-$i-Lab"
    
    # Get project UUID using project info command
    PROJECT_UUID=$(gns3util -s $GNS3_SERVER project info "$PROJECT_NAME" | grep "project_id:" | awk '{print $2}' | tr -d '"')
    
    if [ -z "$PROJECT_UUID" ] || [ "$PROJECT_UUID" = "null" ]; then
        echo "Error: Could not find project UUID for $PROJECT_NAME"
        continue
    fi
    
    echo "Adding nodes to project $PROJECT_NAME (UUID: $PROJECT_UUID)"
    
    # Add routers (using valid node types that don't require additional config)
    gns3util -s $GNS3_SERVER node create "$PROJECT_UUID" --name "R1" --node-type "qemu" --compute-id "local"
    gns3util -s $GNS3_SERVER node create "$PROJECT_UUID" --name "R2" --node-type "qemu" --compute-id "local"
    
    # Add switches
    gns3util -s $GNS3_SERVER node create "$PROJECT_UUID" --name "SW1" --node-type "ethernet_switch" --compute-id "local"
    
    # Add PCs
    gns3util -s $GNS3_SERVER node create "$PROJECT_UUID" --name "PC1" --node-type "vpcs" --compute-id "local"
    gns3util -s $GNS3_SERVER node create "$PROJECT_UUID" --name "PC2" --node-type "vpcs" --compute-id "local"
done
