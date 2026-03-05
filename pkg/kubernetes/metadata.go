package kubernetes

import (
	"strings"
)

// ObjectMeta represents the metadata for a Kubernetes object
type ObjectMeta struct {
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
	Group      string
}

// GetGroupFromAPIVersion extracts the group from the APIVersion field
func (k *ObjectMeta) GetGroupFromAPIVersion() string {
	fields := strings.Split(k.APIVersion, "/")
	if len(fields) == 2 {
		return strings.ToLower(fields[0])
	}

	return ""
}

// Empty checks if all fields in the metadata are empty
func (k *ObjectMeta) Empty() bool {
	return k.APIVersion == "" && k.Kind == "" && k.Name == "" && k.Namespace == ""
}

// String returns a string representation of the metadata
func (k *ObjectMeta) String() string {
	return strings.TrimSpace(strings.Join([]string{
		"kind " + k.Kind,
		"name " + k.Name,
		"apiVersion " + k.APIVersion,
	}, ", "))
}

// CheckStringInMap checks if a string is in a map, and returns its value if found
func CheckStringInMap(local map[string]interface{}, key string) string {
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

// ExtractMetadata extracts Kubernetes metadata from a YAML manifest
func ExtractMetadata(manifest map[string]interface{}) *ObjectMeta {
	metadata := &ObjectMeta{
		APIVersion: CheckStringInMap(manifest, "apiVersion"),
		Kind:       CheckStringInMap(manifest, "kind"),
	}

	if md, found := manifest["metadata"]; found {
		if mdMap, ok := md.(map[string]interface{}); ok {
			metadata.Name = CheckStringInMap(mdMap, "name")
			metadata.Namespace = CheckStringInMap(mdMap, "namespace")
		}
	}

	return metadata
}
