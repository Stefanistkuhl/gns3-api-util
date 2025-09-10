#!/usr/bin/env bash
# cleanup-class.sh
# Usage: ./cleanup-class.sh <server_url> [project_prefix]

if [ $# -lt 1 ]; then
    echo "Usage: $0 <server_url> [project_prefix]"
    echo "Example: $0 http://gns3-server:3080 Student-"
    exit 1
fi

GNS3_SERVER=$1
PROJECT_PREFIX=${2:-"Student-"}

# Get all projects with the specified prefix and delete them
./gns3util -s $GNS3_SERVER project ls | grep "name:" | grep "$PROJECT_PREFIX" | awk '{print $2}' | tr -d '"' | while read project; do
    if [ -n "$project" ] && [ "$project" != "null" ]; then
        echo "Deleting project: $project"
        ./gns3util -s $GNS3_SERVER project delete "$project"
    fi
done
