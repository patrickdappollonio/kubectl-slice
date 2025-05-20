package slice

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mb0/glob"
	"gopkg.in/yaml.v3"

	"github.com/patrickdappollonio/kubectl-slice/pkg/errors"
	"github.com/patrickdappollonio/kubectl-slice/pkg/kubernetes"
)

// parseYAMLManifest parses a single YAML file as received by contents. It also renders the
// template needed to generate its name
func (s *Split) parseYAMLManifest(contents []byte) (kubernetes.YAMLFile, error) {
	// All resources we'll handle are Kubernetes manifest, and even those who are lists,
	// they're still Kubernetes Objects of type List, so we can use a map
	manifest := make(map[string]interface{})

	s.log.Println("Parsing YAML from buffer up to this point")
	if err := yaml.Unmarshal(contents, &manifest); err != nil {
		return kubernetes.YAMLFile{}, fmt.Errorf("unable to parse YAML file number %d: %w", s.fileCount, err)
	}

	// Render the name to a buffer using the Go Template
	s.log.Println("Rendering filename template from Go Template")
	name, err := s.template.Execute(manifest)
	if err != nil {
		return kubernetes.YAMLFile{}, fmt.Errorf("unable to render file name for YAML file number %d: %w", s.fileCount, err)
	}

	// Check if file contains the required Kubernetes metadata
	k8smeta := kubernetes.ExtractMetadata(manifest)

	// Check if at least the three fields are not empty
	if err := kubernetes.ValidateRequiredFields(k8smeta, s.opts.StrictKubernetes); err != nil {
		return kubernetes.YAMLFile{}, err
	}

	// Check before handling if we're about to filter resources
	var (
		hasIncluded = len(s.opts.Included) > 0
		hasExcluded = len(s.opts.Excluded) > 0
	)

	s.log.Printf("Applying filters -> Included: %v; Excluded: %v", s.opts.Included, s.opts.Excluded)
	s.log.Printf("Kubernetes metadata found -> %#v", k8smeta)

	// Check if we have a Kubernetes kind and we're requesting inclusion or exclusion
	if k8smeta.Kind == "" && !s.opts.AllowEmptyKinds && (hasIncluded || hasExcluded) {
		return kubernetes.YAMLFile{}, &errors.CantFindFieldErr{FieldName: "kind", FileCount: s.fileCount, Meta: k8smeta}
	}

	// Check if we have a Kubernetes name and we're requesting inclusion or exclusion
	if k8smeta.Name == "" && !s.opts.AllowEmptyNames && (hasIncluded || hasExcluded) {
		return kubernetes.YAMLFile{}, &errors.CantFindFieldErr{FieldName: "metadata.name", FileCount: s.fileCount, Meta: k8smeta}
	}

	// We need to check if the file should be skipped
	if hasExcluded || hasIncluded {
		// If we're working with including only specific resources, then filter by them
		if hasIncluded && !inSliceIgnoreCaseGlob(s.opts.Included, fmt.Sprintf("%s/%s", k8smeta.Kind, k8smeta.Name)) {
			return kubernetes.YAMLFile{}, &errors.SkipErr{Kind: "kind/name", Name: fmt.Sprintf("%s/%s", k8smeta.Kind, k8smeta.Name)}
		}

		// Otherwise exclude resources based on the parameter received
		if hasExcluded && inSliceIgnoreCaseGlob(s.opts.Excluded, fmt.Sprintf("%s/%s", k8smeta.Kind, k8smeta.Name)) {
			return kubernetes.YAMLFile{}, &errors.SkipErr{Kind: "kind/name", Name: fmt.Sprintf("%s/%s", k8smeta.Kind, k8smeta.Name)}
		}
	}

	if len(s.opts.IncludedGroups) > 0 || len(s.opts.ExcludedGroups) > 0 {
		if k8smeta.APIVersion == "" {
			return kubernetes.YAMLFile{}, &errors.CantFindFieldErr{FieldName: "apiVersion", FileCount: s.fileCount, Meta: k8smeta}
		}

		var groups []string
		included := len(s.opts.IncludedGroups) > 0
		if included {
			groups = s.opts.IncludedGroups
		} else {
			groups = s.opts.ExcludedGroups
		}

		if err := kubernetes.CheckGroupInclusion(k8smeta, groups, included); err != nil {
			return kubernetes.YAMLFile{}, err
		}
	}

	if str := strings.TrimSuffix(name, filepath.Ext(name)); str == "" {
		return kubernetes.YAMLFile{}, fmt.Errorf("file name rendered will yield no file name for YAML file number %d (original name: %q, metadata: %v)", s.fileCount, name, k8smeta)
	}

	return kubernetes.YAMLFile{
		Filename: name,
		Meta:     k8smeta,
	}, nil
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
