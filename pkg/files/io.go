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

// LoadFolder reads contents from files with matching extensions in the specified folder.
// Returns a buffer with all file contents concatenated with "---" separators between them,
// a count of files processed, and any error encountered.
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

// LoadFile reads a file from the filesystem and returns its contents as a buffer.
// Handles errors for file access issues.
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

// OpenFile opens a file for reading with special handling for stdin.
// When the filename is "-", it returns os.Stdin instead of attempting to open a file.
func OpenFile(fp string) (*os.File, error) {
	if fp == os.Stdin.Name() || fp == "-" {
		return os.Stdin, nil
	}

	f, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %q: %s", fp, err.Error())
	}

	return f, nil
}

// DeleteFolderContents removes all files and subdirectories within the specified directory.
// The directory itself is preserved.
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
