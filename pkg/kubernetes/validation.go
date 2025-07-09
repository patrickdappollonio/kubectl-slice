package kubernetes

import (
	"fmt"
	"strings"

	"github.com/patrickdappollonio/kubectl-slice/pkg/errors"
)

// CheckGroupInclusion validates if a resource belongs to any of the specified groups
// Returns nil if the resource should be included, or an error if it should be skipped
func CheckGroupInclusion(objmeta *ObjectMeta, groupNames []string, included bool) error {
	resourceGroup := objmeta.GetGroupFromAPIVersion()
	
	for _, group := range groupNames {
		if included {
			if resourceGroup == strings.ToLower(group) {
				return nil
			}
		} else {
			if resourceGroup == strings.ToLower(group) {
				return &errors.SkipErr{
					Name:   objmeta.Name,
					Kind:   objmeta.Kind,
					Group:  resourceGroup,
					Reason: fmt.Sprintf("matches excluded group %q", group),
				}
			}
		}
	}

	if included {
		var reason string
		if len(groupNames) > 0 {
			reason = fmt.Sprintf("does not match any included groups %v", groupNames)
		} else {
			reason = "no included groups specified"
		}
		
		return &errors.SkipErr{
			Name:   objmeta.Name,
			Kind:   objmeta.Kind,
			Group:  resourceGroup,
			Reason: reason,
		}
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
