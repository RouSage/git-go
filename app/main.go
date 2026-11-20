package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/codecrafters-io/git-starter-go/internal/plumbing"
)

// Usage: your_program.sh <command> <arg1> <arg2> ...
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		err := plumbing.Init()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	case "cat-file":
		os.Args = os.Args[1:]

		var print bool
		flag.BoolVar(&print, "p", false, "Print the contents of the object")
		flag.Parse()

		if !print {
			fmt.Println("usage: mygit cat-file -p <hash>")
			os.Exit(1)
		}

		content, err := plumbing.CatFile(flag.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(content)

	case "hash-object":
		os.Args = os.Args[1:]

		var write bool
		flag.BoolVar(&write, "w", false, "Write the object to the objects directory")
		flag.Parse()

		hash, err := plumbing.HashObject(flag.Arg(0), write)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(hash)

	case "ls-tree":
		os.Args = os.Args[1:]

		var nameOnly bool
		flag.BoolVar(&nameOnly, "name-only", false, "List only the names of the tree object")
		flag.Parse()

		tree, err := plumbing.LsTree(flag.Arg(0), nameOnly)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(tree)

	case "write-tree":
		treeHash, err := plumbing.WriteTree()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(treeHash)

	case "commit-tree":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: mygit commit-tree <tree_hash> [-p <parent>] -m <message>")
			os.Exit(1)
		}

		objHash := os.Args[2]
		if len(objHash) == 0 {
			fmt.Fprintln(os.Stderr, "object hash is required")
			os.Exit(1)
		}

		fs := flag.NewFlagSet("commit-tree", flag.ExitOnError)
		parent := fs.String("p", "", "Provide a parent commit hash")
		msg := fs.String("m", "", "Provide a commit message")
		if err := fs.Parse(os.Args[3:]); err != nil {
			fmt.Fprintln(os.Stderr, fmt.Errorf("error parsing arugments: %w", err))
			os.Exit(1)
		}

		if len(*msg) == 0 {
			fmt.Fprintln(os.Stderr, "commit message is required")
			os.Exit(1)
		}

		commitHash, err := plumbing.CommitTree(objHash, *msg, *parent)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(commitHash)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
