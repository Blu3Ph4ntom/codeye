#!/usr/bin/env sh
# install.sh — one-line installer for codeye
# Usage: curl -sSfL https://codeye.bluephantom.dev/install.sh | sh

set -e

REPO="blu3ph4ntom/codeye"
BINARY="codeye"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)  OS="linux" ;;
        darwin) OS="darwin" ;;
        freebsd) OS="freebsd" ;;
        mingw*|msys*|cygwin*) OS="windows" ;;
        *) echo "Unsupported OS: $OS" && exit 1 ;;
    esac

    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        armv7l) ARCH="armv7" ;;
        i386|i686) ARCH="386" ;;
        *) echo "Unsupported arch: $ARCH" && exit 1 ;;
    esac

    EXT="tar.gz"
    [ "$OS" = "windows" ] && EXT="zip"
}

fetch_latest_version() {
    VERSION=$(curl -sSfL "https://api.github.com/repos/$REPO/releases/latest" \
        | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    [ -z "$VERSION" ] && echo "Could not determine latest version" && exit 1
}

download_and_install() {
    TARBALL="${BINARY}_${VERSION#v}_${OS}_${ARCH}.${EXT}"
    URL="https://github.com/$REPO/releases/download/$VERSION/$TARBALL"

    TMP=$(mktemp -d)
    trap 'rm -rf "$TMP"' EXIT

    echo "→ Downloading $URL"
    curl -sSfL "$URL" -o "$TMP/$TARBALL"

    echo "→ Extracting..."
    if [ "$EXT" = "tar.gz" ]; then
        tar -xzf "$TMP/$TARBALL" -C "$TMP"
    else
        unzip -q "$TMP/$TARBALL" -d "$TMP"
    fi

    echo "→ Installing to $INSTALL_DIR/$BINARY"
    if [ -w "$INSTALL_DIR" ]; then
        cp "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
        chmod +x "$INSTALL_DIR/$BINARY"
    else
        sudo cp "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
        sudo chmod +x "$INSTALL_DIR/$BINARY"
    fi

    echo "✓ codeye $VERSION installed successfully"
    "$INSTALL_DIR/$BINARY" --version
}

detect_platform
fetch_latest_version
download_and_install
