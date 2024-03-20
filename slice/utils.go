package slice

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

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
