#!/bin/bash

# Extract variables from the version.json file
version=$(jq -r '.version' version.json)
name=$(jq -r '.name' version.json)
author=$(jq -r '.author' version.json)
encryption=$(jq -r '.encryption' version.json)

# Use these variables in the go build command to set them in the main package
go build -ldflags="-X 'main.Version=$version' -X 'main.Name=$name' -X 'main.Author=$author' -X 'main.Encryption=$encryption'" -o $name
