package slice

import (
	"os"
	"path/filepath"
	"testing"
	"text/template"

	local "github.com/patrickdappollonio/kubectl-slice/slice/template"
	"github.com/stretchr/testify/require"
)

func TestExecute_processSingleFile(t *testing.T) {
	tests := []struct {
		name          string
		fields        Options
		fileInput     string
		wantErr       bool
		wantFilterErr bool
		fileOutput    *yamlFile
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
			fileOutput: &yamlFile{
				filename: "pod-nginx-ingress.yaml",
				meta: kubeObjectMeta{
					APIVersion: "v1",
					Kind:       "Pod",
					Name:       "nginx-ingress",
				},
			},
		},
		// ----------------------------------------------------------------
		{
			name:      "empty file",
			fields:    Options{},
			fileInput: `---`,
			fileOutput: &yamlFile{
				filename: "-.yaml",
			},
		},
		// ----------------------------------------------------------------
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
			fileOutput: &yamlFile{
				filename: "pod-nginx-ingress.yaml",
				meta: kubeObjectMeta{
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
			fileOutput: &yamlFile{
				filename: "pod-nginx-ingress.yaml",
				meta: kubeObjectMeta{
					APIVersion: "v1",
					Kind:       "Pod",
					Name:       "nginx-ingress",
				},
			},
		},
		// ----------------------------------------------------------------
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
		// ----------------------------------------------------------------
		{
			name:   "non kubernetes file",
			fields: Options{},
			fileInput: `
#
# This is a comment
#
`,
			fileOutput: &yamlFile{
				filename: "-.yaml",
			},
		},
		// ----------------------------------------------------------------
		{
			name:   "file with only spaces",
			fields: Options{},
			fileInput: `
`,
		},
		// ----------------------------------------------------------------
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
		// ----------------------------------------------------------------
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
		// ----------------------------------------------------------------
		{
			name:      "invalid YAML",
			fields:    Options{},
			fileInput: `kind: "Namespace`,
			wantErr:   true,
		},
		// ----------------------------------------------------------------
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
			s := &Split{
				opts:      tt.fields,
				log:       nolog,
				template:  template.Must(template.New("split").Funcs(local.Functions).Parse(DefaultTemplateName)),
				fileCount: 1,
			}

			if err := s.validateFilters(); (err != nil) != tt.wantFilterErr {
				require.Error(t, err)
			}

			if err := s.processSingleFile([]byte(tt.fileInput)); (err != nil) != tt.wantErr {
				require.Error(t, err)
			}

			expectingFile := tt.fileOutput != nil

			if expectingFile {
				require.Lenf(t, s.filesFound, 1, "expected 1 file from list, got %d", len(s.filesFound))

				current := s.filesFound[0]
				require.Equal(t, tt.fileOutput.filename, current.filename)
				require.Equal(t, tt.fileOutput.meta.APIVersion, current.meta.APIVersion)
				require.Equal(t, tt.fileOutput.meta.Kind, current.meta.Kind)
				require.Equal(t, tt.fileOutput.meta.Name, current.meta.Name)
				require.Equal(t, tt.fileOutput.meta.Namespace, current.meta.Namespace)
			} else {
				require.Lenf(t, s.filesFound, 0, "expected 0 files from list, got %d", len(s.filesFound))
			}
		})
	}
}

func TestExecute_writeToFileCases(t *testing.T) {
	tempdir := t.TempDir()
	s := &Split{log: nolog}

	t.Run("write new file", func(tt *testing.T) {
		require.NoError(tt, s.writeToFile(filepath.Join(tempdir, "test.txt"), []byte("test")))
		content, err := os.ReadFile(filepath.Join(tempdir, "test.txt"))
		require.NoError(tt, err)
		require.Equal(tt, "test\n", string(content))
	})

	t.Run("truncate existent file", func(tt *testing.T) {
		preexistent := filepath.Join(tempdir, "test_no_newline.txt")

		require.NoError(tt, os.WriteFile(preexistent, []byte("foobarbaz"), 0644))
		require.NoError(tt, s.writeToFile(preexistent, []byte("test")))

		content, err := os.ReadFile(preexistent)
		require.NoError(tt, err)
		require.Equal(tt, "test\n", string(content))
	})

	t.Run("attempt writing to a read only directory", func(tt *testing.T) {
		require.NoError(tt, os.MkdirAll(filepath.Join(tempdir, "readonly"), 0444))
		require.Error(tt, s.writeToFile(filepath.Join(tempdir, "readonly", "test.txt"), []byte("test")))
	})

	t.Run("attempt writing to a read only sub-directory", func(tt *testing.T) {
		require.NoError(tt, os.MkdirAll(filepath.Join(tempdir, "readonly_sub"), 0444))
		require.Error(tt, s.writeToFile(filepath.Join(tempdir, "readonly_sub", "readonly", "test.txt"), []byte("test")))
	})
}
