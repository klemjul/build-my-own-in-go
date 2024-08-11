package test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func RunCli(cliPath string, dirName string, args ...string) (string, string, int) {
	cmd := exec.Command(cliPath, args...)
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

func RunMyGitCli(dirName string, args ...string) (string, string, int) {
	binDirAbs, _ := filepath.Abs("../../bin/git-go/app")
	return RunCli(binDirAbs, dirName, args...)
}

func RunGitCli(dirName string, args ...string) (string, string, int) {
	return RunCli("git", dirName, args...)
}

func TestGitInit(t *testing.T) {
	dirName := SetupTestDir()
	defer CleanTestDir(dirName)

	gitFolderName := dirName + "/.git"
	gitObjectName := dirName + "/.git/objects"
	gitRefsName := dirName + "/.git/refs"
	gitHeadName := dirName + "/.git/HEAD"

	RunMyGitCli(dirName, "init")

	assert.DirExists(t, gitFolderName)
	assert.DirExists(t, gitObjectName)
	assert.DirExists(t, gitRefsName)
	assert.FileExists(t, gitHeadName)

	file, _ := os.ReadFile(gitHeadName)
	assert.Equal(t, "ref: refs/heads/main\n", string(file))
}

func TestUnknownCommand(t *testing.T) {
	dirName := SetupTestDir()
	defer CleanTestDir(dirName)

	_, stderr, exitCode := RunMyGitCli(dirName, "unknown")

	assert.Equal(t, 1, exitCode)
	assert.Equal(t, "unknown command\n", stderr)
}

func TestCatFile(t *testing.T) {
	dirName := SetupTestDir()
	defer CleanTestDir(dirName)

	RunGitCli(dirName, "init")

	fileName := dirName + "/" + "hellofile.txt"
	fileContent := "Hello world !"
	os.WriteFile(fileName, []byte(fileContent), 0755)
	hash, stderr, errcode := RunGitCli(dirName, "hash-object", "-w", fileName)
	if errcode != 0 {
		fmt.Println(stderr)
	}

	stdout, stderr, errcode := RunMyGitCli(dirName, "cat-file", "-p", strings.TrimSuffix(hash, "\n"))
	if errcode != 0 {
		fmt.Println(stderr)
	}

	assert.Equal(t, fileContent, stdout)
}

func TestHashObject(t *testing.T) {
	dirName := SetupTestDir()
	defer CleanTestDir(dirName)

	RunGitCli(dirName, "init")

	fileName := dirName + "/" + "hellofile.txt"
	fileContent := "Hello world !"
	os.WriteFile(fileName, []byte(fileContent), 0755)

	hash, stderr, errcode := RunMyGitCli(dirName, "hash-object", "-w", fileName)
	if errcode != 0 {
		fmt.Println(stderr)
	}

	stdout, stderr, errcode := RunGitCli(dirName, "cat-file", "-p", strings.TrimSuffix(hash, "\n"))
	if errcode != 0 {
		fmt.Println(stderr)
	}

	assert.Equal(t, fileContent, stdout)

}

func TestWriteTree(t *testing.T) {
	dirName := SetupTestDir()
	defer CleanTestDir(dirName)

	RunGitCli(dirName, "init")

	os.WriteFile(dirName+"/test_file_1.txt", []byte("hello world 1"), 0755)
	os.Mkdir(dirName+"/test_dir_1", 0755)
	os.WriteFile(dirName+"/test_dir_1/test_file_2.txt", []byte("hello world 2"), 0755)
	os.WriteFile(dirName+"/test_dir_1/test_file_3.txt", []byte("hello world 3"), 0755)

	treeHash, stderr, errcode := RunMyGitCli(dirName, "write-tree")
	if errcode != 0 {
		fmt.Println(stderr)
	}

	lsTreeOut, stderr, errcode := RunGitCli(dirName, "ls-tree", "--name-only", treeHash)
	if errcode != 0 {
		fmt.Println(stderr)
	}

	assert.Equal(t, fmt.Sprintf("%v", "test_dir_1\ntest_file_1.txt\n"), lsTreeOut)

}
