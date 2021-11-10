package slice

import "fmt"

type strictModeSkipErr struct {
	fieldName string
}

func (s *strictModeSkipErr) Error() string {
	return fmt.Sprintf(
		"resource does not have a Kubernetes %q field or the field is invalid or empty", s.fieldName,
	)
}

type skipErr struct {
	name string
	kind string
}

func (e *skipErr) Error() string {
	return fmt.Sprintf("resource %s %q is configured to be skipped", e.kind, e.name)
}
