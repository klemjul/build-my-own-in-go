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
	binDirAbs, _ := filepath.Abs("../../bin/gitgo")
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

	lsTreeOut, stderr, errcode := RunGitCli(dirName, "ls-tree", "--name-only", strings.TrimSuffix(treeHash, "\n"))
	if errcode != 0 {
		fmt.Println(stderr)
	}

	assert.Equal(t, fmt.Sprintf("%v", "test_dir_1\ntest_file_1.txt\n"), lsTreeOut)
}

func TestLsTree(t *testing.T) {
	dirName := SetupTestDir()
	defer CleanTestDir(dirName)

	RunGitCli(dirName, "init")

	os.WriteFile(dirName+"/test_file_1.txt", []byte("hello world 1"), 0755)
	os.Mkdir(dirName+"/test_dir_1", 0755)
	os.WriteFile(dirName+"/test_dir_1/test_file_2.txt", []byte("hello world 2"), 0755)
	os.WriteFile(dirName+"/test_dir_1/test_file_3.txt", []byte("hello world 3"), 0755)

	RunGitCli(dirName, "add", ".")
	treeHash, _, _ := RunGitCli(dirName, "write-tree")
	lsTreeOut, stderr, errcode := RunMyGitCli(dirName, "ls-tree", "--name-only", strings.TrimSuffix(treeHash, "\n"))
	if errcode != 0 {
		fmt.Println(stderr)
	}

	assert.Equal(t, fmt.Sprintf("%v", "test_dir_1\ntest_file_1.txt\n"), lsTreeOut)
}

func TestCommit(t *testing.T) {
	dirName := SetupTestDir()
	defer CleanTestDir(dirName)

	RunGitCli(dirName, "init")

	os.WriteFile(dirName+"/test_file_1.txt", []byte("hello world 1"), 0755)
	os.Mkdir(dirName+"/test_dir_1", 0755)
	os.WriteFile(dirName+"/test_dir_1/test_file_2.txt", []byte("hello world 2"), 0755)
	os.WriteFile(dirName+"/test_dir_1/test_file_3.txt", []byte("hello world 3"), 0755)

	RunGitCli(dirName, "add", ".")
	treeHash, _, _ := RunGitCli(dirName, "write-tree")
	commitHash, stderr, errcode := RunMyGitCli(dirName, "commit-tree", strings.TrimSuffix(treeHash, "\n"), "-m", "Initial commit")
	if errcode != 0 {
		fmt.Println(stderr)
	}

	showRes, stderr, errcode := RunGitCli(dirName, "show", strings.TrimSuffix(commitHash, "\n"))
	if errcode != 0 {
		fmt.Println(stderr)
	}

	assert.Contains(t, showRes, "+hello world 1")
	assert.Contains(t, showRes, "+hello world 2")
	assert.Contains(t, showRes, "Initial commit")
	assert.Contains(t, showRes, commitHash)

	os.Remove(dirName + "/test_file_1.txt")
	os.WriteFile(dirName+"/test_file_1.txt", []byte("hello world 9"), 0755)

	RunGitCli(dirName, "add", ".")
	treeHash, _, _ = RunGitCli(dirName, "write-tree")

	newCommitHash, stderr, errcode := RunMyGitCli(dirName, "commit-tree", strings.TrimSuffix(treeHash, "\n"), "-m", "Second commit", "-p", commitHash)
	if errcode != 0 {
		fmt.Println(stderr)
	}

	showRes, stderr, errcode = RunGitCli(dirName, "show", strings.TrimSuffix(newCommitHash, "\n"))
	if errcode != 0 {
		fmt.Println(stderr)
	}

	assert.Contains(t, showRes, "+hello world 9")
	assert.Contains(t, showRes, "Second commit")
	assert.Contains(t, showRes, newCommitHash)
}

func TestClone(t *testing.T) {
	dirName := SetupTestDir()
	defer CleanTestDir(dirName)

	stdout, stderr, errcode := RunMyGitCli(dirName, "clone", "https://github.com/codecrafters-io/git-sample-1")
	if errcode != 0 {
		fmt.Println(stderr)
	}
	assert.Contains(t, stdout, "Receiving objects: (329,329), done.")
	assert.Contains(t, stdout, "Receiving deltas: (3,3), done.")

	stdout, stderr, errcode = RunGitCli(fmt.Sprintf("%s/git-sample-1", dirName), "cat-file", "-p", "47b37f1a82bfe85f6d8df52b6258b75e4343b7fd")
	if errcode != 0 {
		fmt.Println(stderr)
	}
	assert.Contains(t, stdout, "Get back to version 1")

}
