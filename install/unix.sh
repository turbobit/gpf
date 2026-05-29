#!/bin/sh
# install.sh — Install gpf (Greenfield Port Forwarding)
# Inspired by ggh (https://github.com/byawitz/ggh)
# Usage:
#   curl -sSfL https://raw.githubusercontent.com/turbobit/gpf/master/install/unix.sh | sh -s -- v0.1.0
#   ./install/unix.sh v0.1.0

set -e

REPO="turbobit/gpf"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${1:-latest}"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
  linux|darwin) ;;
  *)
    echo "Error: unsupported OS '$OS'"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64)
    ARCH="amd64"
    ;;
  aarch64|arm64)
    ARCH="arm64"
    ;;
  *)
    echo "Error: unsupported architecture '$ARCH'"
    exit 1
    ;;
esac

BINARY="gpf_${OS}_${ARCH}"

if [ "$VERSION" = "latest" ]; then
  DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${BINARY}"
else
  version_clean=$(echo "$VERSION" | sed 's/^v//')
  DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${version_clean}/${BINARY}"
fi

TMPFILE="$(mktemp)"

echo "Installing gpf ${VERSION} for ${OS}/${ARCH}..."
echo "Downloading from: $DOWNLOAD_URL"

if command -v curl >/dev/null 2>&1; then
  curl -fsSL --progress-bar "$DOWNLOAD_URL" -o "$TMPFILE"
elif command -v wget >/dev/null 2>&1; then
  wget -q --show-progress "$DOWNLOAD_URL" -O "$TMPFILE"
else
  echo "Error: curl or wget is required to download the binary"
  exit 1
fi

if [ ! -s "$TMPFILE" ]; then
  echo "Error: failed to download gpf ${VERSION} for ${OS}/${ARCH}"
  rm -f "$TMPFILE"
  exit 1
fi

mkdir -p "$INSTALL_DIR"
mv "$TMPFILE" "$INSTALL_DIR/gpf"
chmod +x "$INSTALL_DIR/gpf"

echo "Installed gpf to $INSTALL_DIR/gpf"

# Verify installation
if [ -n "${PATH##*$INSTALL_DIR}" ]; then
  "$INSTALL_DIR/gpf" --version
else
  gpf --version
fi

echo ""
echo "Make sure $INSTALL_DIR is in your PATH."
