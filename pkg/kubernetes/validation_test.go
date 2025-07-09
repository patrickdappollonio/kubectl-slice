package kubernetes

import (
	"errors"
	"testing"

	apperrors "github.com/patrickdappollonio/kubectl-slice/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestCheckGroupInclusion(t *testing.T) {
	tests := []struct {
		name           string
		objmeta        *ObjectMeta
		groupNames     []string
		included       bool
		expectSkipErr  bool
		expectedGroup  string
		expectedReason string
	}{
		{
			name: "included mode - group matches",
			objmeta: &ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test-app",
			},
			groupNames:    []string{"apps"},
			included:      true,
			expectSkipErr: false,
		},
		{
			name: "included mode - group doesn't match",
			objmeta: &ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test-app",
			},
			groupNames:     []string{"networking.k8s.io"},
			included:       true,
			expectSkipErr:  true,
			expectedGroup:  "apps",
			expectedReason: "does not match any included groups [networking.k8s.io]",
		},
		{
			name: "included mode - core group",
			objmeta: &ObjectMeta{
				APIVersion: "v1",
				Kind:       "Pod",
				Name:       "test-pod",
			},
			groupNames:    []string{""},
			included:      true,
			expectSkipErr: false,
		},
		{
			name: "excluded mode - group matches",
			objmeta: &ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test-app",
			},
			groupNames:     []string{"apps"},
			included:       false,
			expectSkipErr:  true,
			expectedGroup:  "apps",
			expectedReason: "matches excluded group \"apps\"",
		},
		{
			name: "excluded mode - group doesn't match",
			objmeta: &ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test-app",
			},
			groupNames:    []string{"networking.k8s.io"},
			included:      false,
			expectSkipErr: false,
		},
		{
			name: "multiple groups - included mode - one matches",
			objmeta: &ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test-app",
			},
			groupNames:    []string{"networking.k8s.io", "apps", "batch"},
			included:      true,
			expectSkipErr: false,
		},
		{
			name: "multiple groups - excluded mode - one matches",
			objmeta: &ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test-app",
			},
			groupNames:     []string{"networking.k8s.io", "apps", "batch"},
			included:       false,
			expectSkipErr:  true,
			expectedGroup:  "apps",
			expectedReason: "matches excluded group \"apps\"",
		},
		{
			name: "case insensitive comparison",
			objmeta: &ObjectMeta{
				APIVersion: "Apps/v1",
				Kind:       "Deployment",
				Name:       "test-app",
			},
			groupNames:    []string{"apps"},
			included:      true,
			expectSkipErr: false,
		},
		{
			name: "empty group list - included mode",
			objmeta: &ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test-app",
			},
			groupNames:     []string{},
			included:       true,
			expectSkipErr:  true,
			expectedGroup:  "apps",
			expectedReason: "no included groups specified",
		},
		{
			name: "empty group list - excluded mode",
			objmeta: &ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test-app",
			},
			groupNames:    []string{},
			included:      false,
			expectSkipErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckGroupInclusion(tt.objmeta, tt.groupNames, tt.included)

			if tt.expectSkipErr {
				require.Error(t, err)

				// Use errors.As for type checking instead of type assertion
				var skipErr *apperrors.SkipErr
				require.True(t, errors.As(err, &skipErr), "Expected error of type *errors.SkipErr")

				// Additional checks for the error details when we have expected values
				if tt.expectedGroup != "" {
					require.Equal(t, tt.expectedGroup, skipErr.Group)

					if tt.expectedReason != "" {
						require.Equal(t, tt.expectedReason, skipErr.Reason)
					}

					require.Equal(t, tt.objmeta.Kind, skipErr.Kind)
					require.Equal(t, tt.objmeta.Name, skipErr.Name)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateRequiredFields(t *testing.T) {
	tests := []struct {
		name          string
		meta          *ObjectMeta
		strictMode    bool
		expectErr     bool
		expectedField string
	}{
		{
			name: "strict mode - all fields present",
			meta: &ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test-deployment",
			},
			strictMode: true,
			expectErr:  false,
		},
		{
			name: "non-strict mode - all fields present",
			meta: &ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "test-deployment",
			},
			strictMode: false,
			expectErr:  false,
		},
		{
			name:       "non-strict mode - missing fields",
			meta:       &ObjectMeta{},
			strictMode: false,
			expectErr:  false,
		},
		{
			name: "strict mode - missing apiVersion",
			meta: &ObjectMeta{
				Kind: "Deployment",
				Name: "test-deployment",
			},
			strictMode:    true,
			expectErr:     true,
			expectedField: "apiVersion",
		},
		{
			name: "strict mode - missing kind",
			meta: &ObjectMeta{
				APIVersion: "apps/v1",
				Name:       "test-deployment",
			},
			strictMode:    true,
			expectErr:     true,
			expectedField: "kind",
		},
		{
			name: "strict mode - missing name",
			meta: &ObjectMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
			strictMode:    true,
			expectErr:     true,
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
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequiredFields(tt.meta, tt.strictMode)

			if tt.expectErr {
				require.Error(t, err)
				var strictErr *apperrors.StrictModeSkipErr
				require.True(t, errors.As(err, &strictErr), "Expected error of type *errors.StrictModeSkipErr")
				require.Equal(t, tt.expectedField, strictErr.FieldName)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
