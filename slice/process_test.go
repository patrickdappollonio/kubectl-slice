package slice

import (
	"testing"
)

func Test_getKindFromYAML(t *testing.T) {
	type args struct {
		manifest   map[string]interface{}
		fileNumber int
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "successful parse",
			args: args{
				manifest: map[string]interface{}{
					"kind": "Deployment",
				},
				fileNumber: 1,
			},
			want: "Deployment",
		},
		{
			name: "unsuccessful parse",
			args: args{
				manifest: map[string]interface{}{
					"foo": "bar",
				},
				fileNumber: 1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getKindFromYAML(tt.args.manifest, tt.args.fileNumber)

			if (err != nil) != tt.wantErr {
				t.Errorf("getKindFromYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("getKindFromYAML() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inSliceIgnoreCase(tt.args.slice, tt.args.expected); got != tt.want {
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
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "key found in map",
			args: args{
				local: map[string]interface{}{
					"foo": "bar",
				},
				key: "foo",
			},
		},
		{
			name: "key not found in map",
			args: args{
				local: map[string]interface{}{
					"foo": "bar",
				},
				key: "baz",
			},
			wantErr: true,
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
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkStringInMap(tt.args.local, tt.args.key, ""); (err != nil) != tt.wantErr {
				t.Errorf("checkStringInMap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_checkKubernetesBasics(t *testing.T) {
	type args struct {
		manifest map[string]interface{}
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
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
		},
		{
			name: "no fields found",
			args: args{
				manifest: map[string]interface{}{},
			},
			wantErr: true,
		},
		{
			name: "missing metadata fields",
			args: args{
				manifest: map[string]interface{}{
					"kind":       "Deployment",
					"apiVersion": "apps/v1",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkKubernetesBasics(tt.args.manifest); (err != nil) != tt.wantErr {
				t.Errorf("checkKubernetesBasics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
