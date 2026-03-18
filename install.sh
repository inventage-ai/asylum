#!/bin/sh
set -e

REPO="inventage-ai/asylum"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.asylum/bin}"
SYMLINK_DIR="${SYMLINK_DIR:-/usr/local/bin}"

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

mkdir -p "$INSTALL_DIR"
mv /tmp/asylum "${INSTALL_DIR}/asylum"

# Create symlink in a well-known PATH directory.
# Remove any existing file/symlink first to handle legacy installs
# where the binary lived directly in SYMLINK_DIR.
target="${SYMLINK_DIR}/asylum"
if [ -w "$SYMLINK_DIR" ]; then
    rm -f "$target"
    ln -s "${INSTALL_DIR}/asylum" "$target"
elif [ -L "$target" ] || [ -e "$target" ]; then
    echo "Updating symlink in ${SYMLINK_DIR} (requires sudo)..."
    sudo rm -f "$target"
    sudo ln -s "${INSTALL_DIR}/asylum" "$target"
else
    echo "Creating symlink in ${SYMLINK_DIR} (requires sudo)..."
    sudo ln -s "${INSTALL_DIR}/asylum" "$target"
fi

echo "asylum ${VERSION} installed to ${INSTALL_DIR}/asylum"
echo "Symlinked from ${target}"
