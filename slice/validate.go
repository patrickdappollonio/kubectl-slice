package slice

import (
	"fmt"
)

func (s *Split) init() error {
	if len(s.opts.IncludedKinds) > 0 && len(s.opts.ExcludedKinds) > 0 {
		return fmt.Errorf("cannot specify both included and excluded kinds")
	}

	if len(s.opts.IncludedNames) > 0 && len(s.opts.ExcludedNames) > 0 {
		return fmt.Errorf("cannot specify both included and excluded names")
	}

	s.log.Printf("Loading file %s", s.opts.InputFile)
	buf, err := loadfile(s.opts.InputFile)
	if err != nil {
		return err
	}

	s.data = buf

	if s.opts.OutputToStdout {
		if s.opts.OutputDirectory != "" {
			return fmt.Errorf("cannot specify both output to stdout and output to file: output directory is present")
		}
	} else {
		if s.opts.OutputDirectory == "" {
			return fmt.Errorf("output directory is empty")
		}
	}

	if err := s.compileTemplate(); err != nil {
		return err
	}

	return nil
}
