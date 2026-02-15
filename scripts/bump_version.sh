#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
version_file="$root_dir/VERSION"

if [[ ! -f "$version_file" ]]; then
  echo "0.1.0" > "$version_file"
  exit 0
fi

version="$(tr -d ' \t\r\n' < "$version_file")"
IFS='.' read -r major minor patch <<< "$version"

if [[ -z "${major:-}" || -z "${minor:-}" || -z "${patch:-}" ]]; then
  echo "0.1.0" > "$version_file"
  exit 0
fi

patch=$((patch + 1))

printf "%s.%s.%s\n" "$major" "$minor" "$patch" > "$version_file"
