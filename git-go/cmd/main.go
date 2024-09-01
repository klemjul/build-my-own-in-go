package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/klemjul/build-my-own-in-go/git-go/internal"
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
	local := internal.LocalRepository{
		RootName: wd,
	}

	switch command {
	case "init":
		err = local.Init()
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
		file, err := local.CatFile(p)
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
		_, hashHex, err := local.WriteBlobObject(w)
		handleError(err)
		fmt.Printf("%v\n", hashHex)
	case "write-tree":
		_, hashHex, err := local.WriteTreeObject(local.RootName)
		handleError(err)
		fmt.Printf("%v\n", hashHex)
	case "ls-tree":
		lstree := flag.NewFlagSet("ls-tree", flag.ExitOnError)
		nameOnly := lstree.Bool("name-only", true, "get only the file name")
		lstree.Parse(os.Args[2:])
		hashHex := os.Args[len(os.Args)-1]
		treeNames, err := local.ReadTreeObject(hashHex, *nameOnly)
		handleError(err)

		fmt.Printf("%v\n", strings.Join(treeNames, "\n"))
	case "commit-tree":
		committree := flag.NewFlagSet("commit-tree", flag.ExitOnError)
		var p, m string
		committree.StringVar(&p, "p", "", "previous commit hash")
		committree.StringVar(&m, "m", "", "commit message")
		committree.Parse(os.Args[3:])
		treeHash := os.Args[2]
		commitHash, err := local.WriteCommitObject(treeHash, p, m)
		handleError(err)

		fmt.Printf("%v\n", commitHash)
	case "clone":
		if len(os.Args) < 3 {
			handleError(errors.New("no clone url provided"))
		}
		parsedUrl, err := url.Parse(os.Args[2])
		if err != nil {
			handleError(nil)
		}

		pathParts := strings.Split(parsedUrl.Path, "/")
		projectName := pathParts[len(pathParts)-1]

		remote, err := internal.NewRemoteRepository(parsedUrl.String())
		handleError(err)
		res, err := remote.DiscoveringReferences()
		handleError(err)

		local = internal.LocalRepository{
			RootName: filepath.Join(wd, projectName),
		}

		fmt.Printf("Cloning into '%s'...\n", projectName)
		err = os.Mkdir(projectName, 0755)
		if err != nil {
			handleError(err)
		}
		local.Init()

		objects, deltas, err := remote.UploadPack([]string{res[0].Ref})
		handleError(err)

		fmt.Printf("remote: Enumerating objects: %v, done.\n", len(objects))
		for i := range objects {
			fmt.Printf("Receiving objects: (%v,%v), done.\n", i+1, len(objects))
			local.WriteObjectWithType(objects[i].ObjectName, objects[i].Content)
		}
		for i := range deltas {
			fmt.Printf("Receiving deltas: (%v,%v), done.\n", i+1, len(deltas))
			err := internal.ApplyObjectDelta(local, deltas[i])
			handleError(err)
		}
	default:
		handleError(errors.New("unknown command"))
	}

	os.Exit(0)

}
