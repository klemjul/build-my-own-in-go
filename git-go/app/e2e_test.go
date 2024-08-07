package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func SetupTestDir() string {
	dirName := os.TempDir() + "/" + uuid.New().String()
	err := os.Mkdir(dirName, os.ModeTemporary)
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

func RunGitCli(args []string) string {
	// Redirect os.Stdout to a pipe
	originalStdout := os.Stdout
	readStdout, writeStdout, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stdout = writeStdout

	// Update os.Args
	originalArgs := os.Args
	os.Args = args

	main()

	// Restore original Stdout and Args
	os.Stdout = originalStdout
	os.Args = originalArgs

	// Write and return Stdout pipe
	writeStdout.Close()
	var buf bytes.Buffer
	io.Copy(&buf, readStdout)
	return buf.String()
}

func TestCli(t *testing.T) {
	dirName := SetupTestDir()
	defer CleanTestDir(dirName)

	stdout := RunGitCli([]string{"./git-go", "arg1", "arg2"})
	assert.Equal(t, stdout, "[./git-go arg1 arg2]\n", "they should be equal")
}
