package slice

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func inarray[T comparable](needle T, haystack []T) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}

	return false
}

// loadfolder reads the folder contents recursively for `.yaml` and `.yml` files
// and returns a buffer with the contents of all files found; returns the buffer
// with all the files separated by `---` and the number of files found
func loadfolder(extensions []string, folderPath string, recurse bool) (*bytes.Buffer, int, error) {
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
		if inarray(ext, extensions) {
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

func loadfile(fp string) (*bytes.Buffer, error) {
	f, err := openFile(fp)
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

func openFile(fp string) (*os.File, error) {
	if fp == os.Stdin.Name() {
		// On Windows, the name in Go for stdin is `/dev/stdin` which doesn't
		// exist. It must use the syscall to point to the file and open it
		return os.Stdin, nil
	}

	// Any other file that's not stdin can be opened normally
	f, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %q: %s", fp, err.Error())
	}

	return f, nil
}

func deleteFolderContents(location string) error {
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
