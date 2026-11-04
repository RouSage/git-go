package plumbing

import (
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
		return "", fmt.Errorf("error reading file: %w", err)
	}

	// blob object format: blob <size>\x00<content>
	parts := strings.Split(string(content), "\x00")
	return parts[1], nil
}

func HashObject(path string, isWrite bool) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	object := fmt.Sprintf("blob %d\x00%s", len(content), content)
	hash := fmt.Sprintf("%x", sha1.Sum([]byte(object)))

	if isWrite {
		path := fmt.Sprintf("%s/objects/%s/%s", GIT_DIR, hash[:2], hash[2:])

		err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("error creating directory: %w", err)
		}

		file, err := os.Create(path)
		if err != nil {
			return "", fmt.Errorf("error creating file: %w", err)
		}
		defer file.Close()

		w := zlib.NewWriter(file)
		w.Write([]byte(object))
		w.Close()
	}

	return hash, nil
}
