package slice

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/patrickdappollonio/kubectl-slice/pkg/files"
)

var (
	regKN      = regexp.MustCompile(`^[^/]+/[^/]+$`)
	extensions = []string{".yaml", ".yml"}
)

func (s *Split) init() error {
	s.log.Printf("Initializing with settings: %#v", s.opts)

	if s.opts.InputFile != "" && s.opts.InputFolder != "" {
		return fmt.Errorf("cannot specify both input file and input folder")
	}

	if s.opts.InputFile == "" && s.opts.InputFolder == "" {
		return fmt.Errorf("input file or input folder is required")
	}

	var buf *bytes.Buffer

	if s.opts.InputFile != "" {
		s.log.Printf("Loading file %s", s.opts.InputFile)
		var err error
		buf, err = files.LoadFile(s.opts.InputFile)
		if err != nil {
			return err
		}
	}

	if s.opts.InputFolder != "" {
		exts := extensions
		s.opts.InputFolder = filepath.Clean(s.opts.InputFolder)

		if len(s.opts.InputFolderExt) > 0 {
			exts = s.opts.InputFolderExt
		}

		s.log.Printf("Loading folder %q", s.opts.InputFolder)
		var err error
		var count int
		buf, count, err = files.LoadFolder(exts, s.opts.InputFolder, s.opts.Recurse)
		if err != nil {
			return err
		}
		s.log.Printf("Found %d files in folder %q", count, s.opts.InputFolder)
	}

	if buf == nil || buf.Len() == 0 {
		return fmt.Errorf("no data found in input file or folder")
	}

	s.data = buf

	if s.opts.OutputToStdout {
		if s.opts.OutputDirectory != "" {
			return fmt.Errorf("cannot specify both output to stdout and output to file: output directory flag is set")
		}
	} else {
		if s.opts.OutputDirectory == "" {
			return fmt.Errorf("output directory flag is empty or not set")
		}
	}

	if err := s.compileTemplate(); err != nil {
		return err
	}

	return s.validateFilters()
}

func (s *Split) validateFilters() error {
	if len(s.opts.IncludedKinds) > 0 && s.opts.AllowEmptyKinds {
		return fmt.Errorf("cannot specify both included kinds and allow empty kinds")
	}

	if len(s.opts.ExcludedKinds) > 0 && s.opts.AllowEmptyKinds {
		return fmt.Errorf("cannot specify both excluded kinds and allow empty kinds")
	}

	if len(s.opts.IncludedNames) > 0 && s.opts.AllowEmptyNames {
		return fmt.Errorf("cannot specify both included names and allow empty names")
	}

	if len(s.opts.ExcludedNames) > 0 && s.opts.AllowEmptyNames {
		return fmt.Errorf("cannot specify both excluded names and allow empty names")
	}

	if len(s.opts.IncludedKinds) > 0 && len(s.opts.ExcludedKinds) > 0 {
		return fmt.Errorf("cannot specify both included and excluded kinds")
	}

	if len(s.opts.IncludedNames) > 0 && len(s.opts.ExcludedNames) > 0 {
		return fmt.Errorf("cannot specify both included and excluded names")
	}

	if len(s.opts.Included) > 0 && len(s.opts.Excluded) > 0 {
		return fmt.Errorf("cannot specify both included and excluded")
	}

	if len(s.opts.ExcludedGroups) > 0 && len(s.opts.IncludedGroups) > 0 {
		return fmt.Errorf("cannot specify both included and excluded groups")
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
