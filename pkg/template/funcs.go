package template

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// GetTemplateFunctions returns a map of functions that can be used in Go templates.
// These functions provide string manipulation, conversion, and other utilities
// for customizing the output filenames during the slice operation.
func GetTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"pluralize":    Pluralize,
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
		"indexOrEmpty": mapValueByIndexOrEmpty,
	}
}

// Pluralize adds an "s" to the end of a string if n is not 1.
// This is useful for generating grammatically correct output when dealing with counts.
func Pluralize(s string, n int) string {
	if n == 1 {
		return s
	}
	return s + "s"
}

// mapValueByIndexOrEmpty retrieves a value from a map without returning an error if the key is not found.
func mapValueByIndexOrEmpty(index string, m map[string]interface{}) interface{} {
	if m == nil {
		return ""
	}

	if index == "" {
		return ""
	}

	v, ok := m[index]
	if !ok {
		return ""
	}

	return v
}

// mapValueByIndex retrieves a value from a map and returns an error if the key is not found.
func mapValueByIndex(index string, m map[string]interface{}) (interface{}, error) {
	if m == nil {
		return nil, fmt.Errorf("map is nil")
	}

	if index == "" {
		return nil, fmt.Errorf("map key is empty")
	}

	v, ok := m[index]
	if !ok {
		return nil, fmt.Errorf("key %q not found", index)
	}

	return v, nil
}

// jsonLower converts string input to lowercase. It handles various input types
// by converting them to strings first.
func jsonLower(s interface{}) string {
	return strings.ToLower(toString(s))
}

// jsonUpper converts string input to uppercase. It handles various input types
// by converting them to strings first.
func jsonUpper(s interface{}) string {
	return strings.ToUpper(toString(s))
}

// jsonTitle converts string input to title case. It handles various input types
// by converting them to strings first, then applies proper title casing.
func jsonTitle(s interface{}) string {
	return cases.Title(language.Und).String(toString(s))
}

// jsonTrimSpace trims whitespace from a string
func jsonTrimSpace(s interface{}) string {
	return strings.TrimSpace(toString(s))
}

// jsonTrimPrefix trims a prefix from a string
func jsonTrimPrefix(prefix, s interface{}) string {
	return strings.TrimPrefix(toString(s), toString(prefix))
}

// jsonTrimSuffix trims a suffix from a string
func jsonTrimSuffix(suffix, s interface{}) string {
	return strings.TrimSuffix(toString(s), toString(suffix))
}

// fnDefault returns the default value if the original is empty
func fnDefault(def, orig interface{}) interface{} {
	s := toString(orig)
	if s != "" {
		return s
	}

	return def
}

// sha1sum returns the SHA-1 hash of a string
func sha1sum(s interface{}) string {
	sum := sha1.Sum([]byte(toString(s)))
	return hex.EncodeToString(sum[:])
}

// sha256sum returns the SHA-256 hash of a string
func sha256sum(s interface{}) string {
	sum := sha256.Sum256([]byte(toString(s)))
	return hex.EncodeToString(sum[:])
}

// strJSON converts an object to a JSON string
func strJSON(v interface{}) string {
	b, err := yaml.Marshal(v)
	if err != nil {
		return ""
	}

	return string(b)
}

// jsonRequired returns an error if the value is empty
func jsonRequired(warn string, val interface{}) (interface{}, error) {
	if val == nil {
		return val, fmt.Errorf("%s", warn)
	}

	s := toString(val)
	if s == "" {
		return val, fmt.Errorf("%s", warn)
	}

	return val, nil
}

// env returns the value of an environment variable
func env(key interface{}) string {
	return os.Getenv(toString(key))
}

// jsonReplace replaces all occurrences of a substring
func jsonReplace(old, new string, src interface{}) string {
	return strings.ReplaceAll(toString(src), old, new)
}

// alphanumRegex is a regular expression that matches only alphanumeric characters
var alphanumRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// jsonAlphanumify returns only alphanumeric characters
func jsonAlphanumify(src interface{}) string {
	return alphanumRegex.ReplaceAllString(toString(src), "")
}

// alphanumDashRegex is a regular expression that matches alphanumeric characters and dashes
var alphanumDashRegex = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

// jsonAlphanumdash filters string input to contain only alphanumeric characters and dashes.
// All other characters are removed from the string. Useful for generating safe filenames.
func jsonAlphanumdash(src interface{}) string {
	s := toString(src)
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, ".", "-")
	return alphanumDashRegex.ReplaceAllString(s, "")
}

// jsonDotToDash replaces dots with dashes
func jsonDotToDash(src interface{}) string {
	return strings.ReplaceAll(toString(src), ".", "-")
}

// jsonDotToUnder replaces dots with underscores
func jsonDotToUnder(src interface{}) string {
	return strings.ReplaceAll(toString(src), ".", "_")
}

// toString converts an interface to a string
func toString(s interface{}) string {
	if s == nil {
		return ""
	}

	switch v := s.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", s)
	}
}
