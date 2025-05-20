package errors

import (
	"fmt"
	"strings"
)

// StrictModeSkipErr indicates a resource is being skipped in strict mode due to missing field
type StrictModeSkipErr struct {
	FieldName string
}

func (s *StrictModeSkipErr) Error() string {
	return fmt.Sprintf(
		"resource does not have a Kubernetes %q field or the field is invalid or empty", s.FieldName,
	)
}

// SkipErr indicates a resource is being skipped based on user configuration
type SkipErr struct {
	Name string
	Kind string
}

func (e *SkipErr) Error() string {
	return fmt.Sprintf("resource %s %q is configured to be skipped", e.Kind, e.Name)
}

// NonK8sHelper is a constant message for non-Kubernetes YAML files
const NonK8sHelper = `the file has no Kubernetes metadata: it is most likely a non-Kubernetes YAML file, you can skip it with --skip-non-k8s`

// CantFindFieldErr indicates a field is missing in a Kubernetes resource
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
		sb.WriteString(": " + NonK8sHelper)
	} else if meta, ok := e.Meta.(fmt.Stringer); ok {
		sb.WriteString(fmt.Sprintf(": %s", meta.String()))
	}

	return sb.String()
}
