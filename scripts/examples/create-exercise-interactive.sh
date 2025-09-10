#!/usr/bin/env bash
# create-exercise-interactive.sh
# Usage: ./create-exercise-interactive.sh <server_url> <class_name> <exercise_name>

if [ $# -ne 3 ]; then
    echo "Usage: $0 <server_url> <class_name> <exercise_name>"
    echo "Example: $0 http://gns3-server:3080 CS101 Lab1"
    exit 1
fi

GNS3_SERVER=$1
CLASS_NAME=$2
EXERCISE_NAME=$3

# Create class
echo "Creating class: $CLASS_NAME"
# Note: This script assumes the class already exists or is created via JSON file
# Use: ./gns3util -s $GNS3_SERVER class create --file class.json

# Create exercise with interactive template selection
echo "Creating exercise with template selection..."
./gns3util -s $GNS3_SERVER exercise create --class "$CLASS_NAME" --exercise "$EXERCISE_NAME" --select-template --confirm=false

echo "Exercise created with selected template!"
