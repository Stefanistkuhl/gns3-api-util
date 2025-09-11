#!/bin/bash

# Generate shell completion files
set -e

echo "Generating shell completion files..."

# Create completions directory
mkdir -p completions

# Build the binary
go build -o gns3util .

# Generate completions
echo "Generating bash completion..."
./gns3util completion bash > completions/gns3util.bash

echo "Generating zsh completion..."
./gns3util completion zsh > completions/_gns3util

echo "Generating fish completion..."
./gns3util completion fish > completions/gns3util.fish

echo "Generating PowerShell completion..."
./gns3util completion powershell > completions/gns3util.ps1

echo "Completion files generated in completions/ directory"
ls -la completions/
