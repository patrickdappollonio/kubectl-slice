package errors

import (
	"fmt"
	"strings"
)

// StrictModeSkipErr represents an error when a Kubernetes resource is skipped
// in strict mode because a required field is missing or empty
type StrictModeSkipErr struct {
	FieldName string
}

func (s *StrictModeSkipErr) Error() string {
	return fmt.Sprintf(
		"resource does not have a Kubernetes %q field or the field is invalid or empty", s.FieldName,
	)
}

// SkipErr represents an error when a Kubernetes resource is intentionally skipped
// based on user-provided include/exclude filter configuration
type SkipErr struct {
	Name   string
	Kind   string
	Group  string
	Reason string
}

func (e *SkipErr) Error() string {
	if e.Name == "" && e.Kind == "" {
		if e.Group != "" {
			if e.Reason != "" {
				return fmt.Sprintf("resource with API group %q is skipped: %s", e.Group, e.Reason)
			}
			return fmt.Sprintf("resource with API group %q is configured to be skipped", e.Group)
		}
		return "resource is configured to be skipped"
	}

	if e.Reason != "" {
		return fmt.Sprintf("resource %s %q is skipped: %s", e.Kind, e.Name, e.Reason)
	}
	return fmt.Sprintf("resource %s %q is configured to be skipped", e.Kind, e.Name)
}

// nonKubernetesMessage provides a standard error message for YAML files that don't contain
// standard Kubernetes metadata and are likely not Kubernetes resources
const nonKubernetesMessage = `the file has no Kubernetes metadata: it is most likely a non-Kubernetes YAML file, you can skip it with --skip-non-k8s`

// CantFindFieldErr represents an error when a required field is missing in a Kubernetes
// resource. It includes contextual information about the file and resource.
type CantFindFieldErr struct {
	FieldName string
	FileCount int
	Meta      interface{}
}

func (e *CantFindFieldErr) Error() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(
		"unable to find Kubernetes %q field in file %d",
		e.FieldName, e.FileCount,
	))

	// Type assertion to check if Meta has an empty() method
	if metaWithEmpty, ok := e.Meta.(interface{ empty() bool }); ok && metaWithEmpty.empty() {
		sb.WriteString(": " + nonKubernetesMessage)
	} else if meta, ok := e.Meta.(fmt.Stringer); ok {
		sb.WriteString(fmt.Sprintf(": %s", meta.String()))
	}

	return sb.String()
}
