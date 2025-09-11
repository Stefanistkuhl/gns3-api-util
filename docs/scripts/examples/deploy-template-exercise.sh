#!/usr/bin/env bash
# deploy-template-exercise.sh
# Usage: ./deploy-template-exercise.sh <server_url> <class_name> <exercise_name> <template_project>

if [ $# -ne 4 ]; then
    echo "Usage: $0 <server_url> <class_name> <exercise_name> <template_project>"
    echo "Example: $0 http://gns3-server:3080 CS101 Lab1 NetworkTemplate"
    exit 1
fi

GNS3_SERVER=$1
CLASS_NAME=$2
EXERCISE_NAME=$3
TEMPLATE_PROJECT=$4

# Create class if it doesn't exist
echo "Creating class: $CLASS_NAME"
# Note: This script assumes the class already exists or is created via JSON file
# Use: gns3util -s $GNS3_SERVER class create --file class.json

# Create exercise using template
echo "Creating exercise '$EXERCISE_NAME' using template '$TEMPLATE_PROJECT'"
gns3util -s $GNS3_SERVER exercise create --class "$CLASS_NAME" --exercise "$EXERCISE_NAME" --template "$TEMPLATE_PROJECT" --confirm=false

echo "Exercise deployment complete!"
