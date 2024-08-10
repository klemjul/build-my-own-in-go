package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func handleError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		handleError(errors.New("no command provided"))
	}
	command := os.Args[1]

	wd, err := os.Getwd()
	handleError(err)
	repo := Repository{
		RootName: wd,
	}

	switch command {
	case "init":
		err = repo.Init()
		handleError(err)
		fmt.Printf("Initialized empty Git repository in %v\n", wd)
	case "cat-file":
		flagSet := flag.NewFlagSet(command, flag.ExitOnError)
		var catFileP string
		flagSet.StringVar(&catFileP, "p", "", "sha1 object hash")
		flagSet.Parse(os.Args[2:])
		if catFileP == "" {
			handleError(errors.New("please provide -p flag with object hash"))
		}
		file, err := repo.CatFile(catFileP)
		if err != nil {
			handleError(err)
		}
		fmt.Printf("%v", file)
	case "hash-object":
		flagSet := flag.NewFlagSet(command, flag.ExitOnError)
		var hashFileW string
		flagSet.StringVar(&hashFileW, "w", "", "file path")
		flagSet.Parse(os.Args[2:])
		if hashFileW == "" {
			handleError(errors.New("please provide -w flag with file path"))
		}
		_, hashHex, err := repo.WriteBlobObject(hashFileW)
		handleError(err)
		fmt.Printf("%v", hashHex)
	case "write-tree":
		_, hashHex, err := repo.WriteTreeObject(repo.RootName)
		handleError(err)
		fmt.Printf("%v", hashHex)
	default:
		handleError(errors.New("unknown command"))
	}

	os.Exit(0)

}
