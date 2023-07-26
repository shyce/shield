# Shield - File Encryption & Decryption in your Project

Under the Apache-2.0 License, `Shield` is a utility for encrypting and decrypting files in your project. It uses AES-256 encryption via the `openssl` command line tool.

---

## Downloading and Installing Shield

Shield can be easily installed on macOS, Linux, and Windows. Choose the script appropriate for your operating system and execute it in your terminal to download and install the latest version of Shield:

### macOS/Linux:

```bash
curl -sSL https://raw.githubusercontent.com/shyce/shield/main/install.sh | bash
```

Optionally, you can specify a particular version to install like this:

```bash
curl -sSL https://raw.githubusercontent.com/shyce/shield/main/install.sh | bash -s -- -v <version>
```
Replace `<version>` with the version number that you want to install.

### Windows:

On Windows, open a PowerShell window and run the following command:

```powershell
Set-ExecutionPolicy RemoteSigned
iwr -useb https://raw.githubusercontent.com/shyce/shield/main/install.ps1 | iex
```

Optionally, you can specify a particular version to install like this:

```powershell
Set-ExecutionPolicy RemoteSigned
iwr -useb https://raw.githubusercontent.com/shyce/shield/main/install.ps1 | iex -v <version>
```
Replace `<version>` with the version number that you want to install.

## Self-installation

Shield provides a self-installation feature for users who already have the binary. This functionality can be triggered using the `--install` flag. When the `--install` flag is passed, Shield will attempt to copy its binary to a system-wide location, making it available from any directory in your terminal. This can be especially helpful when working with multiple projects that all utilize Shield. 

To use this feature, navigate to the directory where the Shield binary is located and run the following command:

```bash
./shield --install
```

The binary will then be copied to the appropriate system-wide location, depending on your operating system.

After successful installation, you should be able to run Shield from any directory by just typing `shield` followed by the desired flags. 

Please ensure you have the necessary permissions to write to the target directory. On Unix-based systems, you may need to prepend the command with `sudo` if you encounter permission issues.

## Setup

1. Create a `~/.ssh/vault` file with a secure password and `chmod 600 ~/.ssh/vault`. This is used to encrypt and decrypt files in the repository and should be shared between developers of your project.

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

4. Generate a pre-commit hook in your project with `shield -g`

## Usage

- To **encrypt files**, run the command: `shield -e`. This will encrypt all files that match the patterns in your `.shield` file and do not match any patterns in your `.shieldignore` file.
  
- To **decrypt files**, run the command: `shield -d`. This will decrypt all encrypted files that match the patterns in your `.shield` file and do not match any patterns in your `.shieldignore` file.
  
- To **generate a git pre-commit hook** that checks for unencrypted files (from the `.shield` patterns) and encrypts them, run the command: `shield -g`. After running this command, every time you try to commit, the hook will check for unencrypted files and encrypt them.

## Running Shield with Docker

Shield can be run inside a Docker container without installing anything else on your system. The included `docker.sh` script will handle building and running the Docker container for you.

Before running the `docker.sh` script, ensure you have [Docker](https://www.docker.com/get-started) installed and running on your system.

1. Make sure the `docker.sh` script is executable. If it's not, you can make it executable by running `chmod +x docker.sh`.

2. Run the `docker.sh` script:

    ```bash
    ./docker.sh
    ```

    This script will ask you to enter the directory you want to mount inside the Docker container. This should be the directory containing the files you want to encrypt or decrypt with Shield.

    The script will automatically build the Shield Docker image and then run it, mounting the specified directory and your `vault` file inside the Docker container. This means you can run Shield inside the Docker container just like you would if it was installed directly on your system.

Note: This Docker container runs as the current user on your system and mounts your `vault` file and the specified directory, so it will be able to read and write files in the mounted directory and read your `vault` file.

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
