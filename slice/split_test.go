package slice

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/patrickdappollonio/kubectl-slice/pkg/logger"
	"github.com/patrickdappollonio/kubectl-slice/pkg/template"
	"github.com/stretchr/testify/require"
)

func TestEndToEnd(t *testing.T) {
	cases := []struct {
		name          string
		inputFile     string
		template      string
		expectedFiles []string
	}{
		{
			name:      "end to end",
			inputFile: "full.yaml",
			template:  template.DefaultTemplateName,
			expectedFiles: []string{
				"full/-.yaml",
				"full/deployment-hello-docker.yaml",
				"full/ingress-hello-docker-ing.yaml",
				"full/service-hello-docker-svc.yaml",
			},
		},
		{
			name:          "basic file, non-k8s",
			inputFile:     "simple-no-k8s.yaml",
			template:      "example.yaml",
			expectedFiles: []string{"simple-no-k8s/example.yaml"},
		},
		{
			name:          "basic file, non-k8s, CRLF line endings",
			inputFile:     "simple-no-k8s-crlf.yaml",
			template:      "example.yaml",
			expectedFiles: []string{"simple-no-k8s/example.yaml"},
		},
	}

	for _, v := range cases {
		t.Run(v.name, func(tt *testing.T) {
			dir := t.TempDir()

			opts := Options{
				InputFile:       filepath.Join("testdata", v.inputFile),
				OutputDirectory: dir,
				GoTemplate:      v.template,
				Stdout:          io.Discard,
				Stderr:          io.Discard,
			}

			slice, err := New(opts)
			require.NoError(tt, err, "not expecting an error")
			require.NoError(tt, slice.Execute(), "not expecting an error on Execute()")

			slice.log = logger.NOOPLogger

			files, err := os.ReadDir(dir)
			require.NoError(tt, err, "not expecting an error on ReadDir()")

			converted := make(map[string]string)
			for _, v := range files {
				if v.IsDir() {
					continue
				}

				if !strings.HasSuffix(v.Name(), ".yaml") {
					continue
				}

				f, err := os.ReadFile(dir + "/" + v.Name())
				if err != nil {
					tt.Fatalf("unable to read file %q: %s", v.Name(), err.Error())
				}

				converted[v.Name()] = string(f)
			}

			require.Equalf(tt, len(v.expectedFiles), len(converted), "required vs converted file length mismatch")

			for _, originalName := range v.expectedFiles {
				originalValue, err := os.ReadFile(filepath.Join("testdata", originalName))
				require.NoError(tt, err, "not expecting an error while reading testdata file")

				found := false
				currentFile := ""

				for receivedName, receivedValue := range converted {
					if filepath.Base(originalName) == receivedName {
						found = true
						currentFile = receivedValue
						break
					}
				}

				if !found {
					tt.Fatalf("expecting to find file %q but it was not found", originalName)
				}

				require.Equalf(tt, string(originalValue), currentFile, "on file %q", originalName)
			}
		})
	}
}
