#!/bin/bash
set -e

# Extract version from version.go (macOS compatible)
VERSION=$(grep 'const Version = ' internal/version/version.go | sed 's/.*const Version = "\(.*\)".*/\1/')
echo "Found version: $VERSION"

# Update version in d3.rb
sed -i.bak "s/version \"[^\"]*\"/version \"$VERSION\"/" d3.rb

# Remove backup file
rm d3.rb.bak

# Calculate SHA256 checksums for each binary (assumes binaries are built)
echo "Calculating SHA256 checksums for binaries..."
DARWIN_AMD64=$(shasum -a 256 build/d3-darwin-amd64 | cut -d ' ' -f 1)
DARWIN_ARM64=$(shasum -a 256 build/d3-darwin-arm64 | cut -d ' ' -f 1)
LINUX_AMD64=$(shasum -a 256 build/d3-linux-amd64 | cut -d ' ' -f 1)
LINUX_ARM64=$(shasum -a 256 build/d3-linux-arm64 | cut -d ' ' -f 1)

# Update SHA256 checksums in d3.rb
sed -i.bak "s/sha256 \"REPLACE_WITH_SHA256\" # darwin-arm64/sha256 \"$DARWIN_ARM64\" # darwin-arm64/" d3.rb
sed -i.bak "s/sha256 \"REPLACE_WITH_SHA256\" # darwin-amd64/sha256 \"$DARWIN_AMD64\" # darwin-amd64/" d3.rb
sed -i.bak "s/sha256 \"REPLACE_WITH_SHA256\" # linux-arm64/sha256 \"$LINUX_ARM64\" # linux-arm64/" d3.rb
sed -i.bak "s/sha256 \"REPLACE_WITH_SHA256\" # linux-amd64/sha256 \"$LINUX_AMD64\" # linux-amd64/" d3.rb

# Remove backup file
rm d3.rb.bak

echo "Formula updated with version $VERSION and SHA256 checksums" 