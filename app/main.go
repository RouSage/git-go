package main

import (
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
		objHash := os.Args[3]

		content, err := plumbing.CatFile(objHash)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		fmt.Print(content)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
