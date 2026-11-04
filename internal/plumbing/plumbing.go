package plumbing

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	GIT_DIR = ".git"
)

func Init() error {
	for _, dir := range []string{GIT_DIR, fmt.Sprintf("%s/objects", GIT_DIR), fmt.Sprintf("%s/refs", GIT_DIR)} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
	}

	headFileContents := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	fmt.Println("Initialized git directory")

	return nil
}

func CatFile(objHash string) (string, error) {
	path := fmt.Sprintf("%s/objects/%s/%s", GIT_DIR, objHash[:2], objHash[2:])
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("error opening file: %w", err)
	}

	r, err := zlib.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("error decompressing file: %w", err)
	}
	defer r.Close()

	content, err := io.ReadAll(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
		return "", fmt.Errorf("error reading file: %w", err)
	}

	// blob object format: blob <size>\x00<content>
	parts := strings.Split(string(content), "\x00")
	return parts[1], nil
}
