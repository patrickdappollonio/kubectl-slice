package slice

import (
	"log"
	"os"
	"testing"
	"text/template"
)

func TestSplit_processSingleFile(t *testing.T) {
	tests := []struct {
		name       string
		fields     Options
		fileInput  string
		wantErr    bool
		fileOutput *yamlFile
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
				meta:     kubeObjectMeta{},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Split{
				opts:     tt.fields,
				log:      log.New(os.Stderr, "", log.LstdFlags),
				template: template.Must(template.New("split").Funcs(templateFuncs).Parse(DefaultTemplateName)),
			}

			if err := s.processSingleFile([]byte(tt.fileInput)); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			expectingFile := tt.fileOutput != nil

			if expectingFile && len(s.filesFound) != 1 {
				t.Errorf("expected 1 file from list, got %d", len(s.filesFound))
			}

			if expectingFile {
				current := s.filesFound[0]

				if current.filename != tt.fileOutput.filename {
					t.Errorf("expected filename %s, got %s", tt.fileOutput.filename, current.filename)
				}

				if current.meta.APIVersion != tt.fileOutput.meta.APIVersion {
					t.Errorf("expected apiVersion %s, got %s", tt.fileOutput.meta.APIVersion, current.meta.APIVersion)
				}

				if current.meta.Kind != tt.fileOutput.meta.Kind {
					t.Errorf("expected kind %s, got %s", tt.fileOutput.meta.Kind, current.meta.Kind)
				}

				if current.meta.Name != tt.fileOutput.meta.Name {
					t.Errorf("expected name %s, got %s", tt.fileOutput.meta.Name, current.meta.Name)
				}

				if current.meta.Namespace != tt.fileOutput.meta.Namespace {
					t.Errorf("expected namespace %s, got %s", tt.fileOutput.meta.Namespace, current.meta.Namespace)
				}
			} else {
				if len(s.filesFound) != 0 {
					t.Errorf("expected 0 files from list, got %d: %s", len(s.filesFound), s.filesFound[0].filename)
				}
			}
		})
	}
}
