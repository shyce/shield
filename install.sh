#!/bin/bash

# Set repository details
USERNAME="shyce"
TOOL="shield"

# Set default version to latest
VERSION="latest"

# Read flags
while getopts v: flag
do
    case "${flag}" in
        v) VERSION=${OPTARG};;
    esac
done

# Set default download URL to latest
DOWNLOAD_URL="https://api.github.com/repos/$USERNAME/$TOOL/releases/latest"

# If a version is specified, adjust the download URL
if [ "$VERSION" != "latest" ]; then
    DOWNLOAD_URL="https://api.github.com/repos/$USERNAME/$TOOL/releases/tags/$VERSION"
fi

# Get the version if it's still set to latest
if [ "$VERSION" == "latest" ]; then
    VERSION=$(curl --silent "$DOWNLOAD_URL" | jq -r '.tag_name')
fi

# Identify the OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# Identify the architecture
ARCH=$(uname -m)

# Map certain architectures to values used in the asset names
case "$ARCH" in
    "x86_64") ARCH="amd64";;
    "aarch64") ARCH="arm64";;
esac

# Set the asset name pattern
ASSET="$TOOL-$VERSION-$OS-$ARCH"

# Download the binary
curl -sL $(curl --silent "$DOWNLOAD_URL" | jq -r ".assets[] | select(.name | test(\"$ASSET\")) | .browser_download_url") -o /tmp/$TOOL

# Make it executable
chmod +x /tmp/$TOOL

# Move it to the /usr/local/bin directory
sudo mv /tmp/$TOOL /usr/local/bin/$TOOL

echo "Installed $TOOL version $VERSION"
