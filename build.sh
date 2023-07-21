#!/bin/bash

# Extract variables from the version.json file
version=$(jq -r '.version' version.json)
name=$(jq -r '.name' version.json)
author=$(jq -r '.author' version.json)
encryption=$(jq -r '.encryption' version.json)

# Array of GOOS and GOARCH to build for
platforms=("windows/amd64" "darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64")

# Loop through each platform and build
for platform in "${platforms[@]}"
do 
    # Split platform into OS and architecture
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}

    # Name the output file based on the platform
    output_name=$name'-'$version'-'$GOOS'-'$GOARCH

    # Print a message to indicate which platform we're building for
    echo "Building for $output_name"

    # Build for the current platform
    env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-X 'main.Version=$version' -X 'main.Name=$name' -X 'main.Author=$author' -X 'main.Encryption=$encryption'" -o build/$output_name
done
