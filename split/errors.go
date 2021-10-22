package split

import "fmt"

type strictModeErr struct {
	fileNumber int
	fieldName  string
}

func (s *strictModeErr) Error() string {
	return fmt.Sprintf(
		"YAML file number %d does not contain a Kubernetes %q field or the field is invalid or empty",
		s.fileNumber, s.fieldName,
	)
}
