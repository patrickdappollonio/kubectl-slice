package slice

import (
	"fmt"
	"regexp"
)

var regKN = regexp.MustCompile(`^[^/]+/[^/]+$`)


func (s *Split) init() error {
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

	return s.validateFilters()
}

func (s *Split) validateFilters() error {
	if len(s.opts.IncludedKinds) > 0 && len(s.opts.ExcludedKinds) > 0 {
		return fmt.Errorf("cannot specify both included and excluded kinds")
	}

	if len(s.opts.IncludedNames) > 0 && len(s.opts.ExcludedNames) > 0 {
		return fmt.Errorf("cannot specify both included and excluded names")
	}

	if len(s.opts.Included) > 0 && len(s.opts.Excluded) > 0 {
		return fmt.Errorf("cannot specify both included and excluded")
	}

	// Merge all filters into excluded and included.
	for _, v := range s.opts.IncludedKinds {
		s.opts.Included = append(s.opts.Included, fmt.Sprintf("%s/*", v))
	}

	for _, v := range s.opts.ExcludedKinds {
		s.opts.Excluded = append(s.opts.Excluded, fmt.Sprintf("%s/*", v))
	}

	for _, v := range s.opts.IncludedNames {
		s.opts.Included = append(s.opts.Included, fmt.Sprintf("*/%s", v))
	}

	for _, v := range s.opts.ExcludedNames {
		s.opts.Excluded = append(s.opts.Excluded, fmt.Sprintf("*/%s", v))
	}

	// Validate included and excluded filters.
	for _, included := range s.opts.Included {
		if !regKN.MatchString(included) {
			return fmt.Errorf("invalid included pattern %q should be <kind>/<name>", included)
		}
	}

	for _, excluded := range s.opts.Excluded {
		if !regKN.MatchString(excluded) {
			return fmt.Errorf("invalid excluded pattern %q should be <kind>/<name>", excluded)
		}
	}

	return nil
}
