#!/bin/sh
# install.sh — Install gpf (Greenfield Port Forwarding)
# Usage:
#   curl -sSfL https://raw.githubusercontent.com/user/port-forwarding/main/install/unix.sh | sh -s -- v0.1.0
#   ./install/unix.sh v0.1.0

set -e

REPO="user/port-forwarding"
INSTALL_DIR="$HOME/.local/bin"
VERSION="${1#v}"

if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 v0.1.0"
  exit 1
fi

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
TMPFILE="$(mktemp)/gpf"

echo "Installing gpf $VERSION for $OS/$ARCH..."

curl --location --progress-bar "https://github.com/${REPO}/releases/download/v${VERSION}/${BINARY}" -o "$TMPFILE"

mkdir -p "$INSTALL_DIR"
mv "$TMPFILE" "$INSTALL_DIR/gpf"
chmod +x "$INSTALL_DIR/gpf"

echo "Installed gpf to $INSTALL_DIR/gpf"
echo "Make sure $INSTALL_DIR is in your PATH."
