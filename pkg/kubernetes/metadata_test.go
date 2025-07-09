package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetGroupFromAPIVersion(t *testing.T) {
	tests := []struct {
		name       string
		apiVersion string
		want       string
	}{
		{
			name:       "with group",
			apiVersion: "apps/v1",
			want:       "apps",
		},
		{
			name:       "core group",
			apiVersion: "v1",
			want:       "",
		},
		{
			name:       "empty string",
			apiVersion: "",
			want:       "",
		},
		{
			name:       "multiple slashes",
			apiVersion: "example.com/group/v1",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &ObjectMeta{
				APIVersion: tt.apiVersion,
			}
			require.Equal(t, tt.want, k.GetGroupFromAPIVersion())
		})
	}
}

func TestEmpty(t *testing.T) {
	tests := []struct {
		name     string
		metadata ObjectMeta
		want     bool
	}{
		{
			name:     "empty metadata",
			metadata: ObjectMeta{},
			want:     true,
		},
		{
			name: "only APIVersion",
			metadata: ObjectMeta{
				APIVersion: "v1",
			},
			want: false,
		},
		{
			name: "only Kind",
			metadata: ObjectMeta{
				Kind: "Pod",
			},
			want: false,
		},
		{
			name: "only Name",
			metadata: ObjectMeta{
				Name: "test-pod",
			},
			want: false,
		},
		{
			name: "only Namespace",
			metadata: ObjectMeta{
				Namespace: "default",
			},
			want: false,
		},
		{
			name: "fully populated",
			metadata: ObjectMeta{
				APIVersion: "v1",
				Kind:       "Pod",
				Name:       "test-pod",
				Namespace:  "default",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.metadata.Empty())
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		metadata ObjectMeta
		want     string
	}{
		{
			name:     "empty metadata",
			metadata: ObjectMeta{},
			want:     "kind , name , apiVersion",
		},
		{
			name: "fully populated",
			metadata: ObjectMeta{
				APIVersion: "v1",
				Kind:       "Pod",
				Name:       "test-pod",
				Namespace:  "default",
			},
			want: "kind Pod, name test-pod, apiVersion v1",
		},
		{
			name: "partial data",
			metadata: ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
			want: "kind Deployment, name , apiVersion apps/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.metadata.String())
		})
	}
}

func TestCheckStringInMap(t *testing.T) {
	tests := []struct {
		name  string
		local map[string]interface{}
		key   string
		want  string
	}{
		{
			name:  "empty map",
			local: map[string]interface{}{},
			key:   "test",
			want:  "",
		},
		{
			name: "key exists with string value",
			local: map[string]interface{}{
				"test": "value",
			},
			key:  "test",
			want: "value",
		},
		{
			name: "key exists with non-string value",
			local: map[string]interface{}{
				"test": 123,
			},
			key:  "test",
			want: "",
		},
		{
			name: "key doesn't exist",
			local: map[string]interface{}{
				"other": "value",
			},
			key:  "test",
			want: "",
		},
		{
			name: "multiple keys",
			local: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			key:  "key2",
			want: "value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, CheckStringInMap(tt.local, tt.key))
		})
	}
}

func TestExtractMetadata(t *testing.T) {
	tests := []struct {
		name     string
		manifest map[string]interface{}
		want     ObjectMeta
	}{
		{
			name:     "empty manifest",
			manifest: map[string]interface{}{},
			want:     ObjectMeta{},
		},
		{
			name: "minimal manifest",
			manifest: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
			},
			want: ObjectMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
		},
		{
			name: "full manifest",
			manifest: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      "test-deployment",
					"namespace": "default",
				},
			},
			want: ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test-deployment",
				Namespace:  "default",
			},
		},
		{
			name: "metadata is not a map",
			manifest: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata":   "invalid",
			},
			want: ObjectMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
		},
		{
			name: "metadata with non-string values",
			manifest: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name":      123,
					"namespace": true,
				},
			},
			want: ObjectMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractMetadata(tt.manifest)

			require.Equal(t, tt.want.APIVersion, got.APIVersion)
			require.Equal(t, tt.want.Kind, got.Kind)
			require.Equal(t, tt.want.Name, got.Name)
			require.Equal(t, tt.want.Namespace, got.Namespace)
		})
	}
}
