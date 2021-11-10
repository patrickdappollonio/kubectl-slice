package slice

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// parseYAMLManifest parses a single YAML file as received by contents. It also renders the
// template needed to generate its name
func (s *Split) parseYAMLManifest(contents []byte) (yamlFile, error) {
	// All resources we'll handle are Kubernetes manifest, and even those who are lists,
	// they're still Kubernetes Objects of type List, so we can use a map
	manifest := make(map[string]interface{})

	s.log.Println("Parsing YAML from buffer up to this point")
	if err := yaml.Unmarshal(contents, &manifest); err != nil {
		return yamlFile{}, fmt.Errorf("unable to parse YAML file number %d: %w", s.fileCount, err)
	}

	// Render the name to a buffer using the Go Template
	s.log.Println("Rendering filename template from Go Template")
	var buf bytes.Buffer
	if err := s.template.Execute(&buf, manifest); err != nil {
		return yamlFile{}, fmt.Errorf("unable to render file name for YAML file number %d: %w", s.fileCount, improveExecError(err))
	}

	// Check if file contains the required Kubernetes metadata
	k8smeta := checkKubernetesBasics(manifest)

	// Check if at least the three fields are not empty
	if s.opts.StrictKubernetes {
		if k8smeta.APIVersion == "" {
			return yamlFile{}, &strictModeSkipErr{fieldName: "apiVersion"}
		}

		if k8smeta.Kind == "" {
			return yamlFile{}, &strictModeSkipErr{fieldName: "kind"}
		}

		if k8smeta.Name == "" {
			return yamlFile{}, &strictModeSkipErr{fieldName: "metadata.name"}
		}
	}

	// Check before handling if we're about to filter resources
	hasIncluded, hasExcluded := len(s.opts.IncludedKinds) > 0, len(s.opts.ExcludedKinds) > 0

	// Check if we have a Kubernetes kind
	if k8smeta.Kind == "" && (hasIncluded || hasExcluded) {
		return yamlFile{}, fmt.Errorf("unable to find Kubernetes resource kind in file number %d", s.fileCount)
	}

	// We need to check if the file is skipped by kind
	if hasIncluded || hasExcluded {
		// If we're working with including only, then filter by it
		if hasIncluded && !inSliceIgnoreCase(s.opts.IncludedKinds, k8smeta.Kind) {
			return yamlFile{}, &kindSkipErr{Kind: k8smeta.Kind}
		}

		// Otherwise exclude based on the parameter received
		if hasExcluded && inSliceIgnoreCase(s.opts.ExcludedKinds, k8smeta.Kind) {
			return yamlFile{}, &kindSkipErr{Kind: k8smeta.Kind}
		}
	}

	// Trim the file name
	name := strings.TrimSpace(buf.String())

	// Fix for text/template Go issue #24963, as well as removing any linebreaks
	name = strings.NewReplacer("<no value>", "", "\n", "").Replace(name)

	if str := strings.TrimSuffix(name, filepath.Ext(name)); str == "" {
		return yamlFile{}, fmt.Errorf("file name rendered will yield no file name for YAML file number %d", s.fileCount)
	}

	return yamlFile{filename: name, meta: k8smeta}, nil
}

// inSliceIgnoreCase checks if a string is in a slice, ignoring case
func inSliceIgnoreCase(slice []string, expected string) bool {
	expected = strings.ToLower(expected)

	for _, a := range slice {
		if strings.ToLower(a) == expected {
			return true
		}
	}

	return false
}

// checkStringInMap checks if a string is in a map, and if not, returns an error
func checkStringInMap(local map[string]interface{}, key string) string {
	iface, found := local[key]

	if !found {
		return ""
	}

	str, ok := iface.(string)
	if !ok {
		return ""
	}

	return str
}

// checkKubernetesBasics check if the minimum required keys are there for a Kubernetes Object
func checkKubernetesBasics(manifest map[string]interface{}) kubeObjectMeta {
	var metadata kubeObjectMeta

	metadata.APIVersion = checkStringInMap(manifest, "apiVersion")
	metadata.Kind = checkStringInMap(manifest, "kind")

	if md, found := manifest["metadata"]; found {
		metadata.Name = checkStringInMap(md.(map[string]interface{}), "name")
	}

	return metadata
}
