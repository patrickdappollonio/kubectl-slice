package slice

import (
	"fmt"

	"github.com/patrickdappollonio/kubectl-slice/pkg/template"
)

func (s *Split) compileTemplate() error {
	if s.template != nil {
		s.log.Println("Template already compiled, skipping")
		return nil
	}

	s.log.Printf("About to compile template: %q", s.opts.GoTemplate)
	tmpl, err := template.New(s.opts.GoTemplate)
	if err != nil {
		return fmt.Errorf("file name template parse failed: %w", err)
	}

	s.template = tmpl
	return nil
}
