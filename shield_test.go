package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestShield(t *testing.T) {
	t.Log("----- Creating and Setting Environment -----")
	Version = os.Getenv("VERSION")
	Name = os.Getenv("NAME")
	Author = os.Getenv("AUTHOR")
	Encryption = os.Getenv("ENCRYPTION")
	tmpDir, removeTmpDir := createTempDir(t)
	SetDirectory(tmpDir)
	defer removeTmpDir()
	generatePreCommitHook()

	paths := []string{
		"test1/testfile1.txt",
		"test1/testfile2.secret",
		"test2/testfile3.txt",
		"test2/testfile4.secret",
		"test/temp.secret",
		"vendors/testfile5.secret",
		"secrets/testfile6.txt",
		"secrets/testfile7.pem",
	}
	for _, path := range paths {
		dir, _ := filepath.Split(path)
		fullPath := filepath.Join(tmpDir, path)
		os.MkdirAll(filepath.Join(tmpDir, dir), os.ModePerm)
		os.WriteFile(fullPath, []byte("test"), os.ModePerm)
		t.Logf("Created: %s", fullPath)
	}

	err := os.WriteFile(filepath.Join(tmpDir, ".shieldpass"), []byte("broy"), os.ModePerm)
	if err != nil {
		t.Errorf("Error writing .shield file: %v", err)
	}

	SetPasswordFile(filepath.Join(tmpDir, ".shieldpass"))
	SetEncryptionTag()

	shieldConfig := `**/*.secret
**/secrets/*.txt
secrets/**/*.pem`

	shieldIgnoreConfig := `test/*
temp.secret
**/vendors/**`

	err = os.WriteFile(filepath.Join(tmpDir, ".shield"), []byte(shieldConfig), os.ModePerm)
	if err != nil {
		t.Errorf("Error writing .shield file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, ".shieldignore"), []byte(shieldIgnoreConfig), os.ModePerm)
	if err != nil {
		t.Errorf("Error writing .shieldignore file: %v", err)
	}

	t.Log("----- File List -----")
	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		t.Log(path)
		return nil
	})
	if err != nil {
		t.Errorf("failed to walk the directory: %v", err)
	}

	t.Log("----- Encrypting Files -----")
	handleEncryption()

	expectedEncrypted := []string{
		"test1/testfile2.secret",
		"test2/testfile4.secret",
		"secrets/testfile6.txt",
		"secrets/testfile7.pem",
	}

	expectedUnencrypted := []string{
		"test1/testfile1.txt",
		"test2/testfile3.txt",
		"test/temp.secret",
		"vendors/testfile5.secret",
	}

	t.Log("----- Encryption Check -----")
	for _, path := range expectedEncrypted {
		encrypted, err := isFileEncrypted(path)
		if err != nil {
			fmt.Println(err)
		}
		if !encrypted {
			t.Errorf("file %q was not encrypted", path)
		} else {
			t.Logf("file %s was encrypted", path)
		}
	}

	t.Log("----- Expected Unencrypted Check -----")
	for _, path := range expectedUnencrypted {
		encrypted, err := isFileEncrypted(path)
		if err != nil {
			fmt.Println(err)
		}
		if encrypted {
			t.Errorf("file %q was encrypted but it shouldn't be", path)
		} else {
			t.Logf("file %s was not encrypted as expected", path)
		}
	}

	handleDecryption()

	t.Log("----- Decryption Check -----")
	for _, path := range expectedEncrypted {
		encrypted, err := isFileEncrypted(path)
		if err != nil {
			fmt.Println(err)
		}
		if encrypted {
			t.Errorf("file %q was not decrypted", path)
		} else {
			t.Logf("file %s was decrypted", path)
		}
	}

	t.Log("----- Testing Hook -----")
	// Add files to the Git repository
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not add files to git repository: %v", err)
	}

	// Commit files
	cmd = exec.Command("git", "commit", "-m", "Sensitive files should not be committed")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not commit files: %v", err)
	}

	t.Log("----- Hook Encryption Check -----")
	for _, path := range expectedEncrypted {
		encrypted, err := isFileEncrypted(path)
		if err != nil {
			fmt.Println(err)
		}
		if !encrypted {
			t.Errorf("file %q was not encrypted", path)
		} else {
			t.Logf("file %s was encrypted", path)
		}
	}

	t.Log("----- Hook Expected Unencrypted Check -----")
	for _, path := range expectedUnencrypted {
		encrypted, err := isFileEncrypted(path)
		if err != nil {
			fmt.Println(err)
		}
		if encrypted {
			t.Errorf("file %q was encrypted but it shouldn't be", path)
		} else {
			t.Logf("file %s was not encrypted as expected", path)
		}
	}
}

func createTempDir(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "shield-test")
	if err != nil {
		t.Fatalf("could not create temporary directory: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not initialize git repository: %v", err)
	}

	remove := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, remove
}
