#!/bin/bash

################################################################################
# Title: Ansible Vault Shield
# Version: v0.1.0
# Description: This script provides functionality for encrypting and decrypting
#              files using AES encryption and Vault key management system.
################################################################################
#
# Copyright (c) 2023 Brian James Royer (brianroyer@gmail.com)
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.
#
################################################################################

set -e

FILE_PATHS=(
  # Add more file paths as needed
  ".env.example"
)

VAULT_PASSWORD_FILE="~/.ssh/vault"

function generate_pre_commit_hook {
  PRE_COMMIT_HOOK="./.git/hooks/pre-commit"

  if [ -f "$PRE_COMMIT_HOOK" ]; then
    echo "The pre-commit hook already exists. Skipping generation."
  else
    echo "Generating pre-commit hook..."
    echo '#!/bin/bash' > "$PRE_COMMIT_HOOK"
    echo 'set -e' >> "$PRE_COMMIT_HOOK"
    echo >> "$PRE_COMMIT_HOOK"
    echo 'FILE_PATHS=(' >> "$PRE_COMMIT_HOOK"
    for FILE_PATH in "${FILE_PATHS[@]}"; do
      echo "  \"$FILE_PATH\"" >> "$PRE_COMMIT_HOOK"
    done
    echo ')' >> "$PRE_COMMIT_HOOK"
    echo >> "$PRE_COMMIT_HOOK"
    echo 'files_encrypted=0' >> "$PRE_COMMIT_HOOK"
    echo 'for FILE_PATH in "${FILE_PATHS[@]}"; do' >> "$PRE_COMMIT_HOOK"
    echo '  if ! ansible-vault view --vault-password-file ~/.ssh/vault "$FILE_PATH" > /dev/null 2>&1; then' >> "$PRE_COMMIT_HOOK"
    echo '    echo "ERROR: The file $FILE_PATH is not encrypted. Encrypting it now..."' >> "$PRE_COMMIT_HOOK"
    echo '    ansible-vault encrypt --vault-password-file ~/.ssh/vault "$FILE_PATH"' >> "$PRE_COMMIT_HOOK"
    echo '    echo "The file $FILE_PATH has been encrypted. Adding it to the commit..."' >> "$PRE_COMMIT_HOOK"
    echo '    git add "$FILE_PATH"' >> "$PRE_COMMIT_HOOK"
    echo '    files_encrypted=1' >> "$PRE_COMMIT_HOOK"
    echo '  fi' >> "$PRE_COMMIT_HOOK"
    echo 'done' >> "$PRE_COMMIT_HOOK"
    echo 'if [ "$files_encrypted" -eq 1 ]; then' >> "$PRE_COMMIT_HOOK"
    echo '  echo "Some files were not encrypted. They have been encrypted and added to the commit. Please re-run the commit command."' >> "$PRE_COMMIT_HOOK"
    echo '  exit 1' >> "$PRE_COMMIT_HOOK"
    echo 'fi' >> "$PRE_COMMIT_HOOK"
    echo 'exit 0' >> "$PRE_COMMIT_HOOK"

    chmod +x "$PRE_COMMIT_HOOK"
    echo "Pre-commit hook generated successfully."
  fi
}

function encrypt_file {
  for FILE_PATH in "${FILE_PATHS[@]}"; do
    # Check if the file is already encrypted
    if ansible-vault view --vault-password-file $VAULT_PASSWORD_FILE "$FILE_PATH" > /dev/null 2>&1; then
      echo "${FILE_PATH} seems to be already encrypted. Skipping..."
      continue
    fi

    echo "Encrypting ${FILE_PATH}..."
    ansible-vault encrypt --vault-password-file $VAULT_PASSWORD_FILE "$FILE_PATH"
    echo "${FILE_PATH} has been encrypted."
  done
}

function decrypt_file {
  for FILE_PATH in "${FILE_PATHS[@]}"; do
    # Check if the file is already decrypted
    if ! ansible-vault view --vault-password-file $VAULT_PASSWORD_FILE "$FILE_PATH" > /dev/null 2>&1; then
      echo "${FILE_PATH} seems to be already decrypted. Skipping..."
      continue
    fi

    echo "Decrypting ${FILE_PATH}..."
    ansible-vault decrypt --vault-password-file $VAULT_PASSWORD_FILE "$FILE_PATH"
    echo "${FILE_PATH} has been decrypted."
  done
}

generate_pre_commit_hook

# Parsing command line argument
if [ "$1" == "--encrypt" ] || [ "$1" == "-e" ]; then
  encrypt_file
elif [ "$1" == "--decrypt" ] || [ "$1" == "-d" ]; then
  decrypt_file
else
  echo "Invalid argument. Please use either '--encrypt', '-e', '--decrypt', or '-d'."
  exit 1
fi
