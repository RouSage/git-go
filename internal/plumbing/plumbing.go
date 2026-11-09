package plumbing

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	GIT_DIR     = ".git"
	GIT_OBJECTS = "objects"
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
	content, err := readGitObject(objHash)
	if err != nil {
		return "", err
	}

	// blob object format: blob <size>\x00<content>
	parts := strings.Split(string(content), "\x00")
	return parts[1], nil
}

func HashObject(file string, write bool) (string, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	object := fmt.Sprintf("blob %d\x00%s", len(content), content)
	hash := fmt.Sprintf("%x", sha1.Sum([]byte(object)))

	if write {
		path := path.Join(GIT_DIR, GIT_OBJECTS, hash[:2], hash[2:])

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

func LsTree(objHash string, nameOnly bool) (string, error) {
	data, err := readGitObject(objHash)
	if err != nil {
		return "", err
	}

	entries, err := parseTree(data)
	if err != nil {
		return "", err
	}

	output := ""
	for _, entry := range entries {
		if nameOnly {
			output += fmt.Sprintln(entry.name)
		} else {
			output += fmt.Sprintf("%d %s %s\t%s\n", entry.mode, entryType(entry.mode), entry.hash, entry.name)
		}
	}

	return output, nil
}

type treeEntryMode int

const (
	modeRegularFile    treeEntryMode = 100644
	modeExecutableFile treeEntryMode = 100755
	modeSymLink        treeEntryMode = 120000
	modeDirectory      treeEntryMode = 40000
)

type treeEntry struct {
	mode treeEntryMode
	name string
	hash string
}

func parseTree(data []byte) ([]treeEntry, error) {
	entries := []treeEntry{}

	if len(data) == 0 {
		return entries, nil
	}

	objHeader, content, ok := bytes.Cut(data, []byte{0})
	if !ok {
		return nil, fmt.Errorf("invalid tree object format")
	}
	objType, _, ok := bytes.Cut(objHeader, []byte(" "))
	if !ok {
		return nil, fmt.Errorf("invalid tree object format")
	}
	if string(objType) != "tree" {
		return nil, fmt.Errorf("not a tree object")
	}

	// each tree entry is in the format: <mode> <name>\0<20_byte_sha>
	for len(content) > 0 {
		header, rest, ok := bytes.Cut(content, []byte{0})
		if !ok || len(rest) < sha1.Size {
			return nil, fmt.Errorf("invalid tree entry format: %q", content)
		}

		modeStr, name, ok := bytes.Cut(header, []byte(" "))
		if !ok {
			return nil, fmt.Errorf("missing mode in tree entry: %q", header)
		}
		mode, err := strconv.Atoi(string(modeStr))
		if err != nil {
			return nil, fmt.Errorf("invalid mode in tree entry: %q", modeStr)
		}

		entry := treeEntry{
			mode: treeEntryMode(mode),
			name: string(name),
			hash: hex.EncodeToString(rest[:sha1.Size]),
		}
		entries = append(entries, entry)

		content = rest[sha1.Size:]
	}

	return entries, nil
}

func entryType(mode treeEntryMode) string {
	switch mode {
	case modeDirectory:
		return "tree"
	default:
		return "blob"
	}
}

func readGitObject(objHash string) ([]byte, error) {
	hashPrefix := objHash[:2]
	hashSuffix := objHash[2:]

	path := path.Join(GIT_DIR, GIT_OBJECTS, hashPrefix, hashSuffix)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	r, err := zlib.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("error decompressing file: %w", err)
	}
	defer r.Close()

	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return content, nil
}
