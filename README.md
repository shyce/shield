# Shield - File Encryption & Decryption in your Project

`Shield` is a utility for encrypting and decrypting files in your project. It uses AES-256 encryption via the `openssl` command line tool. 

## Setup

1. Compile the `shield` utility and place the resulting binary at the root of your project.

2. Create a `.shield` file at the root of your project. This file will list all the glob patterns for files you want to encrypt. Each pattern should be on a new line. For example:
    ```
    *.secret
    */secrets/*
    ```

3. Create a `.shieldignore` file at the root of your project. This file will list all the glob patterns for files you want to exclude from encryption. Each pattern should be on a new line. This is useful for ignoring certain files or directories. For example:
    ```
    test/*
    temp.secret
    ```

4. Make sure to have your encryption password stored in a `vault` file located in your home directory under `.ssh`. This is referred to in the code as `vaultPasswordFile`.

## Usage

- To **encrypt files**, run the command: `./shield -e`
  
- To **decrypt files**, run the command: `./shield -d`
  
- To **generate a git pre-commit hook** that checks for unencrypted files (from the `.shield` patterns) and encrypts them, run the command: `./shield -g`. After running this command, every time you try to commit, the hook will check for unencrypted files and encrypt them.

## Note

The encrypted files are prefixed with a specific tag "SHIELD[1.0]:" to help recognize them. The encryption and decryption processes are managed by OpenSSL. If you encounter any issues, make sure OpenSSL is correctly installed and the path is set.