#!/bin/sh
set -e

# Config
OWNER="DishIs"
REPO="fce-cli"
BINARY_NAME="fce"
GITHUB_API="https://api.github.com/repos/$OWNER/$REPO/releases/latest"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux*)   OS='linux';;
  darwin*)  OS='darwin';;
  msys*|cygwin*|mingw*) OS='windows';;
  *)        echo "Unsupported OS: $OS"; exit 1;;
esac

# Detect Arch
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH='amd64';;
  arm64|aarch64) ARCH='arm64';;
  *)      echo "Unsupported architecture: $ARCH"; exit 1;;
esac

echo "Installing $BINARY_NAME for $OS/$ARCH..."

# Get latest version and download URL
DOWNLOAD_URL=$(curl -s $GITHUB_API | grep "browser_download_url" | grep "${OS}_${ARCH}" | cut -d '"' -f 4 | head -n 1)

if [ -z "$DOWNLOAD_URL" ]; then
  echo "Could not find a release for $OS/$ARCH"
  exit 1
fi

# Create temp directory
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# Download and extract
echo "Downloading $DOWNLOAD_URL..."
curl -sL "$DOWNLOAD_URL" -o "$TMP_DIR/archive"

if [ "$OS" = "windows" ]; then
  unzip -q "$TMP_DIR/archive" -d "$TMP_DIR"
else
  tar -xzf "$TMP_DIR/archive" -C "$TMP_DIR"
fi

# Move to /usr/local/bin
echo "Installing to /usr/local/bin/$BINARY_NAME (requires sudo)..."
sudo mv "$TMP_DIR/$BINARY_NAME" /usr/local/bin/

echo "Successfully installed $BINARY_NAME!"
$BINARY_NAME version || echo "Installation complete. Run '$BINARY_NAME' to get started."
