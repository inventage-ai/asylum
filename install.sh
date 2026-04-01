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

CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

echo "Downloading asylum ${VERSION} for ${OS}/${ARCH}..."
curl -fsSL -o /tmp/asylum "$URL"
chmod +x /tmp/asylum

# Verify checksum if checksums.txt is available
if curl -fsSL -o /tmp/asylum-checksums.txt "$CHECKSUMS_URL" 2>/dev/null; then
    EXPECTED=$(grep "${BINARY}$" /tmp/asylum-checksums.txt | awk '{print $1}')
    if [ -n "$EXPECTED" ]; then
        if command -v sha256sum >/dev/null 2>&1; then
            ACTUAL=$(sha256sum /tmp/asylum | awk '{print $1}')
        elif command -v shasum >/dev/null 2>&1; then
            ACTUAL=$(shasum -a 256 /tmp/asylum | awk '{print $1}')
        else
            echo "Warning: no sha256sum or shasum available, skipping checksum verification"
            ACTUAL=""
        fi
        if [ -n "$ACTUAL" ] && [ "$ACTUAL" != "$EXPECTED" ]; then
            echo "Checksum verification failed!" >&2
            echo "  Expected: $EXPECTED" >&2
            echo "  Actual:   $ACTUAL" >&2
            rm -f /tmp/asylum /tmp/asylum-checksums.txt
            exit 1
        fi
    fi
    rm -f /tmp/asylum-checksums.txt
fi

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
