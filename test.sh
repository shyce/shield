#!/bin/bash

# Extract variables from the version.json file
export VERSION=$(jq -r '.version' version.json)
export NAME=$(jq -r '.name' version.json)
export AUTHOR=$(jq -r '.author' version.json)
export ENCRYPTION=$(jq -r '.encryption' version.json)

# Run the tests
go test -v
