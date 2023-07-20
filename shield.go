package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"sync"
	"runtime"
	"path/filepath"
)
var (
	vaultPasswordFile string
)

const (
	EncryptionTag      = "SHIELD[1.0]:"
	EncryptionTagBytes = len(EncryptionTag)
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Insufficient arguments. Exiting.")
		os.Exit(1)
	}

	usr, err := user.Current()
	if err != nil {
		fmt.Printf("cannot get current user: %v", err)
	}
	vaultPasswordFile = filepath.Join(usr.HomeDir, ".ssh", "vault")

	switch os.Args[1] {
	case "--encrypt", "-e":
		fmt.Println("Encrypting files...")
		encryptFiles()
	case "--decrypt", "-d":
		fmt.Println("Decrypting files...")
		decryptFiles()
	case "--generate-precommit", "-g":
		fmt.Println("Generating Git pre-commit hook...")
		generatePreCommitHook()
	default:
		fmt.Println("Invalid argument. Please use either '--encrypt', '-e', '--decrypt', or '-d'.")
		os.Exit(1)
	}
}

func generatePreCommitHook() {
	preCommitHookPath := ".git/hooks/pre-commit"

	// Remove the existing pre-commit hook file if it exists
	if err := os.Remove(preCommitHookPath); err != nil && !os.IsNotExist(err) {
		fmt.Println("Error removing existing pre-commit hook:", err)
		os.Exit(1)
	}

	preCommitHookScript := `#!/bin/bash
set -e

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
	err := ioutil.WriteFile(preCommitHookPath, []byte(preCommitHookScript), 0755)
	if err != nil {
		fmt.Println("Error writing pre-commit hook:", err)
		os.Exit(1)
	}
	fmt.Println("Git pre-commit hook successfully generated!")
}


func readPatternsFromFile(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		patterns = append(patterns, scanner.Text())
	}

	if scanner.Err() != nil {
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
		fmt.Println("Error reading .shield file, please ensure it exists and is correctly formatted.")
		os.Exit(1)
	}

	shieldIgnorePatterns, err := readPatternsFromFile(".shieldignore")
	if err != nil {
		fmt.Println("Error reading .shieldignore file, please ensure it exists and is correctly formatted.")
		os.Exit(1)
	}

	var filesToEncrypt []string
	for _, pattern := range shieldPatterns {
		fmt.Printf("Looking for files matching pattern: %s\n", pattern)
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("Error while walking file path: %s\n", err)
				return err
			}

			for _, omitPattern := range shieldIgnorePatterns {
				matched, err := filepath.Match(omitPattern, path)
				if err != nil {
					fmt.Printf("Error while matching omit pattern: %s\n", err)
					return err
				}
				if matched {
					fmt.Printf("File: %s matches the omit pattern. Skipping it.\n", path)
					return nil
				}
			}

			matched, err := filepath.Match(pattern, path)
			if err != nil {
				fmt.Printf("Error while matching glob pattern: %s\n", err)
				return err
			}
			if matched && info.Mode().IsRegular() {
				fmt.Printf("Found a file matching the pattern: %s\n", path)
				encrypted, _ := isFileEncrypted(path)
				if !encrypted {
					filesToEncrypt = append(filesToEncrypt, path)
				}
			}
			return nil
		})
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, runtime.NumCPU())
	processFiles(filesToEncrypt, encryptFile, &wg, semaphore)
	wg.Wait()
}

func decryptFiles() {
	shieldPatterns, err := readPatternsFromFile(".shield")
	if err != nil {
		fmt.Println("Error reading .shield file, please ensure it exists and is correctly formatted.")
		os.Exit(1)
	}

	shieldIgnorePatterns, err := readPatternsFromFile(".shieldignore")
	if err != nil {
		fmt.Println("Error reading .shieldignore file, please ensure it exists and is correctly formatted.")
		os.Exit(1)
	}

	var filesToDecrypt []string
	for _, pattern := range shieldPatterns {
		fmt.Printf("Looking for files matching pattern: %s\n", pattern)
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("Error while walking file path: %s\n", err)
				return err
			}

			for _, omitPattern := range shieldIgnorePatterns {
				matched, err := filepath.Match(omitPattern, path)
				if err != nil {
					fmt.Printf("Error while matching omit pattern: %s\n", err)
					return err
				}
				if matched {
					fmt.Printf("File: %s matches the omit pattern. Skipping it.\n", path)
					return nil
				}
			}

			matched, err := filepath.Match(pattern, path)
			if err != nil {
				fmt.Printf("Error while matching glob pattern: %s\n", err)
				return err
			}
			if matched && info.Mode().IsRegular() {
				fmt.Printf("Found a file matching the pattern: %s\n", path)
				encrypted, _ := isFileEncrypted(path)
				if encrypted {
					filesToDecrypt = append(filesToDecrypt, path)
				}
			}
			return nil
		})
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, runtime.NumCPU())
	processFiles(filesToDecrypt, decryptFile, &wg, semaphore)
	wg.Wait()
}

func encryptFile(path string) {
	fmt.Printf("Attempting to encrypt file: %s\n", path)

	cmd := exec.Command("openssl", "enc", "-aes-256-cbc", "-nosalt", "-pass", fmt.Sprintf("file:%s", vaultPasswordFile), "-in", path, "-out", path+".enc")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to encrypt file: %s\n", err)
	} else {
		err = addEncryptionTag(path + ".enc") // add tag to encrypted file
		if err != nil {
			fmt.Printf("Failed to add encryption tag: %s\n", err)
			os.Remove(path + ".enc") // remove failed encrypted file
		} else {
			fmt.Printf("Encrypted file: %s\n", path)
			os.Remove(path)
			os.Rename(path+".enc", path)
		}
	}
}

func decryptFile(path string) {
	fmt.Printf("Attempting to decrypt file: %s\n", path)

	err := removeEncryptionTag(path) // remove tag before decryption
	if err != nil {
		fmt.Printf("Failed to remove encryption tag: %s\n", err)
	} else {
		cmd := exec.Command("openssl", "enc", "-d", "-aes-256-cbc", "-nosalt", "-pass", fmt.Sprintf("file:%s", vaultPasswordFile), "-in", path, "-out", path+".dec")
		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Failed to decrypt file: %s\n", err)
			addEncryptionTag(path) // add tag back if decryption failed
		} else {
			fmt.Printf("Decrypted file: %s\n", path)
			os.Remove(path)
			os.Rename(path+".dec", path)
		}
	}
}

func addEncryptionTag(path string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	content = append([]byte(EncryptionTag), content...)

	return ioutil.WriteFile(path, content, 0666)
}

func removeEncryptionTag(path string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	content = content[EncryptionTagBytes:]

	return ioutil.WriteFile(path, content, 0666)
}

func isFileEncrypted(path string) (bool, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}

	if len(content) >= EncryptionTagBytes && string(content[:EncryptionTagBytes]) == EncryptionTag {
		return true, nil
	}

	return false, nil
}
