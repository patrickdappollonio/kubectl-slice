package kubernetes

import (
	"strings"

	"github.com/patrickdappollonio/kubectl-slice/pkg/errors"
)

// CheckGroupInclusion validates if a resource belongs to any of the specified groups
// Returns nil if the resource should be included, or an error if it should be skipped
func CheckGroupInclusion(objmeta *ObjectMeta, groupNames []string, included bool) error {
	for _, group := range groupNames {
		if included {
			if objmeta.GetGroupFromAPIVersion() == strings.ToLower(group) {
				return nil
			}
		} else {
			if objmeta.GetGroupFromAPIVersion() == strings.ToLower(group) {
				return &errors.SkipErr{}
			}
		}
	}

	if included {
		return &errors.SkipErr{}
	}

	return nil
}

// ValidateRequiredFields verifies if a resource has all required Kubernetes fields
// when operating in strict mode
func ValidateRequiredFields(meta *ObjectMeta, strictMode bool) error {
	if !strictMode {
		return nil
	}

	if meta.APIVersion == "" {
		return &errors.StrictModeSkipErr{FieldName: "apiVersion"}
	}

	if meta.Kind == "" {
		return &errors.StrictModeSkipErr{FieldName: "kind"}
	}

	if meta.Name == "" {
		return &errors.StrictModeSkipErr{FieldName: "metadata.name"}
	}

	return nil
}
