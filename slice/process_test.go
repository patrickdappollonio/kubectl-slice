package slice

import (
	"testing"

	"github.com/patrickdappollonio/kubectl-slice/pkg/kubernetes"
	"github.com/patrickdappollonio/kubectl-slice/pkg/logger"
	"github.com/patrickdappollonio/kubectl-slice/pkg/template"
	"github.com/stretchr/testify/require"
)

func Test_inSliceIgnoreCase(t *testing.T) {
	type args struct {
		slice    []string
		expected string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "in slice",
			args: args{
				slice:    []string{"foo", "bar"},
				expected: "foo",
			},
			want: true,
		},
		{
			name: "not in slice",
			args: args{
				slice:    []string{"foo", "bar"},
				expected: "baz",
			},
			want: false,
		},
		{
			name: "pattern fo-star without glob support",
			args: args{
				slice:    []string{"fo*", "bar"},
				expected: "foo",
			},
			want: false,
		},
		{
			name: "pattern anything without glob support",
			args: args{
				slice:    []string{"*"},
				expected: "foo",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, inSliceIgnoreCase(tt.args.slice, tt.args.expected))
		})
	}
}

func Test_inSliceIgnoreCaseGlob(t *testing.T) {
	type args struct {
		slice    []string
		expected string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "in slice",
			args: args{
				slice:    []string{"foo", "bar"},
				expected: "foo",
			},
			want: true,
		},
		{
			name: "not in slice",
			args: args{
				slice:    []string{"foo", "bar"},
				expected: "baz",
			},
			want: false,
		},
		{
			name: "pattern fo-star",
			args: args{
				slice:    []string{"fo*", "bar"},
				expected: "foo",
			},
			want: true,
		},
		{
			name: "pattern anything",
			args: args{
				slice:    []string{"*"},
				expected: "foo",
			},
			want: true,
		},
		{
			name: "kubernetes annotation",
			args: args{
				slice:    []string{"kubernetes.io/*"},
				expected: "kubernetes.io/ingress.class",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, inSliceIgnoreCaseGlob(tt.args.slice, tt.args.expected))
		})
	}
}

func Test_checkStringInMap(t *testing.T) {
	type args struct {
		local map[string]interface{}
		key   string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "key found in map",
			args: args{
				local: map[string]interface{}{
					"foo": "bar",
				},
				key: "foo",
			},
			want: "bar",
		},
		{
			name: "key not found in map",
			args: args{
				local: map[string]interface{}{
					"foo": "bar",
				},
				key: "baz",
			},
			want: "",
		},
		{
			name: "key with non string value",
			args: args{
				local: map[string]interface{}{
					"foo": map[string]string{
						"bar": "baz",
					},
				},
				key: "foo",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, kubernetes.CheckStringInMap(tt.args.local, tt.args.key))
		})
	}
}

func Test_checkKubernetesBasics(t *testing.T) {
	type args struct {
		manifest map[string]interface{}
	}

	tests := []struct {
		name string
		args args
		want kubernetes.ObjectMeta
	}{
		{
			name: "all fields found",
			args: args{
				manifest: map[string]interface{}{
					"kind":       "Deployment",
					"apiVersion": "apps/v1",
					"metadata": map[string]interface{}{
						"name": "foo",
					},
				},
			},
			want: kubernetes.ObjectMeta{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
				Name:       "foo",
			},
		},
		{
			name: "no fields found",
			args: args{
				manifest: map[string]interface{}{},
			},
			want: kubernetes.ObjectMeta{},
		},
		{
			name: "missing metadata fields",
			args: args{
				manifest: map[string]interface{}{
					"kind":       "Deployment",
					"apiVersion": "apps/v1",
				},
			},
			want: kubernetes.ObjectMeta{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, *kubernetes.ExtractMetadata(tt.args.manifest))
		})
	}
}

func TestSplit_parseYAMLManifest(t *testing.T) {
	tests := []struct {
		name       string
		contents   []byte
		strictKube bool
		want       *kubernetes.YAMLFile
		wantErr    bool
	}{
		{
			name: "valid yaml with namespace",
			contents: []byte(`---
apiVersion: v1
kind: Service
metadata:
  name: foo
  namespace: bar
`),
			want: &kubernetes.YAMLFile{
				Filename: "service-foo.yaml",
				Meta: &kubernetes.ObjectMeta{
					APIVersion: "v1",
					Kind:       "Service",
					Name:       "foo",
					Namespace:  "bar",
				},
			},
		},
		{
			name: "valid yaml without namespace",
			contents: []byte(`---
apiVersion: v1
kind: Service
metadata:
  name: foo
`),
			want: &kubernetes.YAMLFile{
				Filename: "service-foo.yaml",
				Meta: &kubernetes.ObjectMeta{
					APIVersion: "v1",
					Kind:       "Service",
					Name:       "foo",
					Namespace:  "",
				},
			},
		},
		{
			name: "valid yaml with namespace, strict kubernetes",
			contents: []byte(`---
apiVersion: v1
kind: Service
metadata:
  name: foo
  namespace: bar
`),
			strictKube: true,
			want: &kubernetes.YAMLFile{
				Filename: "service-foo.yaml",
				Meta: &kubernetes.ObjectMeta{
					APIVersion: "v1",
					Kind:       "Service",
					Name:       "foo",
					Namespace:  "bar",
				},
			},
		},
		{
			name: "valid yaml without namespace, strict kubernetes",
			contents: []byte(`---
apiVersion: v1
kind: Service
metadata:
  name: foo
`),
			strictKube: true,
			want: &kubernetes.YAMLFile{
				Filename: "service-foo.yaml",
				Meta: &kubernetes.ObjectMeta{
					APIVersion: "v1",
					Kind:       "Service",
					Name:       "foo",
					Namespace:  "",
				},
			},
		},
		{
			name: "apiVersion only, strict kubernetes",
			contents: []byte(`---
apiVersion: v1
`),
			strictKube: true,
			wantErr:    true,
		},
		{
			name: "apiVersion + kind, strict kubernetes",
			contents: []byte(`---
apiVersion: v1
kind: Foo
`),
			strictKube: true,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &Split{
				log: logger.NOOPLogger,
				template: func() *template.Renderer {
					tmpl, err := template.New(template.DefaultTemplateName)
					if err != nil {
						t.Fatalf("unable to create template: %s", err)
					}

					return tmpl
				}(),
			}
			s.opts.StrictKubernetes = tt.strictKube

			got, err := s.parseYAMLManifest(tt.contents)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSplit_parseYamlManifestAllowingEmpties(t *testing.T) {
	tests := []struct {
		name          string
		contents      []byte
		skipEmptyName bool
		skipEmptyKind bool
		includeKind   string
		includeName   string
		want          *kubernetes.YAMLFile
		wantErr       bool
	}{
		{
			name: "include name and kind",
			contents: []byte(`---
apiVersion: v1
kind: Foo
metadata:
  name: bar
`),
			want: &kubernetes.YAMLFile{
				Filename: "foo-bar.yaml",
				Meta:     &kubernetes.ObjectMeta{APIVersion: "v1", Kind: "Foo", Name: "bar"},
			},
			includeKind:   "Foo",
			skipEmptyName: false,
			skipEmptyKind: false,
		},
		{
			name: "allow empty kind",
			contents: []byte(`---
apiVersion: v1
kind: ""
metadata:
  name: bar
`),
			want: &kubernetes.YAMLFile{
				Filename: "-bar.yaml",
				Meta:     &kubernetes.ObjectMeta{APIVersion: "v1", Kind: "", Name: "bar"},
			},
			includeName:   "bar",
			skipEmptyName: false,
			skipEmptyKind: true,
		},
		{
			name: "dont allow empty kind",
			contents: []byte(`---
apiVersion: v1
metadata:
  name: bar
`),
			wantErr:       true,
			includeName:   "bar",
			skipEmptyName: false,
			skipEmptyKind: false,
		},
		{
			name: "allow empty name",
			contents: []byte(`---
apiVersion: v1
kind: Foo
metadata:
  name: ""
`),
			want: &kubernetes.YAMLFile{
				Filename: "foo-.yaml",
				Meta:     &kubernetes.ObjectMeta{APIVersion: "v1", Kind: "Foo", Name: ""},
			},
			includeKind:   "Foo",
			skipEmptyName: true,
			skipEmptyKind: false,
		},
		{
			name: "dont allow empty name",
			contents: []byte(`---
apiVersion: v1
kind: Foo
`),
			wantErr:       true,
			includeKind:   "Foo",
			skipEmptyName: false,
			skipEmptyKind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &Split{
				log: logger.NOOPLogger,
				template: func() *template.Renderer {
					tmpl, _ := template.New(template.DefaultTemplateName)
					return tmpl
				}(),
			}

			if len(tt.includeKind) > 0 {
				s.opts.IncludedKinds = []string{tt.includeKind}
			}

			if len(tt.includeName) > 0 {
				s.opts.IncludedNames = []string{tt.includeName}
			}

			s.opts.AllowEmptyKinds = tt.skipEmptyKind
			s.opts.AllowEmptyNames = tt.skipEmptyName

			if err := s.validateFilters(); err != nil {
				t.Fatalf("not expecting error validating filters, got: %s", err)
			}

			got, err := s.parseYAMLManifest(tt.contents)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
