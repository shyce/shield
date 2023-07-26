package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
)

const (
	Black   = "\u001b[30m"
	Red     = "\u001b[31m"
	Green   = "\u001b[32m"
	Yellow  = "\u001b[33m"
	Blue    = "\u001b[34m"
	Magenta = "\u001b[35m"
	Cyan    = "\u001b[36m"
	White   = "\u001b[37m"
	Reset   = "\u001b[0m"
)

var (
	Author             string
	Encryption         string
	EncryptionTag      string
	EncryptionTagBytes int
	Name               string
	Version            string
	VaultPasswordFile  string
)

const (
	ShieldLinuxPath   = "/usr/local/bin/shield"
	ShieldWindowsPath = `C:\Windows\System32\shield.exe`
	ShieldMacPath     = "/usr/local/bin/shield"
)

var (
	directory, passwordFile                          string
	encrypt, decrypt, generateHook, version, install bool
)

const ShieldNotFound = "Shield is not globally callable for pre-commit hooks. Please ensure Shield is properly installed and added to your system's PATH, then try again. Refer to the Shield README, Downloading and Installing Shield."

func checkShieldInstallation() bool {
	_, err := exec.LookPath("shield")
	return err == nil
}

func installShield() error {
	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error finding current executable: %v", err)
	}

	var targetPath string
	switch runtime.GOOS {
	case "windows":
		targetPath = ShieldWindowsPath
	case "darwin":
		targetPath = ShieldMacPath
	default: // Linux
		targetPath = ShieldLinuxPath
	}

	// Ensure the directory exists
	dir := filepath.Dir(targetPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory: %v", err)
		}
	}

	// Attempt to copy the current executable to the target path
	input, err := os.ReadFile(currentPath)
	if err != nil {
		return fmt.Errorf("error reading current executable: %v", err)
	}

	err = os.WriteFile(targetPath, input, 0755)
	if err != nil {
		return fmt.Errorf("error writing to target path: %v", err)
	}

	return nil
}

func colorPrint(color string, text string) {
	fmt.Println(string(color), text, string(Reset))
}

func init() {
	flag.StringVar(&directory, "v", ".", "directory to operate on")
	flag.BoolVar(&encrypt, "e", false, "Encrypt files")
	flag.BoolVar(&decrypt, "d", false, "Decrypt files")
	flag.BoolVar(&generateHook, "g", false, "Generate Git pre-commit hook")
	flag.BoolVar(&version, "version", false, "Print version information")
	flag.BoolVar(&install, "install", false, "Install Shield. Copies current binary to local user PATH")
	flag.StringVar(&passwordFile, "passwordFile", "", "Specify the password location (default: ~/.ssh/vault)")
	flag.Usage = func() {
		fmt.Println("Usage: shield [OPTION]...")
		fmt.Println("Available options:")
		flag.PrintDefaults()
	}
}

func handleInstall() {
	err := installShield()
	if err != nil {
		log.Fatalf("Installation failed: %v\n", err)
	}
	fmt.Println("Installation successful!")
	os.Exit(0)
}

func handleEncryption() {
	colorPrint(Green, "Encrypting files...")
	encryptFiles()
}

func handleDecryption() {
	colorPrint(Yellow, "Decrypting files...")
	decryptFiles()
}

func handleGenerateHook() {
	colorPrint(Magenta, "Generating Git pre-commit hook...")
	generatePreCommitHook()
}

func handleVersion() {
	colorPrint(Cyan, "Shield Encryption:")
	colorPrint(Blue, fmt.Sprintf("Version: %s", Version))
	colorPrint(Magenta, fmt.Sprintf("Encryption Version: %s", Encryption))
	colorPrint(Yellow, fmt.Sprintf("Author: %s", Author))
	os.Exit(0)
}

func handleDefault() {
	flag.PrintDefaults()
	os.Exit(1)
}

func getVaultPasswordFile() string {
	home := getHomeDirectory()

	if passwordFile == "" {
		return filepath.Join(home, ".ssh", "vault")
	}
	return passwordFile
}

func getHomeDirectory() string {
	home := os.Getenv("HOME")
	if home == "" {
		usr, err := user.Current()
		if err != nil {
			colorPrint(Red, fmt.Sprintf("Cannot get current user: %v", err))
		}
		home = usr.HomeDir
	}
	return home
}

func main() {
	flag.Parse()

	if install {
		handleInstall()
	}

	if !checkShieldInstallation() {
		colorPrint(Red, ShieldNotFound)
		os.Exit(1)
	}

	absDirectory, err := filepath.Abs(directory)
	if err != nil {
		log.Fatalf("Failed to resolve directory to an absolute path: %v\n", err)
	}

	if directory != "." {
		colorPrint(Green, fmt.Sprintf("Operating on directory: %s", absDirectory))
	}

	directory = absDirectory

	EncryptionTag = "SHIELD[" + Encryption + "]:"
	EncryptionTagBytes = len(EncryptionTag)

	VaultPasswordFile = getVaultPasswordFile()

	if encrypt {
		handleEncryption()
	}

	if decrypt {
		handleDecryption()
	}

	if generateHook {
		handleGenerateHook()
	}

	if version {
		handleVersion()
	}

	if !encrypt && !decrypt && !generateHook && !version && !install {
		handleDefault()
	}
}

func getPreCommitScript() string {
	header := regexp.QuoteMeta("SHIELD[" + Encryption + "]:")
	switch runtime.GOOS {
	case "windows":
		return `#!/usr/bin/env powershell

#Requires -Version 5.0
$ErrorActionPreference = "Stop"

function Convert-GlobToRegex {
    param(
        [Parameter(Mandatory=$true)]
        [string]$Glob
    )

    $Glob = [Regex]::Escape($Glob)  # Escape any regex special characters
    $Glob = $Glob -replace '\\\*\\\*', '.*'  # Replace ** with .*
    $Glob = $Glob -replace '\\\*', '[^/]*'  # Replace * with [^/]*
    $Glob = "^$Glob$"  # Match start and end of string

    return $Glob
}

if (!(Get-Command shield -ErrorAction SilentlyContinue)) {
		Write-Host "Shield is not globally callable for pre-commit hooks. Please ensure Shield is properly installed and added to your system's PATH, then try again. Refer to the Shield README, Downloading and Installing Shield."
		exit 1
}

$GLOB_PATTERNS = Get-Content .shield | ForEach-Object { Convert-GlobToRegex $_ }
$OMIT_PATTERNS = Get-Content .shieldignore | ForEach-Object { Convert-GlobToRegex $_ }

$files_to_encrypt = @()
$git_files = git diff --cached --name-only | ForEach-Object { $_.ToString().Trim() }

foreach ($FILE_PATH in $git_files) {
		if (Test-Path $FILE_PATH) {
				foreach ($glob in $GLOB_PATTERNS) {
						if ($FILE_PATH -match $glob) {
								foreach ($omit in $OMIT_PATTERNS) {
										if ($FILE_PATH -match $omit) {
												continue
										}
								}

								if (!(Select-String -Path $FILE_PATH -Pattern "` + header + `" -Quiet)) {
										Write-Host "ERROR: The file $FILE_PATH is not encrypted."
										$files_to_encrypt += $FILE_PATH
								}
						}
				}
		}
}

if ($files_to_encrypt.Count -ne 0) {
		Write-Host "Some files were not encrypted. Running encryption now..."
		shield.exe -e
		foreach ($file in $files_to_encrypt) {
				git add $file
		}
		Write-Host "Files have been encrypted and added to the commit. Please re-run the commit command."
		exit 1
}

exit 0
`
	default:
		return `#!/bin/bash
set -e
shopt -s globstar

if ! command -v shield &> /dev/null; then
		echo "` + ShieldNotFound + `"
		exit 1
fi

GLOB_PATTERNS=()
while IFS= read -r line; do
		GLOB_PATTERNS+=("$line")
done < .shield

OMIT_PATTERNS=()
while IFS= read -r line; do
		OMIT_PATTERNS+=("$line")
done < .shieldignore

files_to_encrypt=()
for FILE_PATH in $(git diff --cached --name-only); do
		if [[ -e $FILE_PATH ]]; then
				for glob in "${GLOB_PATTERNS[@]}"; do
						if [[ $FILE_PATH == $glob ]]; then
								for omit in "${OMIT_PATTERNS[@]}"; do
										if [[ $FILE_PATH == $omit ]]; then
												continue 2
										fi
								done

								if ! head -n 1 "$FILE_PATH" | grep -q "` + header + `"; then
									echo "ERROR: The file $FILE_PATH is not encrypted."
									files_to_encrypt+=("$FILE_PATH")
								fi
						fi
				done
		fi
done

if [ ${#files_to_encrypt[@]} -ne 0 ]; then
		echo "Some files were not encrypted. Running encryption now..."
		shield -e
		for file in "${files_to_encrypt[@]}"; do
				git add "$file"
		done
		echo "Files have been encrypted and added to the commit. Please re-run the commit command."
		exit 1
fi

exit 0
`
	}
}

func generatePreCommitHook() {
	preCommitHookScript := getPreCommitScript()

	if runtime.GOOS == "windows" {
		// Windows needs both pre-commit and pre-commit.ps1

		preCommitHookPath := filepath.Join(directory, ".git/hooks/pre-commit")
		preCommitHookPSPath := filepath.Join(directory, ".git/hooks/pre-commit.ps1")

		// Remove the existing pre-commit hook file if it exists
		if err := os.Remove(preCommitHookPath); err != nil && !os.IsNotExist(err) {
			colorPrint(Red, fmt.Sprintf("Error removing existing pre-commit hook: %s", err))
			os.Exit(1)
		}

		// Remove the existing pre-commit.ps1 hook file if it exists
		if err := os.Remove(preCommitHookPSPath); err != nil && !os.IsNotExist(err) {
			colorPrint(Red, fmt.Sprintf("Error removing existing pre-commit.ps1 hook: %s", err))
			os.Exit(1)
		}

		// Write pre-commit.ps1 hook
		err := os.WriteFile(preCommitHookPSPath, []byte(preCommitHookScript), 0755)
		if err != nil {
			colorPrint(Red, fmt.Sprintf("Error writing pre-commit.ps1 hook: %s", err))
			os.Exit(1)
		}
		colorPrint(Green, "Git pre-commit.ps1 hook successfully generated!")

		// Write pre-commit hook that calls pre-commit.ps1
		preCommitHook := `#!/bin/sh
powershell.exe -ExecutionPolicy Bypass -File .git/hooks/pre-commit.ps1`

		err = os.WriteFile(preCommitHookPath, []byte(preCommitHook), 0755)
		if err != nil {
			colorPrint(Red, fmt.Sprintf("Error writing pre-commit hook: %s", err))
			os.Exit(1)
		}
		colorPrint(Green, "Git pre-commit hook successfully generated!")

	} else {
		preCommitHookPath := filepath.Join(directory, ".git/hooks/pre-commit")

		// Remove the existing pre-commit hook file if it exists
		if err := os.Remove(preCommitHookPath); err != nil && !os.IsNotExist(err) {
			colorPrint(Red, fmt.Sprintf("Error removing existing pre-commit hook: %s", err))
			os.Exit(1)
		}

		err := os.WriteFile(preCommitHookPath, []byte(preCommitHookScript), 0755)
		if err != nil {
			colorPrint(Red, fmt.Sprintf("Error writing pre-commit hook: %s", err))
			os.Exit(1)
		}
		colorPrint(Green, "Git pre-commit hook successfully generated!")
	}
}

func readPatternsFromFile(file string) ([]string, error) {
	file = filepath.Join(directory, file)
	f, err := os.Open(file)
	if err != nil {
		colorPrint(Red, fmt.Sprintf("Error opening file: %s", err))
		return nil, err
	}
	defer f.Close()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		patterns = append(patterns, scanner.Text())
	}

	if scanner.Err() != nil {
		colorPrint(Red, fmt.Sprintf("Error reading file: %s", scanner.Err()))
		return nil, scanner.Err()
	}

	return patterns, nil
}

func processFiles(files []string, actionFunc func(string), wg *sync.WaitGroup, semaphore chan struct{}) {
	for _, path := range files {
		semaphore <- struct{}{}
		wg.Add(1)

		go func(path string) {
			defer func() {
				<-semaphore
				wg.Done()
			}()

			actionFunc(path)
		}(path)
	}
}

func encryptFiles() {
	shieldPatterns, err := readPatternsFromFile(".shield")
	if err != nil {
		colorPrint(Red, "Error reading .shield file, please ensure it exists and is correctly formatted.")
		os.Exit(1)
	}

	shieldIgnorePatterns, err := readPatternsFromFile(".shieldignore")
	if err != nil {
		colorPrint(Red, "Error reading .shieldignore file, please ensure it exists and is correctly formatted.")
		os.Exit(1)
	}

	fsys := os.DirFS(directory)
	var filesToEncrypt []string
	for _, pattern := range shieldPatterns {
		colorPrint(Green, fmt.Sprintf("Looking for files matching pattern: %s", pattern))
		matchingFiles, err := doublestar.Glob(fsys, pattern)
		if err != nil {
			colorPrint(Red, fmt.Sprintf("Error while matching glob pattern: %s", err))
			os.Exit(1)
		}

		for _, filePath := range matchingFiles {
			isOmitted := false
			for _, omitPattern := range shieldIgnorePatterns {
				matches, _ := doublestar.Match(omitPattern, filePath)
				if matches {
					isOmitted = true
					break
				}
			}
			if isOmitted {
				continue
			}

			encrypted, _ := isFileEncrypted(filePath)
			if !encrypted {
				filesToEncrypt = append(filesToEncrypt, filePath)
			}
		}
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, runtime.NumCPU())
	processFiles(filesToEncrypt, encryptFile, &wg, semaphore)
	wg.Wait()
}

func decryptFiles() {
	shieldPatterns, err := readPatternsFromFile(".shield")
	if err != nil {
		colorPrint(Red, "Error reading .shield file, please ensure it exists and is correctly formatted.")
		os.Exit(1)
	}

	shieldIgnorePatterns, err := readPatternsFromFile(".shieldignore")
	if err != nil {
		colorPrint(Red, "Error reading .shieldignore file, please ensure it exists and is correctly formatted.")
		os.Exit(1)
	}

	fsys := os.DirFS(directory)
	var filesToDecrypt []string
	for _, pattern := range shieldPatterns {
		colorPrint(Green, fmt.Sprintf("Looking for files matching pattern: %s", pattern))
		matchingFiles, err := doublestar.Glob(fsys, pattern)
		if err != nil {
			colorPrint(Red, fmt.Sprintf("Error while matching glob pattern: %s", err))
			os.Exit(1)
		}

		for _, filePath := range matchingFiles {
			isOmitted := false
			for _, omitPattern := range shieldIgnorePatterns {
				matches, _ := doublestar.Match(omitPattern, filePath)
				if matches {
					isOmitted = true
					break
				}
			}
			if isOmitted {
				continue
			}

			encrypted, _ := isFileEncrypted(filePath)
			if encrypted {
				filesToDecrypt = append(filesToDecrypt, filePath)
			}
		}
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, runtime.NumCPU())
	processFiles(filesToDecrypt, decryptFile, &wg, semaphore)
	wg.Wait()
}

func encryptFile(path string) {
	path = filepath.Join(directory, path)
	colorPrint(Yellow, fmt.Sprintf("Attempting to encrypt file: %s", path))

	cmd := exec.Command("openssl", "enc", "-aes-256-cbc", "-nosalt", "-pass", fmt.Sprintf("file:%s", VaultPasswordFile), "-in", path, "-out", path+".enc")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		colorPrint(Red, fmt.Sprintf("Failed to encrypt file: %s", err))
	} else {
		err = addEncryptionTag(path + ".enc") // add tag to encrypted file
		if err != nil {
			colorPrint(Red, fmt.Sprintf("Failed to add encryption tag: %s", err))
			if err := os.Remove(path + ".enc"); err != nil { // remove failed encrypted file
				colorPrint(Red, fmt.Sprintf("Failed to remove file: %s", err))
			}
		} else {
			colorPrint(Green, fmt.Sprintf("Encrypted file: %s", path))
			if err := os.Remove(path); err != nil {
				colorPrint(Red, fmt.Sprintf("Failed to remove original file: %s", err))
			}
			if err := os.Rename(path+".enc", path); err != nil {
				colorPrint(Red, fmt.Sprintf("Failed to rename encrypted file: %s", err))
			}
		}
	}
}

func decryptFile(path string) {
	path = filepath.Join(directory, path)
	colorPrint(Yellow, fmt.Sprintf("Attempting to decrypt file: %s", path))

	err := removeEncryptionTag(path) // remove tag before decryption
	if err != nil {
		colorPrint(Red, fmt.Sprintf("Failed to remove encryption tag: %s", err))
	} else {
		cmd := exec.Command("openssl", "enc", "-d", "-aes-256-cbc", "-nosalt", "-pass", fmt.Sprintf("file:%s", VaultPasswordFile), "-in", path, "-out", path+".dec")
		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			colorPrint(Red, fmt.Sprintf("Failed to decrypt file: %s", err))
			if err := addEncryptionTag(path); err != nil { // add tag back if decryption failed
				colorPrint(Red, fmt.Sprintf("Failed to add encryption tag: %s", err))
			}
		} else {
			colorPrint(Green, fmt.Sprintf("Decrypted file: %s", path))
			if err := os.Remove(path); err != nil {
				colorPrint(Red, fmt.Sprintf("Failed to remove encrypted file: %s", err))
			}
			if err := os.Rename(path+".dec", path); err != nil {
				colorPrint(Red, fmt.Sprintf("Failed to rename decrypted file: %s", err))
			}
		}
	}
}

func addEncryptionTag(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content = append([]byte(EncryptionTag), content...)

	return os.WriteFile(path, content, 0666)
}

func removeEncryptionTag(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content = content[EncryptionTagBytes:]

	return os.WriteFile(path, content, 0666)
}

func isFileEncrypted(path string) (bool, error) {
	path = filepath.Join(directory, path)
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	if len(content) >= EncryptionTagBytes && string(content[:EncryptionTagBytes]) == EncryptionTag {
		return true, nil
	}

	return false, nil
}
