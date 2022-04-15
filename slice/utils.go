package slice

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func loadfile(fp string) (*bytes.Buffer, error) {
	tmplfile, err := filepath.Abs(fp)
	if err != nil {
		return nil, fmt.Errorf("unable to get path to file %q: %s", fp, err.Error())
	}

	f, err := getFile(tmplfile)
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

func getFile(fp string) (*os.File, error) {
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
