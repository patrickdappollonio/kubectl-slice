package split

import (
	"fmt"
)

func (s *Split) init() error {
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
