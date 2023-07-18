#!/bin/bash

################################################################################
# Title: Vault Shield
# Version: v0.3.0
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


if [ -z "$KEY_NAME" ]; then
  read -p "Enter the Vault key name: " KEY_NAME_INPUT
  export KEY_NAME="$KEY_NAME_INPUT"

  if [[ $SHELL == *"/bash"* ]]; then
    if ! grep -q "export KEY_NAME=" "$HOME/.bashrc"; then
      echo "export KEY_NAME=$KEY_NAME" >> "$HOME/.bashrc"
      source $HOME/.bashrc
    fi
  elif [[ $SHELL == *"/zsh"* ]]; then
    if ! grep -q "export KEY_NAME=" "$HOME/.zshrc"; then
      echo "export KEY_NAME=$KEY_NAME" >> "$HOME/.zshrc"
      source $HOME/.zshrc
    fi
  elif [[ $SHELL == *"/fish"* ]]; then
    if ! grep -q "set -x KEY_NAME" "$HOME/.config/fish/config.fish"; then
      echo "set -x KEY_NAME $KEY_NAME" >> "$HOME/.config/fish/config.fish"
      source $HOME/.config/fish/config.fish
    fi
  fi
fi

if [ -z "$VAULT_ADDR" ]; then
  read -p "Enter the Vault address [http://127.0.0.1:8200]: " VAULT_ADDR_INPUT
  export VAULT_ADDR="$VAULT_ADDR_INPUT"

  # Verify the Vault address
  if ! response_code=$(curl -s -o /dev/null -I -w "%{http_code}" "$VAULT_ADDR"); then
    echo "Failed to connect to Vault. Exiting..."
    exit 1
  fi

  if [ "$response_code" -lt 200 ] || [ "$response_code" -ge 400 ]; then
    echo "Invalid Vault address. Exiting..."
    exit 1
  fi

  if [[ $SHELL == *"/bash"* ]]; then
    if ! grep -q "export VAULT_ADDR=" "$HOME/.bashrc"; then
      echo "export VAULT_ADDR=$VAULT_ADDR" >> "$HOME/.bashrc"
      source $HOME/.bashrc
    fi
  elif [[ $SHELL == *"/zsh"* ]]; then
    if ! grep -q "export VAULT_ADDR=" "$HOME/.zshrc"; then
      echo "export VAULT_ADDR=$VAULT_ADDR" >> "$HOME/.zshrc"
      source $HOME/.zshrc
    fi
  elif [[ $SHELL == *"/fish"* ]]; then
    if ! grep -q "set -x VAULT_ADDR" "$HOME/.config/fish/config.fish"; then
      echo "set -x VAULT_ADDR $VAULT_ADDR" >> "$HOME/.config/fish/config.fish"
      source $HOME/.config/fish/config.fish
    fi
  fi
fi

function check_and_create_key {
  # Check if the key exists
  if ! vault read -field=name transit/keys/"${KEY_NAME}" > /dev/null 2>&1; then
    # If not, create the key
    echo "The key ${KEY_NAME} does not exist. Creating it now..."
    vault write -f transit/keys/"${KEY_NAME}"
  else
    echo "The key ${KEY_NAME} already exists."
  fi
}

function generate_pre_commit_hook {
  PRE_COMMIT_HOOK="./.git/hooks/pre-commit"

  if [ -f "$PRE_COMMIT_HOOK" ]; then
    echo "The pre-commit hook already exists. Skipping generation."
  else
    echo "Generating pre-commit hook..."
    echo '#!/bin/bash' > "$PRE_COMMIT_HOOK"
    echo >> "$PRE_COMMIT_HOOK"
    echo 'set -e' >> "$PRE_COMMIT_HOOK"
    echo >> "$PRE_COMMIT_HOOK"
    echo 'FILE_PATHS=(' >> "$PRE_COMMIT_HOOK"
    for FILE_PATH in "${FILE_PATHS[@]}"; do
      echo "  \"$FILE_PATH\"" >> "$PRE_COMMIT_HOOK"
    done
    echo ')' >> "$PRE_COMMIT_HOOK"
    echo >> "$PRE_COMMIT_HOOK"
    echo "VAULT_ADDR=\"$VAULT_ADDR\"" >> "$PRE_COMMIT_HOOK"
    echo 'KEY_NAME="infrastructure"' >> "$PRE_COMMIT_HOOK"
    echo >> "$PRE_COMMIT_HOOK"
    echo 'for FILE_PATH in "${FILE_PATHS[@]}"; do' >> "$PRE_COMMIT_HOOK"
    echo '  CONTENT=$(cat "$FILE_PATH")' >> "$PRE_COMMIT_HOOK"
    echo '  if [[ "$CONTENT" != "vault:"* ]]; then' >> "$PRE_COMMIT_HOOK"
    echo '    echo "ERROR: The file $FILE_PATH is not encrypted. Encrypting it now..."' >> "$PRE_COMMIT_HOOK"
    echo '    ./shield.sh --encrypt "$FILE_PATH"' >> "$PRE_COMMIT_HOOK"
    echo '    echo "The file $FILE_PATH has been encrypted. Please add it to the commit."' >> "$PRE_COMMIT_HOOK"
    echo '    exit 1' >> "$PRE_COMMIT_HOOK"
    echo '  fi' >> "$PRE_COMMIT_HOOK"
    echo 'done' >> "$PRE_COMMIT_HOOK"
    echo >> "$PRE_COMMIT_HOOK"
    echo 'exit 0' >> "$PRE_COMMIT_HOOK"

    chmod +x "$PRE_COMMIT_HOOK"
    echo "Pre-commit hook generated successfully."
  fi
}

################################################################################
# Function: encrypt_file
# Description: Encrypts the specified files using AES encryption and Vault key
#              management system. It generates a local AES key, encrypts the file
#              locally, encrypts the local AES key using Vault, and saves the
#              encrypted AES key and file content.
# Parameters: None
# Returns: None
################################################################################
function encrypt_file {
  for FILE_PATH in "${FILE_PATHS[@]}"; do
    # Check if the file is already encrypted
    if grep -q "vault:" "${FILE_PATH}"; then
      echo "${FILE_PATH} seems to be already encrypted. Skipping..."
      continue
    fi

    echo "Encrypting ${FILE_PATH}..."

    # Generate a local AES key and encrypt the file locally
    LOCAL_AES_KEY=$(openssl rand -base64 32)
    openssl enc -aes-256-cbc -salt -a -k "$LOCAL_AES_KEY" -pbkdf2 -iter 100000 -in "${FILE_PATH}" -out "${FILE_PATH}.enc"
    
    # Encrypt the local AES key using Vault
    VAULT_ENCRYPTED_AES_KEY=$(echo -n "${LOCAL_AES_KEY}" | base64 | vault write -field=ciphertext transit/encrypt/"${KEY_NAME}" plaintext=-)

    # Save the encrypted AES key and the encrypted file content
    echo "${VAULT_ENCRYPTED_AES_KEY}" > "${FILE_PATH}"
    cat "${FILE_PATH}.enc" >> "${FILE_PATH}"

    # Securely remove the temporary encrypted file
    shred -u "${FILE_PATH}.enc"
    
    echo "${FILE_PATH} has been encrypted."
  done
}

################################################################################
# Function: decrypt_file
# Description: Decrypts the specified files that are encrypted using AES encryption
#              and Vault key management system. It extracts the Vault-encrypted
#              AES key and the encrypted file content, decrypts the AES key using
#              Vault, and decrypts the file using the decrypted AES key.
# Parameters: None
# Returns: None
################################################################################
function decrypt_file {
  for FILE_PATH in "${FILE_PATHS[@]}"; do
    # Check if the file is already decrypted
    if ! grep -q "vault:" "${FILE_PATH}"; then
      echo "${FILE_PATH} seems to be already decrypted. Skipping..."
      continue
    fi
    
    echo "Decrypting ${FILE_PATH}..."

    # Extract the Vault-encrypted AES key and the encrypted file content
    VAULT_ENCRYPTED_AES_KEY=$(head -n 1 "${FILE_PATH}")
    tail -n +2 "${FILE_PATH}" > "${FILE_PATH}.enc"
    
    # Decrypt the AES key using Vault
    LOCAL_AES_KEY=$(vault write -field=plaintext transit/decrypt/"${KEY_NAME}" ciphertext="${VAULT_ENCRYPTED_AES_KEY}" | base64 --decode)
    
    # Decrypt the file using the decrypted AES key
    openssl enc -aes-256-cbc -d -a -k "$LOCAL_AES_KEY" -pbkdf2 -iter 100000 -in "${FILE_PATH}.enc" -out "${FILE_PATH}"
    
    # Securely remove the temporary encrypted file
    shred -u "${FILE_PATH}.enc"
    
    echo "${FILE_PATH} has been decrypted."
  done
}

# Check if the pre-commit hook exists, if not, generate it
generate_pre_commit_hook

# Check if the key exists, if not, create it
check_and_create_key

# Parsing command line argument
if [ "$1" == "--encrypt" ] || [ "$1" == "-e" ]; then
  encrypt_file
elif [ "$1" == "--decrypt" ] || [ "$1" == "-d" ]; then
  decrypt_file
else
  echo "Invalid argument. Please use either '--encrypt', '-e', '--decrypt', or '-d'."
  exit 1
fi
