package split

import (
	"fmt"
)

func (s *Split) init() error {
	if len(s.opts.IncludedKinds) > 0 && len(s.opts.ExcludedKinds) > 0 {
		return fmt.Errorf("cannot specify both included and excluded kinds")
	}

	s.log.Printf("Loading file %s", s.opts.InputFile)
	buf, err := loadfile(s.opts.InputFile)
	if err != nil {
		return err
	}

	s.data = buf

	if s.opts.OutputDirectory == "" {
		return fmt.Errorf("output directory is empty")
	}

	if err := s.compileTemplate(); err != nil {
		return err
	}

	return nil
}
