package slice

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// parseYAMLManifest parses a single YAML file as received by contents. It also renders the
// template needed to generate its name
func (s *Split) parseYAMLManifest(contents []byte, position int, template *template.Template) (string, string, error) {
	// All resources we'll handle are Kubernetes manifest, and even those who are lists,
	// they're still Kubernetes Objects of type List, so we can use a map
	manifest := make(map[string]interface{})

	s.log.Println("Parsing YAML from buffer up to this point")
	if err := yaml.Unmarshal(contents, &manifest); err != nil {
		return "", "", fmt.Errorf("unable to parse YAML file number %d: %w", position, err)
	}

	// Render the name to a buffer using the Go Template
	s.log.Println("Rendering filename template from Go Template")
	var buf bytes.Buffer
	if err := template.Execute(&buf, manifest); err != nil {
		return "", "", fmt.Errorf("unable to render file name for YAML file number %d: %w", position, improveExecError(err))
	}

	// Check if file contains at least some Kubernetes keys
	if s.opts.StrictKubernetes {
		if err := checkKubernetesBasics(manifest); err != nil {
			return "", "", err
		}
	}

	// Check before handling if we're about to filter resources
	hasIncluded, hasExcluded := len(s.opts.IncludedKinds) > 0, len(s.opts.ExcludedKinds) > 0

	// Retrieve the kind from the YAML code
	kind, err := getKindFromYAML(manifest, position)

	// Check if we have an error, if we do, we need to ensure we're not
	// filtering too. When there's an error and we are filtering, the error
	// matters. Otherwise, the error can be safely ignored
	if err != nil && (hasIncluded || hasExcluded) {
		return "", "", err
	}

	// We need to check if the file is skipped by kind
	if hasIncluded || hasExcluded {
		// If we're working with including only, then filter by it
		if hasIncluded && !inSliceIgnoreCase(s.opts.IncludedKinds, kind) {
			return "", "", &kindSkipErr{Kind: kind}
		}

		// Otherwise exclude based on the parameter received
		if hasExcluded && inSliceIgnoreCase(s.opts.ExcludedKinds, kind) {
			return "", "", &kindSkipErr{Kind: kind}
		}
	}

	// Trim the file name
	name := strings.TrimSpace(buf.String())

	// Fix for text/template Go issue #24963, as well as removing any linebreaks
	name = strings.NewReplacer("<no value>", "", "\n", "").Replace(name)

	if s := strings.TrimSuffix(name, filepath.Ext(name)); s == "" {
		return "", "", fmt.Errorf("file name rendered will yield no file name for YAML file number %d", position)
	}

	return name, kind, nil
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

func inSliceIgnoreCase(slice []string, expected string) bool {
	expected = strings.ToLower(expected)

	for _, a := range slice {
		if strings.ToLower(a) == expected {
			return true
		}
	}

	return false
}

func checkStringInMap(local map[string]interface{}, key, keyprefix string) error {
	iface, found := local[key]

	if !found {
		return &strictModeErr{keyprefix + key}
	}

	if _, ok := iface.(string); !ok {
		return &strictModeErr{keyprefix + key}
	}

	return nil
}

func checkKubernetesBasics(manifest map[string]interface{}) error {
	if err := checkStringInMap(manifest, "apiVersion", ""); err != nil {
		return err
	}

	if err := checkStringInMap(manifest, "kind", ""); err != nil {
		return err
	}

	metadata, found := manifest["metadata"]

	if !found {
		return &strictModeErr{"metadata"}
	}

	if err := checkStringInMap(metadata.(map[string]interface{}), "name", "metadata."); err != nil {
		return err
	}

	return nil
}
