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

version_clean=$(echo "$VERSION" | sed 's/^v//')

base_url="https://github.com/${REPO}/releases"
if [ "$VERSION" = "latest" ]; then
  base_url="${base_url}/latest"
else
  base_url="${base_url}/download/v${VERSION}"
fi

TMPDIR="$(mktemp -d 2>/dev/null || mkdir -p /tmp/gpf_install && echo /tmp/gpf_install)"
TMPFILE="${TMPDIR}/gpf"

echo "Installing gpf ${VERSION} for ${OS}/${ARCH}..."

# Try various download URLs in order of preference
DOWNLOAD_OK=0
URLS_TO_TRY=""

if [ "$VERSION" != "latest" ]; then
  URLS_TO_TRY="${base_url}/download/gpf_${version_clean}_${OS}_${ARCH}"
  URLS_TO_TRY="${URLS_TO_TRY}
${base_url}/download/gpf_${version_clean}_${OS}_${ARCH}.tar.gz"
fi

URLS_TO_TRY="${URLS_TO_TRY}
${base_url}/download/gpf_${OS}_${ARCH}
${base_url}/download/gpf_${OS}_${ARCH}.tar.gz"

TMPTAR="${TMPDIR}/gpf.tar.gz"

for url in $URLS_TO_TRY; do
  echo "Trying: $url"

  # Try downloading bare binary
  if command -v curl >/dev/null 2>&1; then
    if curl -fsSL --progress-bar "$url" -o "$TMPFILE" 2>/dev/null; then
      if [ -s "$TMPFILE" ]; then
        DOWNLOAD_OK=1
        break
      fi
    fi
  elif command -v wget >/dev/null 2>&1; then
    if wget -q -O "$TMPFILE" "$url" 2>/dev/null; then
      if [ -s "$TMPFILE" ]; then
        DOWNLOAD_OK=1
        break
      fi
    fi
  fi

  # Try downloading tar.gz
  is_tar=0
  case "$url" in
    *.tar.gz) is_tar=1 ;;
  esac

  if [ "$is_tar" = "1" ]; then
    if command -v curl >/dev/null 2>&1; then
      if curl -fsSL --progress-bar "$url" -o "$TMPTAR" 2>/dev/null; then
        tar xzf "$TMPTAR" -C "$TMPDIR" 2>/dev/null
        EXTRACTED=$(find "$TMPDIR" -name "gpf*" -type f ! -name "*.tar.gz" ! -name "*.txt" | head -1)
        if [ -n "$EXTRACTED" ] && [ -s "$EXTRACTED" ]; then
          mv "$EXTRACTED" "$TMPFILE"
          DOWNLOAD_OK=1
          break
        fi
      fi
    elif command -v wget >/dev/null 2>&1; then
      if wget -q -O "$TMPTAR" "$url" 2>/dev/null; then
        tar xzf "$TMPTAR" -C "$TMPDIR" 2>/dev/null
        EXTRACTED=$(find "$TMPDIR" -name "gpf*" -type f ! -name "*.tar.gz" ! -name "*.txt" | head -1)
        if [ -n "$EXTRACTED" ] && [ -s "$EXTRACTED" ]; then
          mv "$EXTRACTED" "$TMPFILE"
          DOWNLOAD_OK=1
          break
        fi
      fi
    fi
  fi
done

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
