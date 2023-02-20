package slice

import (
	"testing"
	"text/template"

	local "github.com/patrickdappollonio/kubectl-slice/slice/template"
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
			if got := inSliceIgnoreCase(tt.args.slice, tt.args.expected); got != tt.want {
				t.Errorf("inSliceIgnoreCase() = %v, want %v", got, tt.want)
			}
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
			if got := inSliceIgnoreCaseGlob(tt.args.slice, tt.args.expected); got != tt.want {
				t.Errorf("inSliceIgnoreCase() = %v, want %v", got, tt.want)
			}
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
			if str := checkStringInMap(tt.args.local, tt.args.key); str != tt.want {
				t.Errorf("checkStringInMap() = %v, want %v", str, tt.want)
			}
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
		want kubeObjectMeta
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
			want: kubeObjectMeta{
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
			want: kubeObjectMeta{},
		},
		{
			name: "missing metadata fields",
			args: args{
				manifest: map[string]interface{}{
					"kind":       "Deployment",
					"apiVersion": "apps/v1",
				},
			},
			want: kubeObjectMeta{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := checkKubernetesBasics(tt.args.manifest)

			if meta.Kind != tt.want.Kind {
				t.Errorf("checkKubernetesBasics() Kind = %v, want %v", meta.Kind, tt.want.Kind)
			}

			if meta.APIVersion != tt.want.APIVersion {
				t.Errorf("checkKubernetesBasics() APIVersion = %v, want %v", meta.APIVersion, tt.want.APIVersion)
			}

			if meta.Name != tt.want.Name {
				t.Errorf("checkKubernetesBasics() Name = %v, want %v", meta.Name, tt.want.Name)
			}
		})
	}
}

func TestSplit_parseYAMLManifest(t *testing.T) {
	tests := []struct {
		name       string
		contents   []byte
		strictKube bool
		want       yamlFile
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
			want: yamlFile{
				filename: "service-foo.yaml",
				meta: kubeObjectMeta{
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
			want: yamlFile{
				filename: "service-foo.yaml",
				meta: kubeObjectMeta{
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
			want: yamlFile{
				filename: "service-foo.yaml",
				meta: kubeObjectMeta{
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
			want: yamlFile{
				filename: "service-foo.yaml",
				meta: kubeObjectMeta{
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

			s := &Split{
				log:      nolog,
				template: template.Must(template.New(DefaultTemplateName).Funcs(local.Functions).Parse(DefaultTemplateName)),
			}
			s.opts.StrictKubernetes = tt.strictKube

			got, err := s.parseYAMLManifest(tt.contents)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}
