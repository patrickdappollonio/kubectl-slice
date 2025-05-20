package files

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// LoadFolder reads the folder contents recursively for files with specified extensions
// and returns a buffer with the contents of all files found separated by `---`
func LoadFolder(extensions []string, folderPath string, recurse bool) (*bytes.Buffer, int, error) {
	var buffer bytes.Buffer
	var count int

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if path != folderPath && !recurse {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if inArray(ext, extensions) {
			count++

			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if buffer.Len() > 0 {
				buffer.WriteString("\n---\n")
			}

			buffer.Write(data)
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	if buffer.Len() == 0 {
		return nil, 0, fmt.Errorf("no files found in %q with extensions: %s", folderPath, strings.Join(extensions, ", "))
	}

	return &buffer, count, nil
}

// LoadFile reads a file and returns its contents as a buffer
func LoadFile(fp string) (*bytes.Buffer, error) {
	f, err := OpenFile(fp)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, f); err != nil {
		return nil, fmt.Errorf("unable to read file %q: %s", fp, err.Error())
	}

	return &buf, nil
}

// OpenFile opens a file for reading, with special handling for stdin
func OpenFile(fp string) (*os.File, error) {
	if fp == os.Stdin.Name() {
		return os.Stdin, nil
	}

	f, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %q: %s", fp, err.Error())
	}

	return f, nil
}

// DeleteFolderContents removes all files in a directory
func DeleteFolderContents(location string) error {
	f, err := os.Open(location)
	if err != nil {
		return fmt.Errorf("unable to open folder %q: %s", location, err.Error())
	}
	defer f.Close()

	names, err := f.Readdirnames(-1)
	if err != nil {
		return fmt.Errorf("unable to read folder %q: %s", location, err.Error())
	}

	for _, name := range names {
		if err := os.RemoveAll(location + "/" + name); err != nil {
			return fmt.Errorf("unable to remove %q: %s", name, err.Error())
		}
	}

	return nil
}

// inArray checks if an element exists in a slice
func inArray[T comparable](needle T, haystack []T) bool {
	return slices.Contains(haystack, needle)
}
