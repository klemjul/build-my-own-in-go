package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
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
		catfile := flag.NewFlagSet("cat-file", flag.ExitOnError)
		var p string
		catfile.StringVar(&p, "p", "", "sha1 object hash")
		catfile.Parse(os.Args[2:])
		if p == "" {
			handleError(errors.New("please provide -p flag with object hash"))
		}
		file, err := repo.CatFile(p)
		if err != nil {
			handleError(err)
		}
		fmt.Printf("%v", file)
	case "hash-object":
		hashobject := flag.NewFlagSet("hash-object", flag.ExitOnError)
		var w string
		hashobject.StringVar(&w, "w", "", "file path")
		hashobject.Parse(os.Args[2:])
		if w == "" {
			handleError(errors.New("please provide -w flag with file path"))
		}
		_, hashHex, err := repo.WriteBlobObject(w)
		handleError(err)
		fmt.Printf("%v\n", hashHex)
	case "write-tree":
		_, hashHex, err := repo.WriteTreeObject(repo.RootName)
		handleError(err)
		fmt.Printf("%v\n", hashHex)
	case "ls-tree":
		lstree := flag.NewFlagSet("ls-tree", flag.ExitOnError)
		nameOnly := lstree.Bool("name-only", true, "get only the file name")
		lstree.Parse(os.Args[2:])
		hashHex := os.Args[len(os.Args)-1]
		treeNames, err := repo.ReadTreeObject(hashHex, *nameOnly)
		handleError(err)

		fmt.Printf("%v\n", strings.Join(treeNames, "\n"))
	default:
		handleError(errors.New("unknown command"))
	}

	os.Exit(0)

}
