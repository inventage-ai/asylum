#!/bin/sh
set -e

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

COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BINARY="asylum-${OS}-${ARCH}"

echo "Building asylum (${COMMIT}) for ${OS}/${ARCH} in Docker..."
docker run --rm \
    -v "$(pwd):/src" \
    -w /src \
    -e "GOOS=${OS}" \
    -e "GOARCH=${ARCH}" \
    golang:1.26 \
    go build -buildvcs=false -ldflags "-X main.version=local -X main.commit=${COMMIT}" -o "/src/build/${BINARY}" ./cmd/asylum

mkdir -p "$INSTALL_DIR"
mv "build/${BINARY}" "${INSTALL_DIR}/asylum"
chmod +x "${INSTALL_DIR}/asylum"

# Create symlink — same logic as install.sh
target="${SYMLINK_DIR}/asylum"
current=$(readlink "$target" 2>/dev/null || true)
if [ "$current" = "${INSTALL_DIR}/asylum" ]; then
    : # symlink already correct
elif [ -w "$SYMLINK_DIR" ]; then
    rm -f "$target"
    ln -s "${INSTALL_DIR}/asylum" "$target"
else
    echo "Updating symlink in ${SYMLINK_DIR} (requires sudo)..."
    sudo rm -f "$target"
    sudo ln -s "${INSTALL_DIR}/asylum" "$target"
fi

echo "asylum local (${COMMIT}) installed to ${INSTALL_DIR}/asylum"
echo "Symlinked from ${target}"
