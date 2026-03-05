package template

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		templateString string
		wantErr        bool
	}{
		{
			name:           "empty template string uses default",
			templateString: "",
			wantErr:        false,
		},
		{
			name:           "valid template string",
			templateString: "{{.kind}}-{{.metadata.name}}",
			wantErr:        false,
		},
		{
			name:           "invalid template string",
			templateString: "{{.kind",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer, err := New(tt.templateString)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, renderer)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, renderer)

			if tt.templateString == "" {
				require.NotNil(t, renderer.tmpl)
			} else {
				require.NotNil(t, renderer.tmpl)
			}
		})
	}
}

func TestRenderer_Execute(t *testing.T) {
	tests := []struct {
		name           string
		templateString string
		data           map[string]any
		want           string
		wantErr        bool
	}{
		{
			name:           "simple template",
			templateString: "{{.kind}}-{{.metadata.name}}",
			data: map[string]any{
				"kind": "Deployment",
				"metadata": map[string]any{
					"name": "nginx",
				},
			},
			want:    "Deployment-nginx",
			wantErr: false,
		},
		{
			name:           "handle missing fields",
			templateString: "{{.kind}}-{{.metadata.name}}",
			data: map[string]any{
				"kind": "Deployment",
				// metadata is missing
			},
			want:    "Deployment-",
			wantErr: false,
		},
		{
			name:           "error accessing non-existent field",
			templateString: "{{.kind}}-{{.missing.field}}",
			data: map[string]any{
				"kind": "Deployment",
			},
			want:    "Deployment-",
			wantErr: false,
		},
		{
			name:           "handle <no value> replacement",
			templateString: "{{.kind}}-{{.nonexistent}}",
			data: map[string]any{
				"kind": "Deployment",
			},
			want:    "Deployment-",
			wantErr: false,
		},
		{
			name:           "trim spaces",
			templateString: "{{ .kind }}-{{ .metadata.name }}",
			data: map[string]any{
				"kind": "Deployment",
				"metadata": map[string]any{
					"name": "nginx",
				},
			},
			want:    "Deployment-nginx",
			wantErr: false,
		},
		{
			name:           "handle line breaks",
			templateString: "{{.kind}}\n{{.metadata.name}}",
			data: map[string]any{
				"kind": "Deployment",
				"metadata": map[string]any{
					"name": "nginx",
				},
			},
			want:    "Deploymentnginx",
			wantErr: false,
		},
		{
			name:           "with template functions",
			templateString: "{{.kind | lower}}-{{.metadata.name}}",
			data: map[string]any{
				"kind": "Deployment",
				"metadata": map[string]any{
					"name": "nginx",
				},
			},
			want:    "deployment-nginx",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer, err := New(tt.templateString)
			require.NoError(t, err)

			got, err := renderer.Execute(tt.data)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRenderer_ExecuteErrorHandling(t *testing.T) {
	// Create a template that will trigger an error when a non-map is treated as a map
	tmpl, err := template.New("filename").Funcs(GetTemplateFunctions()).Parse("{{ .metadata.name.invalid }}")
	require.NoError(t, err)

	renderer := &Renderer{tmpl: tmpl}

	// Create data with metadata.name as a string (not a map)
	data := map[string]any{
		"metadata": map[string]any{
			"name": "test-name", // This is a string, not a map
		},
	}

	// Execute should fail because we're trying to access .invalid on a string value
	result, err := renderer.Execute(data)
	
	// Verify the error behavior
	require.Equal(t, "", result)
	require.Error(t, err)
	require.Contains(t, err.Error(), "can't evaluate field invalid")
	require.Contains(t, err.Error(), "this usually means the field does not exist in the YAML")
}
