#!/bin/bash

# Create directories if they do not exist
mkdir -p secrets vendors test

# Generate random content
randomContent=$(cat /dev/urandom | tr -dc 'a-zA-Z' | fold -w 180 | head -n 1)

# Create secret files
echo "This file should be encrypted: $randomContent" > my.secret
echo "This file should be encrypted: $randomContent" > secrets/info.txt
echo "This file should be encrypted: $randomContent" > secrets/key.pem
echo "This file should not be encrypted: $randomContent" > temp.secret

# Create non-secret files
echo "This file should not be encrypted: $randomContent" > vendors/info.txt
echo "This file should not be encrypted: $randomContent" > test/info.txt
