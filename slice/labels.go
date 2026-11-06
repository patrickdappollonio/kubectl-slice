package slice

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// labelValueErr reports a label whose value Kubernetes would reject
// because it isn't a plain string.
type labelValueErr struct {
	key    string
	reason string
}

func (e *labelValueErr) Error() string {
	return fmt.Sprintf("label %q %s", e.key, e.reason)
}

// readLabels returns the labels under "metadata.labels" in contents.
//
// Values are taken as written in the YAML source rather than as decoded,
// so "1.10" stays "1.10" instead of becoming the float 1.1. A null value
// counts as the empty string, which is how the Kubernetes API reads it.
//
// A list or map value has no text form and always returns a *labelValueErr.
// When strict is true, a number, boolean or any other non-string scalar
// does too, since the API server would reject the manifest.
func readLabels(contents []byte, strict bool) (map[string]string, error) {
	var doc struct {
		Metadata struct {
			Labels yaml.Node `yaml:"labels"`
		} `yaml:"metadata"`
	}

	if err := yaml.Unmarshal(contents, &doc); err != nil {
		return nil, fmt.Errorf("unable to read metadata.labels: %w", err)
	}

	node := resolveAlias(&doc.Metadata.Labels)

	switch {
	case node.Kind == 0, node.Kind == yaml.ScalarNode && node.Tag == "!!null":
		return map[string]string{}, nil
	case node.Kind != yaml.MappingNode:
		return nil, fmt.Errorf("metadata.labels must be a map of strings, found %s", describeNode(node))
	}

	labels := make(map[string]string, len(node.Content)/2)

	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i].Value
		value := resolveAlias(node.Content[i+1])

		if value.Kind != yaml.ScalarNode {
			return nil, &labelValueErr{key: key, reason: fmt.Sprintf("must be a string, found %s", describeNode(value))}
		}

		switch value.Tag {
		case "!!str":
			labels[key] = value.Value
		case "!!null":
			labels[key] = ""
		default:
			if strict {
				return nil, &labelValueErr{key: key, reason: fmt.Sprintf("must be a string, found %s %q", describeNode(value), value.Value)}
			}
			labels[key] = value.Value
		}
	}

	return labels, nil
}

func resolveAlias(n *yaml.Node) *yaml.Node {
	for n.Kind == yaml.AliasNode && n.Alias != nil {
		n = n.Alias
	}
	return n
}

func describeNode(n *yaml.Node) string {
	switch n.Kind {
	case yaml.SequenceNode:
		return "a list"
	case yaml.MappingNode:
		return "a map"
	}

	switch n.Tag {
	case "!!int", "!!float":
		return "a number"
	case "!!bool":
		return "a boolean"
	case "!!timestamp":
		return "a timestamp"
	}

	return fmt.Sprintf("a value of type %s", n.Tag)
}
