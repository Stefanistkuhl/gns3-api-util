#!/usr/bin/env bash
# test-all-scripts.sh
# Usage: ./test-all-scripts.sh <server_url>

if [ $# -ne 1 ]; then
    echo "Usage: $0 <server_url>"
    echo "Example: $0 http://gns3-server:3080"
    exit 1
fi

GNS3_SERVER=$1
TIMESTAMP=$(date +%s)
TEST_PREFIX="TestScript$TIMESTAMP"

# Function to test a script and report results
test_script() {
    local script_name=$1
    local description=$2
    shift 2
    local args=("$@")
    
    echo "--- Testing $script_name: $description ---"
    echo "Command: $script_name ${args[*]}"
    
    if ./scripts/examples/$script_name "${args[@]}" 2>&1; then
        echo "✅ $script_name: PASSED"
    else
        echo "❌ $script_name: FAILED"
    fi
    echo ""
}

# Function to create test class JSON
create_test_class() {
    local class_name=$1
    cat > /tmp/$class_name.json << EOF
{
  "name": "$class_name",
  "groups": [
    {
      "name": "Group1$TIMESTAMP",
      "students": [
        {"username": "student1$TIMESTAMP", "password": "password123"},
        {"username": "student2$TIMESTAMP", "password": "password123"}
      ]
    },
    {
      "name": "Group2$TIMESTAMP", 
      "students": [
        {"username": "student3$TIMESTAMP", "password": "password123"},
        {"username": "student4$TIMESTAMP", "password": "password123"}
      ]
    }
  ]
}
EOF
}

# Function to create test template project
create_test_template() {
    echo "--- Creating test template project ---"
    gns3util -s $GNS3_SERVER project new --name "TestTemplate$TIMESTAMP" --auto-close true
    echo "Template project created: TestTemplate$TIMESTAMP"
    echo ""
}

# Function to clean up test data
cleanup_test_data() {
    echo "--- Cleaning up test data ---"
    
    # Delete test classes
    gns3util -s $GNS3_SERVER class delete --name "TestClass$TIMESTAMP" --no-confirm 2>/dev/null || true
    gns3util -s $GNS3_SERVER class delete --name "TestClass2$TIMESTAMP" --no-confirm 2>/dev/null || true
    gns3util -s $GNS3_SERVER class delete --name "TestClass3$TIMESTAMP" --no-confirm 2>/dev/null || true
    
    # Delete test projects
    gns3util -s $GNS3_SERVER project delete "Student-1-Lab" 2>/dev/null || true
    gns3util -s $GNS3_SERVER project delete "Student-2-Lab" 2>/dev/null || true
    gns3util -s $GNS3_SERVER project delete "TestTemplate$TIMESTAMP" 2>/dev/null || true
    
    # Clean up temp files
    rm -f /tmp/TestClass$TIMESTAMP.json /tmp/TestClass2$TIMESTAMP.json /tmp/TestClass3$TIMESTAMP.json
    
    echo "Cleanup completed"
    echo ""
}

echo "=== Testing All GNS3util Example Scripts ==="
echo "Server: $GNS3_SERVER"
echo "Test Prefix: $TEST_PREFIX"
echo ""

# Clean up any existing test data first
echo "--- Pre-test cleanup ---"
cleanup_test_data

# Create test template project
create_test_template

# Create test classes
echo "--- Creating test classes ---"
create_test_class "TestClass$TIMESTAMP"
create_test_class "TestClass2$TIMESTAMP" 
create_test_class "TestClass3$TIMESTAMP"

# Create the classes
gns3util -s $GNS3_SERVER class create --file /tmp/TestClass$TIMESTAMP.json
gns3util -s $GNS3_SERVER class create --file /tmp/TestClass2$TIMESTAMP.json
gns3util -s $GNS3_SERVER class create --file /tmp/TestClass3$TIMESTAMP.json
echo "Test classes created successfully"
echo ""

# Test 1: Template-based exercise deployment
echo "=== Test 1: Template-based Exercise Deployment ==="
test_script "deploy-template-exercise.sh" "Deploy exercise using existing template" \
    "$GNS3_SERVER" "TestClass$TIMESTAMP" "TestExercise1" "TestTemplate$TIMESTAMP"

# Test 2: Interactive template selection
echo "=== Test 2: Interactive Template Selection ==="
echo "TestTemplate$TIMESTAMP" | test_script "create-exercise-interactive.sh" "Interactive template selection" \
    "$GNS3_SERVER" "TestClass2$TIMESTAMP" "TestExercise2"

# Test 3: File-based template import (if template file exists)
echo "=== Test 3: File-based Template Import ==="
if [ -f "template.gns3project" ]; then
    test_script "import-template-and-create-exercise.sh" "File-based template import" \
        "$GNS3_SERVER" "TestClass3$TIMESTAMP" "TestExercise3" "template.gns3project"
else
    echo "⚠️  Skipping file-based template test (no template.gns3project file found)"
    echo ""
fi

# Test 4: Setup class lab
echo "=== Test 4: Setup Class Lab ==="
echo "admin123" | test_script "setup-class-lab.sh" "Create individual lab projects" \
    "$GNS3_SERVER" "2"

# Test 5: Cleanup script
echo "=== Test 5: Cleanup Script ==="
test_script "cleanup-class.sh" "Clean up projects with prefix" \
    "$GNS3_SERVER" "Student-"

# Final cleanup
cleanup_test_data

echo "=== Test Summary ==="
echo "All script tests completed!"
echo "Check the output above for individual test results."
echo ""
echo "Note: Some scripts may fail due to:"
echo "- Authentication requirements"
echo "- Command syntax differences"
echo "- Missing dependencies (jq, etc.)"
echo "- Existing data conflicts"
