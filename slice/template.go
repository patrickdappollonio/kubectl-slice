package slice

import (
	"errors"
	"fmt"
	"strings"
	"text/template"

	local "github.com/patrickdappollonio/kubectl-slice/slice/template"
)

func (s *Split) compileTemplate() error {
	s.log.Printf("About to compile template: %q", s.opts.GoTemplate)
	t, err := template.New("split").Funcs(local.Functions).Parse(s.opts.GoTemplate)
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
			Err:  errors.New(strings.TrimSpace(s[pos+1:])),
		}
	}

	return err
}
