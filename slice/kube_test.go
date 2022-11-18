package slice

import (
	"testing"
)

func Test_newKubeObjectMeta(t *testing.T) {
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
			meta := newKubeObjectMeta(tt.args.manifest)

			if meta.Kind != tt.want.Kind {
				t.Errorf("newKubeObjectMeta() Kind = %v, want %v", meta.Kind, tt.want.Kind)
			}

			if meta.APIVersion != tt.want.APIVersion {
				t.Errorf("newKubeObjectMeta() APIVersion = %v, want %v", meta.APIVersion, tt.want.APIVersion)
			}

			if meta.Name != tt.want.Name {
				t.Errorf("newKubeObjectMeta() Name = %v, want %v", meta.Name, tt.want.Name)
			}
		})
	}
}

func Test_kubeObjectMeta_findMissingField(t *testing.T) {
	type fields struct {
		APIVersion string
		Kind       string
		Name       string
		Namespace  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "all fields found - missing namespace",
			fields: fields{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "foo",
			},
			want: "",
		},
		{
			name: "all fields found - including namespace",
			fields: fields{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "foo",
				Namespace:  "bar",
			},
			want: "",
		},
		{
			name: "missing apiVersion",
			fields: fields{
				Kind: "Deployment",
				Name: "foo",
			},
			want: "apiVersion",
		},
		{
			name: "missing kind",
			fields: fields{
				APIVersion: "apps/v1",
				Name:       "foo",
			},
			want: "kind",
		},
		{
			name: "missing name",
			fields: fields{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
			want: "metadata.name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &kubeObjectMeta{
				APIVersion: tt.fields.APIVersion,
				Kind:       tt.fields.Kind,
				Name:       tt.fields.Name,
				Namespace:  tt.fields.Namespace,
			}
			if got := k.findMissingField(); got != tt.want {
				t.Errorf("kubeObjectMeta.findMissingField() = %v, want %v", got, tt.want)
			}
		})
	}
}
