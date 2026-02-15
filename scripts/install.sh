#!/usr/bin/env bash
set -euo pipefail

REPO="${REPO:-pratham15541/monitor-agent}"
INSTALL_DIR="${INSTALL_DIR:-}"

if [[ -z "$INSTALL_DIR" ]]; then
  if [[ -w "/usr/local/bin" ]]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
  fi
fi

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "Unsupported architecture: $arch"; exit 1 ;;
 esac

case "$os" in
  linux) os="linux" ;;
  darwin) os="darwin" ;;
  *) echo "Unsupported OS: $os"; exit 1 ;;
 esac

version="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\(v\{0,1\}[^\"]*\)".*/\1/p' | head -n 1)"
if [[ -z "$version" ]]; then
  echo "Failed to resolve latest release for $REPO"
  exit 1
fi

asset="monitor-agent-${os}-${arch}"
url="https://github.com/$REPO/releases/download/$version/$asset"

curl -fsSL "$url" -o "$INSTALL_DIR/monitor-agent"
chmod +x "$INSTALL_DIR/monitor-agent"

echo "Installed monitor-agent to $INSTALL_DIR/monitor-agent"
