package slice

import "fmt"

// WriteStderr writes formatted output to stderr unless quiet mode is enabled
func (s *Split) WriteStderr(format string, args ...interface{}) {
	if s.opts.Quiet {
		return
	}

	fmt.Fprintf(s.opts.Stderr, format+"\n", args...)
}

// WriteStdout writes formatted output to stdout
func (s *Split) WriteStdout(format string, args ...interface{}) {
	fmt.Fprintf(s.opts.Stdout, format+"\n", args...)
}
