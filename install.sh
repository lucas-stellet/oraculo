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
SKILLS_SRC="$SOURCE/skills"

# Find all directories containing a SKILL.md and copy them flat into .claude/skills/
# e.g. oraculo/epic/ → oraculo-epic/
find "$SKILLS_SRC" -name "SKILL.md" -print0 | while IFS= read -r -d '' skill_file; do
  skill_dir="$(dirname "$skill_file")"
  rel_path="${skill_dir#"$SKILLS_SRC"/}"
  flat_name="${rel_path//\//-}"
  dest_dir="$DEST/$flat_name"

  if [ -d "$dest_dir" ]; then
    echo "Updating $flat_name"
    rm -rf "$dest_dir"
  else
    echo "Installing $flat_name"
  fi

  mkdir -p "$dest_dir"
  cp -R "$skill_dir"/ "$dest_dir"/
done

echo "Installed oraculo skills to $DEST"
