#!/bin/bash

# Repository details
GITHUB_USER="BohdanTkachenko"
GITHUB_REPO="virtiofsd-manager"

for cmd in curl jq sha256sum; do
  if ! command -v $cmd &> /dev/null; then
    echo "Error: $cmd is required but not installed. Aborting."
    exit 1
  fi
done

ARCH=$(uname -m)
case "$ARCH" in
  x86_64) PKG_ARCH="amd64" ;;
  i386)   PKG_ARCH="386" ;;
  i686)   PKG_ARCH="386" ;;
  arm64)  PKG_ARCH="arm64" ;;
  aarch64)PKG_ARCH="arm64" ;;
  *)      echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Determine the package manager and system architecture
if command -v apt >/dev/null 2>&1; then
  PKG_FORMAT="deb"
  INSTALL_CMD="sudo apt install ./"
elif command -v dnf >/dev/null 2>&1; then
  PKG_FORMAT="rpm"
  INSTALL_CMD="sudo dnf install ./"
elif command -v yum >/dev/null 2>&1; then
  PKG_FORMAT="rpm"
  INSTALL_CMD="sudo yum install ./"
elif command -v pacman >/dev/null 2>&1; then
  PKG_FORMAT="pkg.tar.zst"
  INSTALL_CMD="sudo pacman -U"
else
  echo "Unsupported package manager. This script supports APT, DNF/YUM, and Pacman."
  exit 1
fi

# Fetch the latest release data from GitHub API
RELEASE_DATA=$(curl -s \
  "https://api.github.com/repos/$GITHUB_USER/$GITHUB_REPO/releases/latest")

# Extract and download the appropriate package URL
PACKAGE_URL=$(echo "$RELEASE_DATA" | \
  jq -r --arg FILE_PATTERN "_linux_${PKG_ARCH}.${PKG_FORMAT}" '.assets[] | select(.name | endswith($FILE_PATTERN)) | .browser_download_url')
PACKAGE_NAME=$(basename "$PACKAGE_URL")
TEMP_PACKAGE_PATH="/tmp/$PACKAGE_NAME"
if [ -z "$PACKAGE_URL" ]; then
  echo "Failed to find a package for the current architecture ($ARCH) and format ($PKG_FORMAT)."
  exit 1
fi
curl -Lo "$TEMP_PACKAGE_PATH" "$PACKAGE_URL"

# Verify checksum
echo "Verifying checksum..."
CHECKSUM_FILE_URL=$(echo "$RELEASE_DATA" \
  | jq -r '.assets[] | select(.name | endswith("_checksums.txt")) | .browser_download_url')
curl -L "$CHECKSUM_FILE_URL" | grep "$PACKAGE_NAME"
CHECKSUM_OK=$(curl -L "$CHECKSUM_FILE_URL" \
  | grep "$PACKAGE_NAME" \
  | sed "s|$PACKAGE_NAME|$TEMP_PACKAGE_PATH|" \
  | sha256sum -c --ignore-missing)

if [[ $CHECKSUM_OK == *": OK"* ]]; then
  echo "Checksum verification passed."
  echo "Installing $PACKAGE_NAME..."
  $INSTALL_CMD "$PACKAGE_NAME" && \
    echo "Cleaning up..." && \
    rm "$TEMP_PACKAGE_PATH"
else
  echo "Checksum verification failed."
  echo "Cleaning up..." && \
  rm "$TEMP_PACKAGE_PATH"
  exit 1
fi
