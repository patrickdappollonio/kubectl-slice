package split

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

	f, err := os.Open(tmplfile)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %q: %s", fp, err.Error())
	}

	defer f.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, f); err != nil {
		return nil, fmt.Errorf("unable to read file %q: %s", fp, err.Error())
	}

	return &buf, nil
}
