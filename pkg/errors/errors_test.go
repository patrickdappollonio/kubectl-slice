package errors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStrictModeSkipErr_Error(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		want      string
	}{
		{
			name:      "with metadata.name field",
			fieldName: "metadata.name",
			want:      "resource does not have a Kubernetes \"metadata.name\" field or the field is invalid or empty",
		},
		{
			name:      "with kind field",
			fieldName: "kind",
			want:      "resource does not have a Kubernetes \"kind\" field or the field is invalid or empty",
		},
		{
			name:      "with empty field",
			fieldName: "",
			want:      "resource does not have a Kubernetes \"\" field or the field is invalid or empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StrictModeSkipErr{
				FieldName: tt.fieldName,
			}

			require.Equal(t, tt.want, s.Error())
		})
	}
}

func TestSkipErr_Error(t *testing.T) {
	tests := []struct {
		name string
		err  SkipErr
		want string
	}{
		{
			name: "with name and kind",
			err: SkipErr{
				Name: "my-pod",
				Kind: "Pod",
			},
			want: "resource Pod \"my-pod\" is configured to be skipped",
		},
		{
			name: "with name, kind and reason",
			err: SkipErr{
				Name:   "my-pod",
				Kind:   "Pod",
				Reason: "matched exclusion filter",
			},
			want: "resource Pod \"my-pod\" is skipped: matched exclusion filter",
		},
		{
			name: "with group only",
			err: SkipErr{
				Group: "apps/v1",
			},
			want: "resource with API group \"apps/v1\" is configured to be skipped",
		},
		{
			name: "with group and reason",
			err: SkipErr{
				Group:  "apps/v1",
				Reason: "matched exclusion filter",
			},
			want: "resource with API group \"apps/v1\" is skipped: matched exclusion filter",
		},
		{
			name: "empty fields",
			err:  SkipErr{},
			want: "resource is configured to be skipped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.err.Error())
		})
	}
}

// mockMeta implements the empty() method for testing CantFindFieldErr
type mockMeta struct {
	isEmpty bool
	str     string
}

func (m mockMeta) empty() bool {
	return m.isEmpty
}

func (m mockMeta) String() string {
	return m.str
}

// mockMetaStringOnly implements just the String() method without empty()
type mockMetaStringOnly struct {
	str string
}

func (m mockMetaStringOnly) String() string {
	return m.str
}

func TestErrorsInterface(t *testing.T) {
	require.Implementsf(t, (*error)(nil), &StrictModeSkipErr{}, "StrictModeSkipErr should implement error")
	require.Implementsf(t, (*error)(nil), &SkipErr{}, "SkipErr should implement error")
	require.Implementsf(t, (*error)(nil), &CantFindFieldErr{}, "CantFindFieldErr should implement error")
}

func TestCantFindFieldErr_Error(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		fileCount int
		meta      interface{}
		want      string
	}{
		{
			name:      "with empty meta",
			fieldName: "metadata.name",
			fileCount: 1,
			meta:      mockMeta{isEmpty: true},
			want:      "unable to find Kubernetes \"metadata.name\" field in file 1: " + nonKubernetesMessage,
		},
		{
			name:      "with non-empty meta with stringer",
			fieldName: "metadata.name",
			fileCount: 2,
			meta:      mockMeta{isEmpty: false, str: "Pod/my-pod"},
			want:      "unable to find Kubernetes \"metadata.name\" field in file 2: Pod/my-pod",
		},
		{
			name:      "with meta implementing only String",
			fieldName: "kind",
			fileCount: 3,
			meta:      mockMetaStringOnly{str: "Kind/Deployment"},
			want:      "unable to find Kubernetes \"kind\" field in file 3: Kind/Deployment",
		},
		{
			name:      "with nil meta",
			fieldName: "kind",
			fileCount: 4,
			meta:      nil,
			want:      "unable to find Kubernetes \"kind\" field in file 4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &CantFindFieldErr{
				FieldName: tt.fieldName,
				FileCount: tt.fileCount,
				Meta:      tt.meta,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("CantFindFieldErr.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}
