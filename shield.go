package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/bmatcuk/doublestar"
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

func colorPrint(color string, text string) {
	fmt.Println(string(color), text, string(Reset))
}

func main() {
	if len(os.Args) < 2 {
		colorPrint(Red, "Insufficient arguments. Please use either '--encrypt', '-e', '--decrypt', '-d', '--generate-hooks', '-g', '--version', '-v', or '--help', '-h'.")
		os.Exit(1)
	}

	EncryptionTag = "SHIELD[" + Encryption + "]:"
	EncryptionTagBytes = len(EncryptionTag)
	usr, err := user.Current()

	if err != nil {
		colorPrint(Red, fmt.Sprintf("Cannot get current user: %v", err))
	}

	VaultPasswordFile = filepath.Join(usr.HomeDir, ".ssh", "vault")

	switch os.Args[1] {
	case "--encrypt", "-e":
		colorPrint(Green, "Encrypting files...")
		encryptFiles()
	case "--decrypt", "-d":
		colorPrint(Yellow, "Decrypting files...")
		decryptFiles()
	case "--generate-hooks", "-g":
		colorPrint(Magenta, "Generating Git pre-commit hook...")
		generatePreCommitHook()
		generatePreCheckoutHook()
	case "--version", "-v", "version":
		colorPrint(Cyan, "Shield Encryption:")
		colorPrint(Blue, fmt.Sprintf("Version: %s", Version))
		colorPrint(Magenta, fmt.Sprintf("Encryption Version: %s", Encryption))
		colorPrint(Yellow, fmt.Sprintf("Author: %s", Author))
	case "--help", "-h":
		colorPrint(Cyan, "Shield Help:")
		colorPrint(Reset, "Usage: shield [OPTION]...")
		colorPrint(Reset, "Available options:")
		colorPrint(Green, "--encrypt, -e			Encrypt files")
		colorPrint(Yellow, "--decrypt, -d			Decrypt files")
		colorPrint(Magenta, "--generate-hooks, -g		Generate Git pre-commit hook")
		colorPrint(Blue, "--version, -v			Print version information")
		colorPrint(Cyan, "--help, -h			Show this help message and exit")
	default:
		colorPrint(Red, "Invalid argument. Please use either '--encrypt', '-e', '--decrypt', '-d', '--generate-hooks', '-g', '--version', '-v', or '--help', '-h'.")
		os.Exit(1)
	}
}

func generatePreCommitHook() {
	preCommitHookPath := ".git/hooks/pre-commit"

	// Remove the existing pre-commit hook file if it exists
    if err := os.Remove(preCommitHookPath); err != nil && !os.IsNotExist(err) {
			colorPrint(Red, fmt.Sprintf("Error removing existing pre-commit hook: %s", err))
			os.Exit(1)
	}

	preCommitHookScript := `#!/bin/bash
set -e
shopt -s globstar

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

        if ! file "$FILE_PATH" | grep -q 'data'; then
          echo "ERROR: The file $FILE_PATH is not encrypted."
          files_to_encrypt+=("$FILE_PATH")
        fi
      fi
    done
  fi
done

if [ ${#files_to_encrypt[@]} -ne 0 ]; then
  echo "Some files were not encrypted. Running encryption now..."
  ./shield -e
  for file in "${files_to_encrypt[@]}"; do
    git add "$file"
  done
  echo "Files have been encrypted and added to the commit. Please re-run the commit command."
  exit 1
fi

exit 0
`
	err := os.WriteFile(preCommitHookPath, []byte(preCommitHookScript), 0755)
	if err != nil {
			colorPrint(Red, fmt.Sprintf("Error writing pre-commit hook: %s", err))
			os.Exit(1)
	}
	colorPrint(Green, "Git pre-commit hook successfully generated!")
}

func generatePreCheckoutHook() {
	preCheckoutHookPath := ".git/hooks/pre-checkout"

	// Remove the existing pre-checkout hook file if it exists
    if err := os.Remove(preCheckoutHookPath); err != nil && !os.IsNotExist(err) {
			colorPrint(Red, fmt.Sprintf("Error removing existing pre-checkout hook: %s", err))
			os.Exit(1)
	}

	preCheckoutHookScript := `#!/bin/bash
set -e
shopt -s globstar

GLOB_PATTERNS=()
while IFS= read -r line; do
  GLOB_PATTERNS+=("$line")
done < .shield

OMIT_PATTERNS=()
while IFS= read -r line; do
  OMIT_PATTERNS+=("$line")
done < .shieldignore

files_to_encrypt=()
for FILE_PATH in **; do
  if [[ -e $FILE_PATH ]]; then
    for glob in "${GLOB_PATTERNS[@]}"; do
      if [[ $FILE_PATH == $glob ]]; then
        for omit in "${OMIT_PATTERNS[@]}"; do
          if [[ $FILE_PATH == $omit ]]; then
            continue 2
          fi
        done

        if ! file "$FILE_PATH" | grep -q 'data'; then
          files_to_encrypt+=("$FILE_PATH")
        fi
      fi
    done
  fi
done

if [ ${#files_to_encrypt[@]} -ne 0 ]; then
  echo "Some files were not encrypted. Running encryption now..."
  ./shield -e
  echo "Files have been encrypted."
  exit 1
fi

exit 0
`
	err := os.WriteFile(preCheckoutHookPath, []byte(preCheckoutHookScript), 0755)
	if err != nil {
			colorPrint(Red, fmt.Sprintf("Error writing pre-checkout hook: %s", err))
			os.Exit(1)
	}
	colorPrint(Green, "Git pre-checkout hook successfully generated!")
}

func readPatternsFromFile(file string) ([]string, error) {
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

	var filesToEncrypt []string
	for _, pattern := range shieldPatterns {
			colorPrint(Green, fmt.Sprintf("Looking for files matching pattern: %s", pattern))
			matchingFiles, err := doublestar.Glob(pattern)
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

	var filesToDecrypt []string
	for _, pattern := range shieldPatterns {
			colorPrint(Green, fmt.Sprintf("Looking for files matching pattern: %s", pattern))
			matchingFiles, err := doublestar.Glob(pattern)
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
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	if len(content) >= EncryptionTagBytes && string(content[:EncryptionTagBytes]) == EncryptionTag {
		return true, nil
	}

	return false, nil
}
