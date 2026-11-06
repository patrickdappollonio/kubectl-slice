package slice

import (
	"testing"
	"text/template"

	local "github.com/patrickdappollonio/kubectl-slice/slice/template"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
)

func Test_readLabels(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		strict   bool
		want     map[string]string
		wantErr  bool
	}{
		{
			name:     "no metadata",
			contents: "apiVersion: v1\nkind: Service\n",
			want:     map[string]string{},
		},
		{
			name:     "no labels",
			contents: "metadata:\n  name: foo\n",
			want:     map[string]string{},
		},
		{
			name:     "null labels",
			contents: "metadata:\n  labels:\n",
			want:     map[string]string{},
		},
		{
			name:     "string values",
			contents: "metadata:\n  labels:\n    app: web\n    version: \"1.10\"\n",
			want:     map[string]string{"app": "web", "version": "1.10"},
		},
		{
			name: "non-string scalars keep their source text",
			contents: `metadata:
  labels:
    float: 1.10
    hex: 0x1F
    exp: 1e3
    enabled: true
`,
			want: map[string]string{"float": "1.10", "hex": "0x1F", "exp": "1e3", "enabled": "true"},
		},
		{
			name:     "null values are empty strings",
			contents: "metadata:\n  labels:\n    a:\n    b: ~\n    c: null\n",
			want:     map[string]string{"a": "", "b": "", "c": ""},
		},
		{
			name:     "null values are empty strings in strict mode",
			contents: "metadata:\n  labels:\n    a: ~\n",
			strict:   true,
			want:     map[string]string{"a": ""},
		},
		{
			name:     "aliases are followed",
			contents: "x: &v web\nmetadata:\n  labels:\n    app: *v\n",
			want:     map[string]string{"app": "web"},
		},
		{
			name:     "number rejected in strict mode",
			contents: "metadata:\n  labels:\n    version: 1.10\n",
			strict:   true,
			wantErr:  true,
		},
		{
			name:     "boolean rejected in strict mode",
			contents: "metadata:\n  labels:\n    enabled: true\n",
			strict:   true,
			wantErr:  true,
		},
		{
			name:     "list rejected",
			contents: "metadata:\n  labels:\n    app: [a, b]\n",
			wantErr:  true,
		},
		{
			name:     "map rejected",
			contents: "metadata:\n  labels:\n    app:\n      a: b\n",
			wantErr:  true,
		},
		{
			name:     "labels that are not a map are rejected",
			contents: "metadata:\n  labels: foo\n",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := readLabels([]byte(tt.contents), tt.strict)
			requireErrorIf(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSplit_parseYAMLManifestWithSelector(t *testing.T) {
	const (
		matching = `apiVersion: v1
kind: Service
metadata:
  name: web
  labels:
    app: web
    env: prod
`
		notMatching = `apiVersion: v1
kind: Service
metadata:
  name: db
  labels:
    app: db
`
		unquotedNumber = `apiVersion: v1
kind: Service
metadata:
  name: web
  labels:
    app: web
    version: 1.10
`
		listValue = `apiVersion: v1
kind: Service
metadata:
  name: web
  labels:
    app: [web]
`
	)

	tests := []struct {
		name       string
		contents   string
		selector   string
		strictKube bool
		wantName   string
		wantSkip   bool
		wantStrict bool
		wantErr    bool
	}{
		{
			name:     "set-based selector matches",
			contents: matching,
			selector: "app=web,env in (prod,staging)",
			wantName: "service-web.yaml",
		},
		{
			name:     "non-matching resource is skipped",
			contents: notMatching,
			selector: "app=web",
			wantSkip: true,
		},
		{
			name:     "negative selector keeps unlabeled resource",
			contents: "apiVersion: v1\nkind: Service\nmetadata:\n  name: web\n",
			selector: "app!=db",
			wantName: "service-web.yaml",
		},
		{
			name:     "number matches its source text",
			contents: unquotedNumber,
			selector: "version=1.10",
			wantName: "service-web.yaml",
		},
		{
			name:     "number does not match its decoded value",
			contents: unquotedNumber,
			selector: "version=1.1",
			wantSkip: true,
		},
		{
			name:       "number skips the resource with skip-non-k8s",
			contents:   unquotedNumber,
			selector:   "app=web",
			strictKube: true,
			wantStrict: true,
		},
		{
			name:     "list value is an error",
			contents: listValue,
			selector: "app=web",
			wantErr:  true,
		},
		{
			name:       "list value skips the resource with skip-non-k8s",
			contents:   listValue,
			selector:   "app=web",
			strictKube: true,
			wantStrict: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &Split{
				log:      nolog,
				template: template.Must(template.New(DefaultTemplateName).Funcs(local.Functions).Parse(DefaultTemplateName)),
			}
			s.opts.StrictKubernetes = tt.strictKube

			selector, err := labels.Parse(tt.selector)
			require.NoError(t, err)
			s.selector = selector

			got, err := s.parseYAMLManifest([]byte(tt.contents))

			switch {
			case tt.wantSkip:
				var skip *skipErr
				require.ErrorAs(t, err, &skip)
			case tt.wantStrict:
				var skip *strictModeSkipErr
				require.ErrorAs(t, err, &skip)
			case tt.wantErr:
				require.Error(t, err)
				require.NotErrorAs(t, err, new(*skipErr))
				require.NotErrorAs(t, err, new(*strictModeSkipErr))
			default:
				require.NoError(t, err)
				require.Equal(t, tt.wantName, got.filename)
			}
		})
	}
}

func TestSplit_validateFiltersSelector(t *testing.T) {
	tests := []struct {
		name         string
		selector     string
		wantSelector bool
		wantErr      bool
	}{
		{name: "empty selector", selector: ""},
		{name: "whitespace selector", selector: "   "},
		{name: "valid selector", selector: "app=web,env notin (dev)", wantSelector: true},
		{name: "invalid selector", selector: "app==(", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Split{opts: Options{Selector: tt.selector}}
			err := s.validateFilters()
			requireErrorIf(t, tt.wantErr, err)
			require.Equal(t, tt.wantSelector, s.selector != nil)
		})
	}
}
