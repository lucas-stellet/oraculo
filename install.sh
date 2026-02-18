#!/usr/bin/env bash
set -euo pipefail

if [ $# -ne 1 ]; then
  echo "Usage: $0 <target-path>"
  echo "Example: $0 /path/to/my-project"
  exit 1
fi

TARGET="$1"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SOURCE="$SCRIPT_DIR/claude-kit"

if [ ! -d "$SOURCE" ]; then
  echo "Error: claude-kit directory not found at $SOURCE"
  exit 1
fi

DEST="$TARGET/.claude/skills"

if [ -d "$DEST" ]; then
  echo "Removing old skills at $DEST"
  rm -rf "$DEST"
fi

mkdir -p "$DEST"
cp -R "$SOURCE"/skills/ "$DEST"/

echo "Installed oraculo skills to $DEST"
