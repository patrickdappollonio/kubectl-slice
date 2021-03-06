package slice

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mb0/glob"
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
	var (
		hasKindIncluded = len(s.opts.IncludedKinds) > 0
		hasKindExcluded = len(s.opts.ExcludedKinds) > 0
		hasNameIncluded = len(s.opts.IncludedNames) > 0
		hasNameExcluded = len(s.opts.ExcludedNames) > 0
	)

	s.log.Printf(
		"Applying filters -> kindIncluded: %v; kindExcluded: %v; nameIncluded: %v; nameExcluded: %v",
		s.opts.IncludedKinds, s.opts.ExcludedKinds, s.opts.IncludedNames, s.opts.ExcludedNames,
	)

	s.log.Printf("Found K8s meta -> %#v", k8smeta)

	// Check if we have a Kubernetes kind and we're requesting inclusion or exclusion
	if k8smeta.Kind == "" && (hasKindIncluded || hasKindExcluded) {
		return yamlFile{}, fmt.Errorf("unable to find Kubernetes \"kind\" field in file number %d", s.fileCount)
	}

	// Check if we have a Kubernetes name and we're requesting inclusion or exclusion
	if k8smeta.Name == "" && (hasNameIncluded || hasNameExcluded) {
		return yamlFile{}, fmt.Errorf("unable to find Kubernetes \"metadata.name\" field in file number %d", s.fileCount)
	}

	// We need to check if the file is skipped by kind
	if hasKindIncluded || hasKindExcluded || hasNameIncluded || hasNameExcluded {
		// If we're working with including only specific kinds, then filter by it
		if hasKindIncluded && !inSliceIgnoreCaseGlob(s.opts.IncludedKinds, k8smeta.Kind) {
			return yamlFile{}, &skipErr{kind: "kind", name: k8smeta.Kind}
		}

		// Otherwise exclude kinds based on the parameter received
		if hasKindExcluded && inSliceIgnoreCaseGlob(s.opts.ExcludedKinds, k8smeta.Kind) {
			return yamlFile{}, &skipErr{kind: "kind", name: k8smeta.Kind}
		}

		// If we're working with including only specific names, then filter by it
		if hasNameIncluded && !inSliceIgnoreCaseGlob(s.opts.IncludedNames, k8smeta.Name) {
			return yamlFile{}, &skipErr{kind: "name", name: k8smeta.Name}
		}

		// Otherwise exclude names based on the parameter received
		if hasNameExcluded && inSliceIgnoreCaseGlob(s.opts.ExcludedNames, k8smeta.Name) {
			return yamlFile{}, &skipErr{kind: "name", name: k8smeta.Name}
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

// inSliceIgnoreCaseGlob checks if a string is in a slice, ignoring case and
// allowing the use of a glob pattern
func inSliceIgnoreCaseGlob(slice []string, expected string) bool {
	expected = strings.ToLower(expected)

	for _, pattern := range slice {
		pattern = strings.ToLower(pattern)

		if match, _ := glob.Match(pattern, expected); match {
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
