#!/bin/bash
set -e

# Extract version from version.go
VERSION=$(grep 'const Version = ' internal/version/version.go | sed 's/.*const Version = "\(.*\)".*/\1/')
echo "Found version: $VERSION"

# Ensure binaries are built to calculate checksums
if [ ! -f "build/d3-darwin-amd64" ]; then
  echo "Binaries not found. Run 'make build-all' first."
  exit 1
fi

# Calculate SHA256 checksums for each binary
echo "Calculating SHA256 checksums for binaries..."
DARWIN_AMD64=$(shasum -a 256 build/d3-darwin-amd64 | cut -d ' ' -f 1)
DARWIN_ARM64=$(shasum -a 256 build/d3-darwin-arm64 | cut -d ' ' -f 1)
LINUX_AMD64=$(shasum -a 256 build/d3-linux-amd64 | cut -d ' ' -f 1)
LINUX_ARM64=$(shasum -a 256 build/d3-linux-arm64 | cut -d ' ' -f 1)

# Update version in d3.rb
echo "Updating formula with version $VERSION..."
sed -i.bak "s/version \"[^\"]*\"/version \"$VERSION\"/" d3.rb
rm d3.rb.bak

# Update SHA256 checksums in d3.rb
echo "Updating SHA256 checksums..."
sed -i.bak "s/sha256 \"[^\"]*\" # darwin-arm64/sha256 \"$DARWIN_ARM64\" # darwin-arm64/" d3.rb
sed -i.bak "s/sha256 \"[^\"]*\" # darwin-amd64/sha256 \"$DARWIN_AMD64\" # darwin-amd64/" d3.rb
sed -i.bak "s/sha256 \"[^\"]*\" # linux-arm64/sha256 \"$LINUX_ARM64\" # linux-arm64/" d3.rb
sed -i.bak "s/sha256 \"[^\"]*\" # linux-amd64/sha256 \"$LINUX_AMD64\" # linux-amd64/" d3.rb
rm d3.rb.bak

echo "Formula updated with version $VERSION and SHA256 checksums" 