package slice

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/patrickdappollonio/kubectl-slice/pkg/kubernetes"
	"github.com/patrickdappollonio/kubectl-slice/pkg/logger"
	"github.com/patrickdappollonio/kubectl-slice/pkg/template"
	"github.com/stretchr/testify/require"
)

func TestExecute_processSingleFile(t *testing.T) {
	type yamlFileOutputTest struct {
		filename string
		meta     kubernetes.ObjectMeta
	}

	tests := []struct {
		name          string
		fields        Options
		fileInput     string
		wantErr       bool
		wantFilterErr bool
		fileOutput    *yamlFileOutputTest
	}{
		{
			name:   "basic pod",
			fields: Options{},
			fileInput: `
apiVersion: v1
kind: Pod
metadata:
  name: nginx-ingress
`,
			fileOutput: &yamlFileOutputTest{
				filename: "pod-nginx-ingress.yaml",
				meta: kubernetes.ObjectMeta{
					APIVersion: "v1",
					Kind:       "Pod",
					Name:       "nginx-ingress",
				},
			},
		},
		{
			name:      "empty file",
			fields:    Options{},
			fileInput: `---`,
			fileOutput: &yamlFileOutputTest{
				filename: "-.yaml",
			},
		},
		{
			name: "include kind",
			fields: Options{
				IncludedKinds: []string{"Pod"},
			},
			fileInput: `
apiVersion: v1
kind: Pod
metadata:
  name: nginx-ingress
`,
			fileOutput: &yamlFileOutputTest{
				filename: "pod-nginx-ingress.yaml",
				meta: kubernetes.ObjectMeta{
					APIVersion: "v1",
					Kind:       "Pod",
					Name:       "nginx-ingress",
				},
			},
		},
		{
			name: "include Pod using include option",
			fields: Options{
				Included: []string{"Pod/*"},
			},
			fileInput: `
apiVersion: v1
kind: Pod
metadata:
  name: nginx-ingress
`,
			fileOutput: &yamlFileOutputTest{
				filename: "pod-nginx-ingress.yaml",
				meta: kubernetes.ObjectMeta{
					APIVersion: "v1",
					Kind:       "Pod",
					Name:       "nginx-ingress",
				},
			},
		},
		{
			name: "non kubernetes files skipped using strict kubernetes",
			fields: Options{
				StrictKubernetes: true,
			},
			fileInput: `
#
# This is a comment
#
`,
		},
		{
			name:   "non kubernetes file",
			fields: Options{},
			fileInput: `
#
# This is a comment
#
`,
			fileOutput: &yamlFileOutputTest{
				filename: "-.yaml",
			},
		},
		{
			name:   "file with only spaces",
			fields: Options{},
			fileInput: `
`,
		},
		{
			name: "skipping kind",
			fields: Options{
				IncludedKinds: []string{"Pod"},
			},
			fileInput: `
apiVersion: v1
kind: Namespace
metadata:
  name: foobar
`,
		},
		{
			name: "skipping name",
			fields: Options{
				IncludedNames: []string{"foofoo"},
			},
			fileInput: `
apiVersion: v1
kind: Namespace
metadata:
  name: foobar
`,
		},
		{
			name:      "invalid YAML",
			fields:    Options{},
			fileInput: `kind: "Namespace`,
			wantErr:   true,
		},
		{
			name:   "invalid YAML",
			fields: Options{},
			fileInput: `
kind: "Namespace
`,
			wantErr: true,
		},
		{
			name: "invalid excluded",
			fields: Options{
				Excluded: []string{"Pod/Namespace/*"},
			},
			wantFilterErr: true,
		},
		{
			name: "invalid included",
			fields: Options{
				Included: []string{"Pod"},
			},
			wantFilterErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &Split{
				opts: tt.fields,
				log:  log.New(io.Discard, "", 0),
				template: func() *template.Renderer {
					tmpl, err := template.New(template.DefaultTemplateName)
					if err != nil {
						t.Fatalf("unable to create template: %s", err)
					}

					return tmpl
				}(),
				fileCount: 1,
			}

			if tt.wantFilterErr {
				require.Error(t, s.validateFilters())
				return
			}
			require.NoError(t, s.validateFilters())

			if tt.wantErr {
				require.Error(t, s.processSingleFile([]byte(tt.fileInput)))
				return
			}
			require.NoError(t, s.processSingleFile([]byte(tt.fileInput)))

			if tt.fileOutput != nil {
				require.Lenf(t, s.filesFound, 1, "expected 1 file from list, got %d", len(s.filesFound))

				current := s.filesFound[0]
				require.Equal(t, tt.fileOutput.filename, current.Filename)
				require.Equal(t, tt.fileOutput.meta.APIVersion, current.Meta.APIVersion)
				require.Equal(t, tt.fileOutput.meta.Kind, current.Meta.Kind)
				require.Equal(t, tt.fileOutput.meta.Name, current.Meta.Name)
				require.Equal(t, tt.fileOutput.meta.Namespace, current.Meta.Namespace)
			} else {
				require.Lenf(t, s.filesFound, 0, "expected 0 files from list, got %d", len(s.filesFound))
			}
		})
	}
}

func TestExecute_writeToFileCases(t *testing.T) {
	tempdir := t.TempDir()
	s := &Split{log: logger.NOOPLogger}

	t.Run("write new file", func(tt *testing.T) {
		t.Parallel()
		require.NoError(tt, s.writeToFile(filepath.Join(tempdir, "test.txt"), []byte("test")))
		content, err := os.ReadFile(filepath.Join(tempdir, "test.txt"))
		require.NoError(tt, err)
		require.Equal(tt, "test\n", string(content))
	})

	t.Run("truncate existent file", func(tt *testing.T) {
		preexistent := filepath.Join(tempdir, "test_no_newline.txt")

		require.NoError(tt, os.WriteFile(preexistent, []byte("foobarbaz"), 0o644))
		require.NoError(tt, s.writeToFile(preexistent, []byte("test")))

		content, err := os.ReadFile(preexistent)
		require.NoError(tt, err)
		require.Equal(tt, "test\n", string(content))
	})

	t.Run("attempt writing to a read only directory", func(tt *testing.T) {
		require.NoError(tt, os.MkdirAll(filepath.Join(tempdir, "readonly"), 0o444))
		require.Error(tt, s.writeToFile(filepath.Join(tempdir, "readonly", "test.txt"), []byte("test")))
	})

	t.Run("attempt writing to a read only sub-directory", func(tt *testing.T) {
		require.NoError(tt, os.MkdirAll(filepath.Join(tempdir, "readonly_sub"), 0o444))
		require.Error(tt, s.writeToFile(filepath.Join(tempdir, "readonly_sub", "readonly", "test.txt"), []byte("test")))
	})
}

func TestAddingTripleDashes(t *testing.T) {
	cases := []struct {
		name          string
		input         string
		includeDashes bool
		output        map[string]string
	}{
		{
			name:   "empty file",
			input:  `---`,
			output: map[string]string{"-.yaml": "---\n"},
		},
		{
			name: "simple no dashes",
			input: `apiVersion: v1
kind: Pod
metadata:
  name: nginx-ingress
---
apiVersion: v1
kind: Namespace
metadata:
  name: production`,
			output: map[string]string{
				"pod-nginx-ingress.yaml":    "apiVersion: v1\nkind: Pod\nmetadata:\n  name: nginx-ingress\n",
				"namespace-production.yaml": "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: production\n",
			},
		},
		{
			name:          "simple with dashes",
			includeDashes: true,
			input: `apiVersion: v1
kind: Pod
metadata:
  name: nginx-ingress
---
apiVersion: v1
kind: Namespace
metadata:
  name: production`,
			output: map[string]string{
				"pod-nginx-ingress.yaml":    "---\napiVersion: v1\nkind: Pod\nmetadata:\n  name: nginx-ingress\n",
				"namespace-production.yaml": "---\napiVersion: v1\nkind: Namespace\nmetadata:\n  name: production\n",
			},
		},
		{
			name:          "simple with dashes - adding empty intermediate files",
			includeDashes: true,
			input: `apiVersion: v1
kind: Pod
metadata:
  name: nginx-ingress
---
---
---
---
apiVersion: v1
kind: Namespace
metadata:
  name: production`,
			output: map[string]string{
				"pod-nginx-ingress.yaml":    "---\napiVersion: v1\nkind: Pod\nmetadata:\n  name: nginx-ingress\n",
				"namespace-production.yaml": "---\napiVersion: v1\nkind: Namespace\nmetadata:\n  name: production\n",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tdinput := t.TempDir()
			tdoutput := t.TempDir()
			require.NotEqual(t, tdinput, tdoutput, "input and output directories should be different")

			err := os.WriteFile(filepath.Join(tdinput, "input.yaml"), []byte(tt.input), 0o644)
			require.NoError(t, err, "error found while writing input file")

			s, err := New(Options{
				GoTemplate:        template.DefaultTemplateName,
				IncludeTripleDash: tt.includeDashes,
				InputFile:         filepath.Join(tdinput, "input.yaml"),
				OutputDirectory:   tdoutput,
				Stderr:            os.Stderr,
				Stdout:            io.Discard,
			})
			require.NoError(t, err, "error found while creating new Split instance")
			require.NoError(t, s.Execute(), "error found while executing slice")

			files, err := os.ReadDir(tdoutput)
			require.NoError(t, err, "error found while reading output directory")

			for _, file := range files {
				content, err := os.ReadFile(filepath.Join(tdoutput, file.Name()))
				require.NoError(t, err, "error found while reading file %q", file.Name())

				expected, found := tt.output[file.Name()]
				require.True(t, found, "expected file %q to be found in the output map", file.Name())
				require.Equal(t, expected, string(content), "expected content to be equal for file %q", file.Name())
			}
		})
	}
}
