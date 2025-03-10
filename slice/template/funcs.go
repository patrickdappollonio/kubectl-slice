package template

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"os"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var Functions = template.FuncMap{
	"lower":        jsonLower,
	"lowercase":    jsonLower,
	"uppercase":    jsonUpper,
	"upper":        jsonUpper,
	"title":        jsonTitle,
	"sprintf":      fmt.Sprintf,
	"printf":       fmt.Sprintf,
	"trim":         jsonTrimSpace,
	"trimPrefix":   jsonTrimPrefix,
	"trimSuffix":   jsonTrimSuffix,
	"default":      fnDefault,
	"sha1sum":      sha1sum,
	"sha256sum":    sha256sum,
	"str":          strJSON,
	"required":     jsonRequired,
	"env":          env,
	"replace":      jsonReplace,
	"alphanumify":  jsonAlphanumify,
	"alphanumdash": jsonAlphanumdash,
	"dottodash":    jsonDotToDash,
	"dottounder":   jsonDotToUnder,
	"index":        mapValueByIndex,
	"namespaced":   namespaced,
}

// kubernetes built-in stable API-s  shouldn't need external definition
// extracted from vanilla k8s cluster using kubectl api-resources
var namespaceScoped = map[string]map[string]bool{
	"v1": {
		"Binding":               true,
		"ConfigMap":             true,
		"Endpoints":             true,
		"Event":                 true,
		"LimitRange":            true,
		"Namespace":             false,
		"Node":                  false,
		"PersistentVolume":      false,
		"PersistentVolumeClaim": true,
		"Pod":                   true,
		"PodTemplate":           true,
		"ReplicationController": true,
		"ResourceQuota":         true,
		"Secret":                true,
		"Service":               true,
		"ServiceAccount":        true,
	},
	"admissionregistration.k8s.io/v1": {
		"MutatingWebhookConfiguration":     false,
		"ValidatingAdmissionPolicy":        false,
		"ValidatingAdmissionPolicyBinding": false,
		"ValidatingWebhookConfiguration":   false,
	},
	"apiextensions.k8s.io/v1": {
		"CustomResourceDefinition": false,
	},
	"apiregistration.k8s.io/v1": {
		"APIService": false,
	},
	"apps/v1": {
		"ControllerRevision": true,
		"DaemonSet":          true,
		"Deployment":         true,
		"ReplicaSet":         true,
		"StatefulSet":        true,
	},
	"authentication.k8s.io/v1": {
		"SelfSubjectReview": false,
		"TokenReview":       false,
	},
	"authorization.k8s.io/v1": {
		"LocalSubjectAccessReview": true,
		"SelfSubjectAccessReview":  false,
		"SelfSubjectRulesReview":   false,
	},
	"autoscaling/v2": {
		"HorizontalPodAutoscaler": true,
	},
	"batch/v1": {
		"CronJob": true,
		"Job":     true,
	},
	"certificates.k8s.io/v1": {
		"CertificateSigningRequest": false,
	},
	"coordination.k8s.io/v1": {
		"Lease": true,
	},
	"discovery.k8s.io/v1": {
		"EndpointSlice": true,
	},
	"flowcontrol.apiserver.k8s.io/v1": {
		"FlowSchema":                 false,
		"PriorityLevelConfiguration": false,
	},
	"networking.k8s.io/v1": {
		"Ingress":       true,
		"IngressClass":  false,
		"NetworkPolicy": true,
	},
	"node.k8s.io/v1": {
		"RuntimeClass": false,
	},
	"policy/v1": {
		"PodDisruptionBudget": true,
	},
	"rbac.authorization.k8s.io/v1": {
		"ClusterRole":        false,
		"ClusterRoleBinding": false,
		"Role":               true,
		"RoleBinding":        true,
	},
	"scheduling.k8s.io/v1": {
		"PriorityClass": false,
	},
	"storage.k8s.io/v1": {
		"CSIDriver":          false,
		"CSINode":            false,
		"StorageClass":       false,
		"VolumeAttachment":   false,
		"CSIStorageCapacity": true,
	},
}

func namespaced(manifest map[string]interface{}) (bool, error) {
	var apiVersion string
	var kind string
	switch v := manifest["apiVersion"].(type) {
	case string:
		apiVersion = v
	default:
		return false, fmt.Errorf("apiVersion is not a string")
	}
	switch v := manifest["kind"].(type) {
	case string:
		kind = v
	default:
		return false, fmt.Errorf("kind is not a string")
	}
	if v, ok := namespaceScoped[apiVersion]; ok {
		if namespaced, ok := v[kind]; ok {
			return namespaced, nil
		}
	}
	// best effort, assume cluster scoped if unknown gvk
	// and resource doesn't have a namespace declared
	switch v := manifest["metadata"].(type) {
	case map[string]interface{}:
		if _, ok := v["namespace"]; ok {
			return true, nil
		}
	default:
		return false, fmt.Errorf("metadata is not a map")
	}
	return false, nil
}

// mapValueByIndex returns the value of the map at the given index
func mapValueByIndex(index string, m map[string]interface{}) (interface{}, error) {
	if m == nil {
		return nil, fmt.Errorf("map is nil")
	}

	if index == "" {
		return nil, fmt.Errorf("index is empty")
	}

	v, ok := m[index]
	if !ok {
		return nil, fmt.Errorf("map does not contain index %q", index)
	}

	return v, nil
}

// strJSON converts a value received from JSON/YAML to string. Since not all data
// types are supported for JSON, we can limit to just the primitives that are
// not arrays, objects or null; see:
// https://pkg.go.dev/encoding/json#Unmarshal
func strJSON(val interface{}) (string, error) {
	if val == nil {
		return "", nil
	}

	switch a := val.(type) {
	case string:
		return a, nil

	case bool:
		return fmt.Sprintf("%v", a), nil

	case float64:
		return fmt.Sprintf("%v", a), nil

	default:
		return "", fmt.Errorf("unexpected data type %T -- can't convert to string", val)
	}
}

var (
	reAlphaNum = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	reSlugify  = regexp.MustCompile(`[^a-zA-Z0-9-]+`)
)

func jsonAlphanumify(val interface{}) (string, error) {
	s, err := strJSON(val)
	if err != nil {
		return "", err
	}

	return reAlphaNum.ReplaceAllString(s, ""), nil
}

func jsonAlphanumdash(val interface{}) (string, error) {
	s, err := strJSON(val)
	if err != nil {
		return "", err
	}

	return reSlugify.ReplaceAllString(s, ""), nil
}

func jsonDotToDash(val interface{}) (string, error) {
	s, err := strJSON(val)
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(s, ".", "-"), nil
}

func jsonDotToUnder(val interface{}) (string, error) {
	s, err := strJSON(val)
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(s, ".", "_"), nil
}

func jsonReplace(search, replace string, val interface{}) (string, error) {
	s, err := strJSON(val)
	if err != nil {
		return "", err
	}

	return strings.NewReplacer(search, replace).Replace(s), nil
}

func env(key string) string {
	return os.Getenv(strings.ToUpper(key))
}

func jsonRequired(val interface{}) (interface{}, error) {
	if val == nil {
		return nil, fmt.Errorf("argument is marked as required, but it renders to empty")
	}

	s, err := strJSON(val)
	if err != nil {
		return nil, err
	}

	if s == "" {
		return nil, fmt.Errorf("argument is marked as required, but it renders to empty or it's an object or an unsupported type")
	}

	return val, nil
}

func jsonLower(val interface{}) (string, error) {
	s, err := strJSON(val)
	if err != nil {
		return "", err
	}

	return strings.ToLower(s), nil
}

func jsonUpper(val interface{}) (string, error) {
	s, err := strJSON(val)
	if err != nil {
		return "", err
	}

	return strings.ToUpper(s), nil
}

func jsonTitle(val interface{}) (string, error) {
	s, err := strJSON(val)
	if err != nil {
		return "", err
	}

	return cases.Title(language.Und).String(s), nil
}

func jsonTrimSpace(val interface{}) (string, error) {
	s, err := strJSON(val)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(s), nil
}

func jsonTrimPrefix(prefix string, val interface{}) (string, error) {
	s, err := strJSON(val)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(s, prefix), nil
}

func jsonTrimSuffix(suffix string, val interface{}) (string, error) {
	s, err := strJSON(val)
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(s, suffix), nil
}

func fnDefault(defval, val interface{}) (string, error) {
	v, err := strJSON(val)
	if err != nil {
		return "", err
	}

	dv, err := strJSON(defval)
	if err != nil {
		return "", err
	}

	if v != "" {
		return v, nil
	}

	return dv, nil
}

func altStrJSON(val interface{}) (string, error) {
	var buf bytes.Buffer
	if err := yaml.NewEncoder(&buf).Encode(val); err != nil {
		return "", fmt.Errorf("unable to encode object to YAML: %w", err)
	}

	return buf.String(), nil
}

func sha256sum(input interface{}) (string, error) {
	s, err := altStrJSON(input)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:]), nil
}

func sha1sum(input interface{}) (string, error) {
	s, err := altStrJSON(input)
	if err != nil {
		return "", err
	}

	hash := sha1.Sum([]byte(s))
	return hex.EncodeToString(hash[:]), nil
}
