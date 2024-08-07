package main

import (
	"os"
)

func handleError(err error) {
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
}

func initCommand() {
	wd, err := os.Getwd()
	handleError(err)
	err = os.Mkdir(wd+"/.git", 0755)
	handleError(err)
	err = os.Mkdir(wd+"/.git/objects", 0755)
	handleError(err)
	err = os.Mkdir(wd+"/.git/refs", 0755)
	handleError(err)
	err = os.WriteFile(wd+"/.git/HEAD", []byte("ref: refs/heads/main\n"), 0755)
	handleError(err)
}

func main() {
	command := os.Args[1]

	switch command {
	case "init":
		initCommand()
	default:
		os.Stderr.WriteString("Unknown command")
		os.Exit(1)
	}

}
