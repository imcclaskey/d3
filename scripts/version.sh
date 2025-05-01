#!/bin/bash
# Simple script for version management in Cursor

show_help() {
  echo "Usage: $0 <command>"
  echo ""
  echo "Commands:"
  echo "  view     - Show current version"
  echo "  patch    - Bump patch version (x.y.Z)"
  echo "  minor    - Bump minor version (x.Y.0)"
  echo "  major    - Bump major version (X.0.0)"
  echo "  push     - Push changes and tags to GitHub"
  echo "  help     - Show this help message"
  echo ""
  echo "Example: $0 patch"
}

check_make_available() {
  if ! command -v make &> /dev/null; then
    echo "Error: make command is required"
    exit 1
  fi
}

case "$1" in
  view)
    grep 'const Version =' internal/version/version.go | sed 's/.*"\(.*\)".*/\1/'
    ;;
  patch)
    check_make_available
    make version-patch
    ;;
  minor)
    check_make_available
    make version-minor
    ;;
  major)
    check_make_available
    make version-major
    ;;
  push)
    echo "Pushing changes and tags to GitHub..."
    git push origin main && git push origin --tags
    ;;
  help|*)
    show_help
    ;;
esac 