#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
version_file="$root_dir/VERSION"

if [[ ! -f "$version_file" ]]; then
  echo "VERSION file not found at $version_file" >&2
  exit 1
fi

version="$(tr -d ' \t\r\n' < "$version_file")"
if [[ -z "$version" ]]; then
  echo "VERSION is empty" >&2
  exit 1
fi

output="${1:-$root_dir/monitor-agent/monitor-agent}"

cd "$root_dir/monitor-agent"

go build -ldflags "-X monitor-agent/cmd.Version=$version" -o "$output" .

echo "Built $output (version $version)"
