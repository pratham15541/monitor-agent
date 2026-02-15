#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
version_file="$root_dir/VERSION"

cd "$root_dir"

if ! git diff --quiet; then
  echo "Working tree is dirty. Commit or stash changes first." >&2
  exit 1
fi

bash "$root_dir/scripts/bump_version.sh"
version="$(tr -d ' \t\r\n' < "$version_file")"

bash "$root_dir/scripts/build_agent.sh"

git add VERSION
if ! git diff --cached --quiet; then
  git commit -m "chore: bump version to $version"
fi

git tag "v$version"

echo "Release prepared: v$version"

echo "Next: git push origin main && git push origin v$version"
