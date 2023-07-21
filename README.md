# Shield - File Encryption & Decryption in your Project

Under the Apache-2.0 License, `Shield` is a utility for encrypting and decrypting files in your project. It uses AES-256 encryption via the `openssl` command line tool.

## Setup

1. Compile the `shield` utility and place the resulting binary at the root of your project.

2. Create a `.shield` file at the root of your project. This file will list all the glob patterns for files you want to encrypt. Each pattern should be on a new line. For example:
    ```
    *.secret
    **/secrets/*.txt
    secrets/**/*.pem
    ```
    The above patterns will match:
    - Any file with the `.secret` extension in the root directory.
    - Any `.txt` file in any `secrets` directory at any level.
    - Any `.pem` file in any directory or sub-directory under a root-level `secrets` directory.

3. Create a `.shieldignore` file at the root of your project. This file will list all the glob patterns for files you want to exclude from encryption. Each pattern should be on a new line. This is useful for ignoring certain files or directories. For example:
    ```
    test/*
    temp.secret
    **/vendors/**
    ```
    The above patterns will ignore:
    - Any files in the `test` directory in the root of your project.
    - The `temp.secret` file in the root directory.
    - Any files in any `vendors` directory at any level.

4. Make sure to have your encryption password stored in a `vault` file located in your home directory under `.ssh`. This is used to encrypt and decrypt files in the repository and should be shared between developers of your project.

## Usage

- To **encrypt files**, run the command: `./shield -e`. This will encrypt all files that match the patterns in your `.shield` file and do not match any patterns in your `.shieldignore` file.
  
- To **decrypt files**, run the command: `./shield -d`. This will decrypt all encrypted files that match the patterns in your `.shield` file and do not match any patterns in your `.shieldignore` file.
  
- To **generate a git pre-commit hook** that checks for unencrypted files (from the `.shield` patterns) and encrypts them, run the command: `./shield -g`. After running this command, every time you try to commit, the hook will check for unencrypted files and encrypt them.

## Developer Guide

### Developer Requirements

- [jq](https://stedolan.github.io/jq/download/)
- [openssl](https://www.openssl.org/source/)
- [git-chglog](https://github.com/git-chglog/git-chglog)
- [Go](https://golang.org/dl/) (version 1.16 or newer)

Before you can build or version `Shield`, please ensure that the above tools are installed.

### Installation

### jq
---
##### macOS
```bash
brew install jq
```

##### Ubuntu/RHEL
Ubuntu:
```bash
sudo apt-get install jq
```
RHEL:
```bash
sudo yum install jq
```

### OpenSSL
---
##### macOS
```bash
brew install openssl
```

##### Ubuntu/RHEL
Ubuntu:
```bash
sudo apt-get install openssl
```
RHEL:
```bash
sudo yum install openssl
```

### git-chglog
---

```bash
go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest
```

### Go
---

Download the latest Go package from [here](https://golang.org/dl/) and follow the installation instructions.

### Building the Project

To build the `shield` utility, run the `build.sh` script at the root of the project:

```bash
./build.sh
```

This script reads version information from `version.json`, builds the Go project with these details, and places the resulting binary in the root of the project.

### Versioning

To bump up the version of the project, run the `version.sh` script with the relevant arguments:

```bash
./version.sh <major|minor|patch>
```

This will increment the specified version part by one. For example, if the current version is `1.2.3`, running `./version.sh patch` will change the version to `1.2.4`.

The version changes are reflected in `version.json` file.

## Note

The encrypted files are prefixed with a specific tag "SHIELD[1.0]:" to help recognize them. The encryption and decryption processes are managed by OpenSSL. If you encounter any issues, make sure OpenSSL is correctly installed and the path is set.

## Warning

- This is a basic implementation and might not cover all use cases or error scenarios.
- Be cautious while using it for production. Test thoroughly to ensure it meets your requirements.
- Be aware that if the encryption process is interrupted or fails, it could leave behind unencrypted data or corrupt your files. Always backup your data.

## Disclaimer

Use this tool at your own risk. The author is not responsible for any data loss or damage.
