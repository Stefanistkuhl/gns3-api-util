#!/bin/bash

# Script to update package files with new release information
set -e

if [ $# -ne 1 ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 1.0.2"
    exit 1
fi

VERSION=$1
REPO="Stefanistkuhl/gns3-api-util"
BASE_URL="https://github.com/${REPO}/releases/download/v${VERSION}"

echo "Updating package files for version ${VERSION}..."

# Function to get SHA256 checksum
get_sha256() {
    local url=$1
    echo "Fetching SHA256 for ${url}..."
    curl -sL "${url}" | sha256sum | cut -d' ' -f1
}

# Update Homebrew formula
echo "Updating Homebrew formula..."
HOMEBREW_FILE="homebrew-tap/Formula/gns3util.rb"

# Get checksums
AMD64_SHA=$(get_sha256 "${BASE_URL}/gns3util-darwin-amd64.tar.gz")
ARM64_SHA=$(get_sha256 "${BASE_URL}/gns3util-darwin-arm64.tar.gz")

# Update the formula
sed -i "s/version \"[^\"]*\"/version \"${VERSION}\"/" "${HOMEBREW_FILE}"
sed -i "s|releases/download/v[^/]*/|releases/download/v${VERSION}/|g" "${HOMEBREW_FILE}"
sed -i "s/sha256 \"[^\"]*\"/sha256 \"${AMD64_SHA}\"/" "${HOMEBREW_FILE}"
sed -i "s/sha256 \"[^\"]*_ARM64\"/sha256 \"${ARM64_SHA}\"/" "${HOMEBREW_FILE}"

echo "Updated Homebrew formula: ${HOMEBREW_FILE}"

# Update Arch PKGBUILD
echo "Updating Arch PKGBUILD..."
ARCH_FILE="arch/gns3util/PKGBUILD"

# Get Linux checksums
LINUX_AMD64_SHA=$(get_sha256 "${BASE_URL}/gns3util-linux-amd64.tar.gz")
LINUX_ARM64_SHA=$(get_sha256 "${BASE_URL}/gns3util-linux-arm64.tar.gz")

# Update the PKGBUILD
sed -i "s/pkgver=[^$]*/pkgver=${VERSION}/" "${ARCH_FILE}"
sed -i "s/sha256sums_x86_64=('[^']*')/sha256sums_x86_64=('${LINUX_AMD64_SHA}')/" "${ARCH_FILE}"
sed -i "s/sha256sums_aarch64=('[^']*')/sha256sums_aarch64=('${LINUX_ARM64_SHA}')/" "${ARCH_FILE}"

echo "Updated Arch PKGBUILD: ${ARCH_FILE}"

echo "✅ Package files updated for version ${VERSION}"
echo ""
echo "Next steps:"
echo "1. Commit and push the changes:"
echo "   git add . && git commit -m \"chore: update package files for v${VERSION}\" && git push"
echo ""
echo "2. For Homebrew tap:"
echo "   cd homebrew-tap && git add . && git commit -m \"gns3util ${VERSION}\" && git push"
echo ""
echo "3. For Arch Linux AUR:"
echo "   cd arch/gns3util && makepkg --printsrcinfo > .SRCINFO"
echo "   git add . && git commit -m \"Update to ${VERSION}\" && git push"
