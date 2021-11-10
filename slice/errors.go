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

type kindSkipErr struct {
	Kind string
}

func (e *kindSkipErr) Error() string {
	return fmt.Sprintf("resource kind %q is configured to be skipped", e.Kind)
}
