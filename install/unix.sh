#!/bin/sh
# install.sh — Install gpf (Greenfield Port Forwarding)
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
TARBALL="gpf_${OS}_${ARCH}.tar.gz"

base_url="https://github.com/${REPO}/releases"
if [ "$VERSION" = "latest" ]; then
  base_url="${base_url}/latest"
else
  base_url="${base_url}/download/v${VERSION}"
fi

TMPDIR="$(mktemp -d 2>/dev/null || mkdir -p /tmp/gpf_install && echo /tmp/gpf_install)"
TMPFILE="${TMPDIR}/gpf"

echo "Installing gpf ${VERSION} for ${OS}/${ARCH}..."

# Try bare binary first, fall back to tar.gz
DOWNLOAD_OK=0
BINARY_URL="${base_url}/download/${BINARY}"
TARBALL_URL="${base_url}/download/${TARBALL}"

if command -v curl >/dev/null 2>&1; then
  if curl -fsSL --progress-bar "$BINARY_URL" -o "$TMPFILE" 2>/dev/null; then
    DOWNLOAD_OK=1
  else
    echo "Downloading tar.gz..."
    TMPTAR="${TMPDIR}/gpf.tar.gz"
    if curl -fsSL --progress-bar "$TARBALL_URL" -o "$TMPTAR" 2>/dev/null; then
      tar xzf "$TMPTAR" -C "$TMPDIR" 2>/dev/null
      # Find the extracted binary
      EXTRACTED=$(find "$TMPDIR" -name "gpf_*" -type f | head -1)
      if [ -n "$EXTRACTED" ]; then
        mv "$EXTRACTED" "$TMPFILE"
        DOWNLOAD_OK=1
      fi
      rm -f "$TMPTAR"
    fi
  fi
elif command -v wget >/dev/null 2>&1; then
  if wget -q -O "$TMPFILE" "$BINARY_URL" 2>/dev/null; then
    DOWNLOAD_OK=1
  else
    echo "Downloading tar.gz..."
    TMPTAR="${TMPDIR}/gpf.tar.gz"
    if wget -q -O "$TMPTAR" "$TARBALL_URL" 2>/dev/null; then
      tar xzf "$TMPTAR" -C "$TMPDIR" 2>/dev/null
      EXTRACTED=$(find "$TMPDIR" -name "gpf_*" -type f | head -1)
      if [ -n "$EXTRACTED" ]; then
        mv "$EXTRACTED" "$TMPFILE"
        DOWNLOAD_OK=1
      fi
      rm -f "$TMPTAR"
    fi
  fi
else
  echo "Error: curl or wget is required to download the binary"
  exit 1
fi

if [ "$DOWNLOAD_OK" != "1" ] || [ ! -s "$TMPFILE" ]; then
  echo "Error: failed to download gpf ${VERSION} for ${OS}/${ARCH}"
  rm -rf "$TMPDIR"
  exit 1
fi

mkdir -p "$INSTALL_DIR"
mv "$TMPFILE" "$INSTALL_DIR/gpf"
chmod +x "$INSTALL_DIR/gpf"
rm -rf "$TMPDIR"

echo "Installed gpf to $INSTALL_DIR/gpf"

# Verify installation
if [ -n "${PATH##*$INSTALL_DIR}" ]; then
  "$INSTALL_DIR/gpf" --version
else
  gpf --version
fi

echo ""
echo "Make sure $INSTALL_DIR is in your PATH."
