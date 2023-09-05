package slice

import (
	"fmt"
	"strings"
)

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

const nonK8sHelper = `the file has no Kubernetes metadata: it is most likely a non-Kubernetes YAML file, you can skip it with --skip-non-k8s`

type cantFindFieldErr struct {
	fieldName string
	fileCount int
	meta      kubeObjectMeta
}

func (e *cantFindFieldErr) Error() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(
		"unable to find Kubernetes %q field in file %d",
		e.fieldName, e.fileCount,
	))

	if e.meta.empty() {
		sb.WriteString(": " + nonK8sHelper)
	} else {
		sb.WriteString(fmt.Sprintf(
			": processed details: kind %q, name %q, apiVersion %q",
			e.meta.Kind, e.meta.Name, e.meta.APIVersion,
		))
	}

	return sb.String()
}
