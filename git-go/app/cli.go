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
	command := os.Args[1]

	wd, err := os.Getwd()
	handleError(err)
	repo := Repository{
		RootName: wd + "/.git",
	}

	switch command {
	case "init":
		err = repo.Init()
		if err != nil {
			handleError(err)
		}
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
		os.Stdout.WriteString(file)
	case "hash-object":
		flagSet := flag.NewFlagSet(command, flag.ExitOnError)
		var hashFileW string
		flagSet.StringVar(&hashFileW, "w", "", "file path")
		flagSet.Parse(os.Args[2:])
		if hashFileW == "" {
			handleError(errors.New("please provide -W flag with file path"))
		}

		hash, err := repo.HashFile(hashFileW)
		if err != nil {
			handleError(err)
		}
		os.Stdout.WriteString(hash)
	default:
		handleError(errors.New("Unknown command"))
	}

	os.Exit(0)

}
