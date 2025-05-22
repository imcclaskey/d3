#!/bin/bash
set -e

# Replace with your GitHub username
GITHUB_USERNAME="imcclaskey"
TAP_REPO_NAME="homebrew-tap"
# Use HTTPS instead of SSH to avoid authentication prompts
TAP_REPO_URL="https://github.com/${GITHUB_USERNAME}/${TAP_REPO_NAME}.git"
TAP_DIR="/tmp/${TAP_REPO_NAME}"

# Extract version from version.go
VERSION=$(grep 'const Version = ' internal/version/version.go | sed 's/.*const Version = "\(.*\)".*/\1/')
echo "Found version: $VERSION"

# Ensure our formula is up to date
if [ ! -f "d3.rb" ]; then
  echo "Formula file d3.rb not found."
  exit 1
fi

# Clean up any existing tap directory to ensure we start fresh
if [ -d "$TAP_DIR" ]; then
  echo "Removing existing tap directory..."
  rm -rf "$TAP_DIR"
fi

# Clone the tap repository
echo "Cloning tap repository..."
if ! git clone "$TAP_REPO_URL" "$TAP_DIR" 2>/dev/null; then
  echo "Error: Could not clone the tap repository."
  echo "Please make sure you have created the 'homebrew-tap' repository on GitHub at:"
  echo "https://github.com/$GITHUB_USERNAME/homebrew-tap"
  echo ""
  echo "To create it:"
  echo "1. Go to https://github.com/new"
  echo "2. Name it 'homebrew-tap'"
  echo "3. Make it public"
  echo "4. Initialize with a README"
  exit 1
fi

# Copy formula to tap repo
echo "Copying formula to tap repository..."
cp d3.rb "$TAP_DIR/"

# Commit and push changes to tap repo
echo "Committing and pushing changes to tap repository..."
(cd "$TAP_DIR" && \
 git add d3.rb && \
 git commit -m "Update d3 to v$VERSION")

# Push changes (handle authentication)
echo "Pushing changes to GitHub..."
echo "You may be prompted for your GitHub username and password/token."
echo "If you have 2FA enabled, use a personal access token as the password."
(cd "$TAP_DIR" && git push)

# Clean up temporary directory
echo "Cleaning up temporary directory..."
rm -rf "$TAP_DIR"

echo "Successfully updated homebrew-tap repository with d3 v$VERSION"