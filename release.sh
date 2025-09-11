#!/bin/bash

# Complete release script for AUR and Homebrew
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check dependencies
if ! command -v jq &> /dev/null || ! command -v curl &> /dev/null; then
    print_error "Missing dependencies: jq, curl"
    exit 1
fi

# Get version
if [ $# -ne 1 ]; then
    print_error "Usage: $0 <version>"
    exit 1
fi

VERSION=$1
REPO="Stefanistkuhl/gns3-api-util"
BASE_URL="https://github.com/${REPO}/releases/download/v${VERSION}"

print_status "Starting release process for version ${VERSION}..."

# Update version in files
print_status "Updating version constants..."
if [ -f "go.mod" ]; then
    sed -i "s/version v[0-9]\+\.[0-9]\+\.[0-9]\+/version v${VERSION}/" go.mod
fi
if [ -f "cmd/root.go" ]; then
    sed -i "s/Version = \"[^\"]*\"/Version = \"${VERSION}\"/" cmd/root.go
fi

# Commit version changes
print_status "Committing version changes..."
git add .
if ! git diff-index --quiet HEAD --; then
    git commit -m "chore: bump version to ${VERSION}"
    print_success "Version changes committed"
else
    print_warning "No version changes to commit"
fi

# Create and push tag
print_status "Creating git tag v${VERSION}..."
if git tag -l | grep -q "^v${VERSION}$"; then
    print_warning "Tag v${VERSION} already exists"
else
    git tag -a "v${VERSION}" -m "Release v${VERSION}"
    print_success "Created tag v${VERSION}"
fi

# Push changes and tags
print_status "Pushing changes and tags..."
if ! git diff-index --quiet HEAD --; then
    git push origin master
fi
git push origin --tags
print_success "Pushed changes and tags"

# Wait for GitHub Actions build
print_status "Waiting for GitHub Actions build to complete..."
print_status "Press Enter when build is finished, or 'q' to quit..."

while true; do
    print_status "Checking build status..."
    
    WORKFLOW_URL="https://api.github.com/repos/${REPO}/actions/workflows/release.yml/runs"
    LATEST_RUN=$(curl -s "${WORKFLOW_URL}" | jq -r '.workflow_runs[0]')
    
    if [ "$LATEST_RUN" = "null" ]; then
        print_warning "No workflow runs found yet..."
    else
        STATUS=$(echo "$LATEST_RUN" | jq -r '.status')
        CONCLUSION=$(echo "$LATEST_RUN" | jq -r '.conclusion')
        
        print_status "Build status: ${STATUS}, conclusion: ${CONCLUSION}"
        
        if [ "$STATUS" = "completed" ]; then
            if [ "$CONCLUSION" = "success" ]; then
                print_success "Build completed successfully!"
                break
            else
                print_error "Build failed with conclusion: ${CONCLUSION}"
                exit 1
            fi
        fi
    fi
    
    print_status "Press Enter to check again, or 'q' to quit: "
    read -r input
    if [ "$input" = "q" ]; then
        print_error "Build check cancelled by user"
        exit 1
    fi
done

# Get checksums
print_status "Fetching checksums..."
DARWIN_AMD64_SHA=$(curl -sL "${BASE_URL}/gns3util-darwin-amd64.tar.gz" | sha256sum | cut -d' ' -f1)
DARWIN_ARM64_SHA=$(curl -sL "${BASE_URL}/gns3util-darwin-arm64.tar.gz" | sha256sum | cut -d' ' -f1)
LINUX_AMD64_SHA=$(curl -sL "${BASE_URL}/gns3util-linux-amd64.tar.gz" | sha256sum | cut -d' ' -f1)
LINUX_ARM64_SHA=$(curl -sL "${BASE_URL}/gns3util-linux-arm64.tar.gz" | sha256sum | cut -d' ' -f1)

print_success "Checksums fetched"

# Update local PKGBUILD files
print_status "Updating local PKGBUILD files..."
if [ -f "PKGBUILD" ]; then
    sed -i "s/pkgver=[^$]*/pkgver=${VERSION}/" "PKGBUILD"
    sed -i "s|sha256sums_x86_64=('[^']*')|sha256sums_x86_64=('${LINUX_AMD64_SHA}')|" "PKGBUILD"
    sed -i "s|sha256sums_aarch64=('[^']*')|sha256sums_aarch64=('${LINUX_ARM64_SHA}')|" "PKGBUILD"
    print_success "Updated local PKGBUILD"
fi

if [ -f "arch/gns3util/PKGBUILD" ]; then
    sed -i "s/pkgver=[^$]*/pkgver=${VERSION}/" "arch/gns3util/PKGBUILD"
    sed -i "s|sha256sums_x86_64=('[^']*')|sha256sums_x86_64=('${LINUX_AMD64_SHA}')|" "arch/gns3util/PKGBUILD"
    sed -i "s|sha256sums_aarch64=('[^']*')|sha256sums_aarch64=('${LINUX_ARM64_SHA}')|" "arch/gns3util/PKGBUILD"
    print_success "Updated local arch/gns3util/PKGBUILD"
fi

# Update Homebrew
print_status "Updating Homebrew tap..."
HOMEBREW_DIR="$HOME/github/homebrew-tap"
if [ -d "$HOMEBREW_DIR" ]; then
    cd "$HOMEBREW_DIR"
    
    # Update formula
    sed -i "s/version \"[^\"]*\"/version \"${VERSION}\"/" "Formula/gns3util.rb"
    sed -i "s|releases/download/v[^/]*/|releases/download/v${VERSION}/|g" "Formula/gns3util.rb"
    sed -i "s/sha256 \"[^\"]*\"/sha256 \"${DARWIN_AMD64_SHA}\"/" "Formula/gns3util.rb"
    sed -i "s/sha256 \"[^\"]*_ARM64\"/sha256 \"${DARWIN_ARM64_SHA}\"/" "Formula/gns3util.rb"
    
    # Commit and push
    git add .
    git commit -m "gns3util ${VERSION}"
    git push origin main
    print_success "Homebrew tap updated"
    
    cd - > /dev/null
else
    print_warning "Homebrew tap directory not found"
fi

# Update AUR
print_status "Updating Arch AUR..."
AUR_DIR="$HOME/github/gns3util/arch/gns3util"
if [ -d "$AUR_DIR" ]; then
    cd "$AUR_DIR"
    
    # Update PKGBUILD
    sed -i "s/pkgver=[^$]*/pkgver=${VERSION}/" "PKGBUILD"
    sed -i "s|sha256sums_x86_64=('[^']*')|sha256sums_x86_64=('${LINUX_AMD64_SHA}')|" "PKGBUILD"
    sed -i "s|sha256sums_aarch64=('[^']*')|sha256sums_aarch64=('${LINUX_ARM64_SHA}')|" "PKGBUILD"
    
    # Generate .SRCINFO
    if command -v makepkg &> /dev/null; then
        makepkg --printsrcinfo > .SRCINFO
    fi
    
    # Commit and push
    git add .
    git commit -m "Update to ${VERSION}"
    git push origin master
    print_success "Arch AUR updated"
    
    cd - > /dev/null
else
    print_warning "Arch AUR directory not found"
fi

print_success "Release ${VERSION} completed!"
print_status "Updated repositories:"
print_status "  - Local PKGBUILD files for local builds"
print_status "  - Homebrew tap: $HOMEBREW_DIR"
print_status "  - Arch AUR: $AUR_DIR"