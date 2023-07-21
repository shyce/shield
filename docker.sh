#!/bin/bash

echo "Enter the directory to mount (leave blank for current directory):"
read directory

# If no directory was provided, default to "."
directory=${directory:-"."}

absDirectory=$(realpath "$directory")

docker build . -t "shield" && docker run --rm -v "$absDirectory:/app" -v "$HOME/.ssh/vault:/home/user/.ssh/vault" -e "HOME=/home/user" --user "$(id -u):$(id -g)" shield $@