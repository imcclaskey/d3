#!/bin/bash
set -e

# Replace with your GitHub username
GITHUB_USERNAME="imcclaskey"
TAP_REPO_NAME="homebrew-tap"
TAP_REPO_URL="git@github.com:${GITHUB_USERNAME}/${TAP_REPO_NAME}.git"
TAP_DIR="/tmp/${TAP_REPO_NAME}"

# Extract version from version.go
VERSION=$(grep 'const Version = ' internal/version/version.go | sed 's/.*const Version = "\(.*\)".*/\1/')
echo "Found version: $VERSION"

# Ensure our formula is up to date
if [ ! -f "d3.rb" ]; then
  echo "Formula file d3.rb not found."
  exit 1
fi

# Clone or update tap repo
if [ -d "$TAP_DIR" ]; then
  echo "Updating existing tap repository..."
  (cd "$TAP_DIR" && git pull)
else
  echo "Cloning tap repository..."
  git clone "$TAP_REPO_URL" "$TAP_DIR"
fi

# Copy formula to tap repo
echo "Copying formula to tap repository..."
cp d3.rb "$TAP_DIR/"

# Commit and push changes to tap repo
echo "Committing and pushing changes to tap repository..."
(cd "$TAP_DIR" && \
 git add d3.rb && \
 git commit -m "Update d3 to v$VERSION" && \
 git push)

# Clean up temporary directory
echo "Cleaning up temporary directory..."
rm -rf "$TAP_DIR"

echo "Successfully updated homebrew-tap repository with d3 v$VERSION"