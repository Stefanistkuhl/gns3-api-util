#!/usr/bin/env bash
# import-template-and-create-exercise.sh
# Usage: ./import-template-and-create-exercise.sh <server_url> <class_name> <exercise_name> <template_file>

if [ $# -ne 4 ]; then
    echo "Usage: $0 <server_url> <class_name> <exercise_name> <template_file>"
    echo "Example: $0 http://gns3-server:3080 CS101 Lab1 /path/to/template.gns3project"
    exit 1
fi

GNS3_SERVER=$1
CLASS_NAME=$2
EXERCISE_NAME=$3
TEMPLATE_FILE=$4

# Create class
echo "Creating class: $CLASS_NAME"
# Note: This script assumes the class already exists or is created via JSON file
# Use: ./gns3util -s $GNS3_SERVER class create --file class.json

# Create exercise using file template
echo "Creating exercise '$EXERCISE_NAME' using template file '$TEMPLATE_FILE'"
./gns3util -s $GNS3_SERVER exercise create --class "$CLASS_NAME" --exercise "$EXERCISE_NAME" --template "$TEMPLATE_FILE" --confirm=false

echo "Exercise created from template file!"
