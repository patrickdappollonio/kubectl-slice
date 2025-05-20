package kubernetes

import (
	"testing"

	"github.com/patrickdappollonio/kubectl-slice/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestCheckGroupInclusion(t *testing.T) {
	tests := []struct {
		name             string
		objmeta          *ObjectMeta
		groupNames       []string
		included         bool
		expectSkipErr    bool
		expectStrictErr  bool
		strictErrField   string
	}{
		{
			name: "included mode - group matches",
			objmeta: &ObjectMeta{
				APIVersion: "apps/v1",
			},
			groupNames: []string{"apps"},
			included:   true,
			expectSkipErr: false,
		},
		{
			name: "included mode - group doesn't match",
			objmeta: &ObjectMeta{
				APIVersion: "apps/v1",
			},
			groupNames: []string{"networking.k8s.io"},
			included:   true,
			expectSkipErr: true,
		},
		{
			name: "included mode - core group",
			objmeta: &ObjectMeta{
				APIVersion: "v1",
			},
			groupNames: []string{""},
			included:   true,
			expectSkipErr: false,
		},
		{
			name: "excluded mode - group matches",
			objmeta: &ObjectMeta{
				APIVersion: "apps/v1",
			},
			groupNames: []string{"apps"},
			included:   false,
			expectSkipErr: true,
		},
		{
			name: "excluded mode - group doesn't match",
			objmeta: &ObjectMeta{
				APIVersion: "apps/v1",
			},
			groupNames: []string{"networking.k8s.io"},
			included:   false,
			expectSkipErr: false,
		},
		{
			name: "multiple groups - included mode - one matches",
			objmeta: &ObjectMeta{
				APIVersion: "apps/v1",
			},
			groupNames: []string{"networking.k8s.io", "apps", "batch"},
			included:   true,
			expectSkipErr: false,
		},
		{
			name: "multiple groups - excluded mode - one matches",
			objmeta: &ObjectMeta{
				APIVersion: "apps/v1",
			},
			groupNames: []string{"networking.k8s.io", "apps", "batch"},
			included:   false,
			expectSkipErr: true,
		},
		{
			name: "case insensitive comparison",
			objmeta: &ObjectMeta{
				APIVersion: "Apps/v1",
			},
			groupNames: []string{"apps"},
			included:   true,
			expectSkipErr: false,
		},
		{
			name: "empty group list - included mode",
			objmeta: &ObjectMeta{
				APIVersion: "apps/v1",
			},
			groupNames: []string{},
			included:   true,
			expectSkipErr: true,
		},
		{
			name: "empty group list - excluded mode",
			objmeta: &ObjectMeta{
				APIVersion: "apps/v1",
			},
			groupNames: []string{},
			included:   false,
			expectSkipErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckGroupInclusion(tt.objmeta, tt.groupNames, tt.included)
			
			if tt.expectSkipErr {
				require.IsType(t, &errors.SkipErr{}, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestValidateRequiredFields(t *testing.T) {
	tests := []struct {
		name             string
		meta             *ObjectMeta
		strictMode       bool
		expectErr        bool
		expectedField    string
	}{
		{
			name: "strict mode - all fields present",
			meta: &ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test-deployment",
			},
			strictMode: true,
			expectErr: false,
		},
		{
			name: "non-strict mode - all fields present",
			meta: &ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test-deployment",
			},
			strictMode: false,
			expectErr: false,
		},
		{
			name: "non-strict mode - missing fields",
			meta: &ObjectMeta{},
			strictMode: false,
			expectErr: false,
		},
		{
			name: "strict mode - missing apiVersion",
			meta: &ObjectMeta{
				Kind: "Deployment",
				Name: "test-deployment",
			},
			strictMode: true,
			expectErr: true,
			expectedField: "apiVersion",
		},
		{
			name: "strict mode - missing kind",
			meta: &ObjectMeta{
				APIVersion: "apps/v1",
				Name:       "test-deployment",
			},
			strictMode: true,
			expectErr: true,
			expectedField: "kind",
		},
		{
			name: "strict mode - missing name",
			meta: &ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
			strictMode: true,
			expectErr: true,
			expectedField: "metadata.name",
		},
		{
			name: "strict mode - namespace optional",
			meta: &ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test-deployment",
			},
			strictMode: true,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequiredFields(tt.meta, tt.strictMode)
			
			if tt.expectErr {
				require.NotNil(t, err)
				strictErr, ok := err.(*errors.StrictModeSkipErr)
				require.True(t, ok)
				require.Equal(t, tt.expectedField, strictErr.FieldName)
			} else {
				require.Nil(t, err)
			}
		})
	}
}
