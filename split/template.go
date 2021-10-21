package split

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

func (s *Split) compileTemplate() error {
	s.log.Println("About to compile template")
	t, err := template.New("split").Funcs(templateFuncs()).Parse(s.opts.GoTemplate)
	if err != nil {
		return fmt.Errorf("file name template parse failed: %w", improveExecError(err))
	}

	s.template = t
	return nil
}

func improveExecError(err error) error {
	// Before you start screaming because I'm handling an error using strings,
	// consider that there's a longstanding open TODO to improve template.ExecError
	// to be more meaningful:
	// https://github.com/golang/go/blob/go1.17/src/text/template/exec.go#L107-L109

	if _, ok := err.(template.ExecError); !ok {
		if !strings.HasPrefix(err.Error(), "template:") {
			return err
		}
	}

	s := err.Error()

	if pos := strings.LastIndex(s, ":"); pos >= 0 {
		return template.ExecError{
			Name: "",
			Err:  fmt.Errorf(strings.TrimSpace(s[pos+1:])),
		}
	}

	return err
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
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
	}
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
		return nil, fmt.Errorf("argument is marked as required, but it renders to empty")
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

	return strings.Title(s), nil
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
