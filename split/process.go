package split

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type ErrKindSkip struct {
	Kind string
}

func (e *ErrKindSkip) Error() string {
	return fmt.Sprintf("file skipped by kind %q", e.Kind)
}

// processSingleYAML parses a single YAML file as received by contents. It also renders the
// template needed to generate its name
func (s *Split) processSingleYAML(contents []byte, position int, template *template.Template) (string, error) {
	// All resources we'll handle are Kubernetes manifest, and even those who are lists,
	// they're still Kubernetes Objects of type List, so we can use a map
	manifest := make(map[string]interface{})

	s.log.Println("Parsing YAML from buffer up to this point")
	if err := yaml.Unmarshal(contents, &manifest); err != nil {
		return "", fmt.Errorf("unable to parse YAML file number %d: %w", position, err)
	}

	// We need to check if the file is skipped by kind
	if hasIncluded, hasExcluded := len(s.opts.IncludedKinds) > 0, len(s.opts.ExcludedKinds) > 0; hasIncluded || hasExcluded {
		// Retrieve the kind from the YAML code
		kind, err := getKindFromYAML(manifest, position)
		if err != nil {
			return "", err
		}

		// If we're working with including only, then filter by it
		if hasIncluded && !inSliceIgnoreCase(s.opts.IncludedKinds, kind) {
			return "", &ErrKindSkip{Kind: kind}
		}

		// Otherwise exclude based on the parameter received
		if hasExcluded && inSliceIgnoreCase(s.opts.ExcludedKinds, kind) {
			return "", &ErrKindSkip{Kind: kind}
		}
	}

	// Render the name to a buffer using the Go Template
	s.log.Println("Rendering filename template from Go Template")
	var buf bytes.Buffer
	if err := template.Execute(&buf, manifest); err != nil {
		return "", fmt.Errorf("unable to render file name for YAML file number %d: %w", position, improveExecError(err))
	}

	// Trim the file name
	name := strings.TrimSpace(buf.String())

	// Fix for text/template Go issue #24963, as well as removing any linebreaks
	name = strings.NewReplacer("<no value>", "", "\n", "").Replace(name)

	if s := strings.TrimSuffix(name, filepath.Ext(name)); s == "" {
		return "", fmt.Errorf("file name rendered will yield no file name")
	}

	return name, nil
}

// getKindFromYAML returns the kind of a Kubernetes resource from a parsed YAML file
func getKindFromYAML(manifest map[string]interface{}, fileNumber int) (string, error) {
	// Find the kind of the current file
	k, found := manifest["kind"]
	if !found {
		return "", fmt.Errorf("unable to find Kubernetes resource kind in file number %d", fileNumber)
	}

	// Check if the kind is a string or another arbitrary object
	kind, ok := k.(string)
	if !ok {
		return "", fmt.Errorf("a Kubernetes resource kind was provided in file number %d, but not as a string", fileNumber)
	}

	return kind, nil
}

func inSliceIgnoreCase(s []string, e string) bool {
	e = strings.ToLower(e)

	for _, a := range s {
		if strings.ToLower(a) == e {
			return true
		}
	}

	return false
}
