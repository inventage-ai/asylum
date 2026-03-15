#!/bin/sh
set -e

REPO="heapoutofspace/asylum"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

case "$OS" in
    linux|darwin) ;;
    *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

BINARY="asylum-${OS}-${ARCH}"

if [ -n "$1" ]; then
    VERSION="$1"
else
    VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        echo "Failed to determine latest version" >&2
        exit 1
    fi
fi

URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}"

echo "Downloading asylum ${VERSION} for ${OS}/${ARCH}..."
curl -fsSL -o /tmp/asylum "$URL"
chmod +x /tmp/asylum

if [ -w "$INSTALL_DIR" ]; then
    mv /tmp/asylum "${INSTALL_DIR}/asylum"
else
    echo "Installing to ${INSTALL_DIR} (requires sudo)..."
    sudo mv /tmp/asylum "${INSTALL_DIR}/asylum"
fi

echo "asylum ${VERSION} installed to ${INSTALL_DIR}/asylum"
