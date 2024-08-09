package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func SetupTestDir() string {
	dirName := os.TempDir() + "/" + uuid.New().String()
	err := os.Mkdir(dirName, 0777)
	if err != nil {
		panic(err)
	}
	return dirName
}

func CleanTestDir(dirName string) {
	err := os.RemoveAll(dirName)
	if err != nil {
		panic(err)
	}
}

func RunGitCli(dirName string, args ...string) (string, string, int) {
	// Define the command to be executed
	binDirAbs, _ := filepath.Abs("../../bin/git-go/app")

	cmd := exec.Command(binDirAbs, args...)
	cmd.Dir = dirName

	// Create buffers to capture stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// Run the command
	err := cmd.Run()

	var exitCode = 0
	// Check for errors, including non-zero exit status
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			fmt.Println(err)
			exitCode = 999
		}
	}
	return stdoutBuf.String(), stderrBuf.String(), exitCode
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func TestGitInit(t *testing.T) {
	dirName := SetupTestDir()
	defer CleanTestDir(dirName)

	gitFolderName := dirName + "/.git"
	gitObjectName := dirName + "/.git/objects"
	gitRefsName := dirName + "/.git/refs"
	gitHeadName := dirName + "/.git/HEAD"

	RunGitCli(dirName, "init")

	assert.True(t, fileExists(gitFolderName), gitFolderName+" should exist")
	assert.True(t, fileExists(gitObjectName), gitObjectName+" should exist")
	assert.True(t, fileExists(gitRefsName), gitRefsName+" should exist")
	assert.True(t, fileExists(gitHeadName), gitHeadName+" should exist")

	file, _ := os.ReadFile(gitHeadName)
	assert.Equal(t, "ref: refs/heads/main\n", string(file))
}

func TestUnknownCommand(t *testing.T) {
	dirName := SetupTestDir()
	defer CleanTestDir(dirName)

	_, stderr, exitCode := RunGitCli(dirName, "unknown")

	assert.Equal(t, 1, exitCode)
	assert.Equal(t, "Unknown command\n", stderr)
}

func TestCatFile(t *testing.T) {
	dirName := SetupTestDir()
	defer CleanTestDir(dirName)

	RunGitCli(dirName, "init")

	hasher := sha1.New()
	initialFile := "Hello world !"
	hasher.Write([]byte(initialFile))
	hash := hex.EncodeToString(hasher.Sum(nil))

	var compressedData bytes.Buffer
	writer := zlib.NewWriter(&compressedData)
	fileContent := "blob\x20" + strconv.Itoa(len(initialFile)) + "\x00" + initialFile
	writer.Write([]byte(fileContent))
	writer.Close()
	hashDir := dirName + "/.git/objects/" + hash[:2]
	os.Mkdir(hashDir, 0755)
	os.WriteFile(hashDir+"/"+hash[2:], compressedData.Bytes(), 0755)

	stdout, _, _ := RunGitCli(dirName, "cat-file", "-p", hash)

	assert.Equal(t, string(initialFile), stdout)
}
