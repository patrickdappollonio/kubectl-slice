package template

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// DefaultTemplateName is the default template for file naming
const DefaultTemplateName = "{{.kind | lower}}-{{.metadata.name}}.yaml"

// Renderer handles template rendering for file names
type Renderer struct {
	tmpl *template.Template
}

// New creates a new template renderer
func New(templateString string) (*Renderer, error) {
	if templateString == "" {
		templateString = DefaultTemplateName
	}

	tmpl, err := template.New("filename").Funcs(GetTemplateFunctions()).Parse(templateString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse template: %w", err)
	}

	return &Renderer{
		tmpl: tmpl,
	}, nil
}

// Execute renders a template with the given data
func (r *Renderer) Execute(data any) (string, error) {
	var buf bytes.Buffer
	if err := r.tmpl.Execute(&buf, data); err != nil {
		return "", improveExecError(err)
	}

	// Get the rendered filename
	name := strings.TrimSpace(buf.String())

	// Fix for text/template Go issue #24963, as well as removing any linebreaks
	name = strings.NewReplacer("<no value>", "", "\n", "").Replace(name)

	return name, nil
}

// improveExecError enhances template execution error messages.
// This uses string comparisons since the Go template engine does not return
// typed error messages.
func improveExecError(err error) error {
	if strings.Contains(err.Error(), "can't evaluate field") {
		return fmt.Errorf("%w (this usually means the field does not exist in the YAML)", err)
	}

	return err
}
