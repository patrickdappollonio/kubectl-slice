package slice

import "fmt"

func (s *Split) WriteStderr(format string, args ...interface{}) {
	if s.opts.Quiet {
		return
	}

	_, _ = fmt.Fprintf(s.opts.Stderr, format+"\n", args...)
}

func (s *Split) WriteStdout(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(s.opts.Stdout, format+"\n", args...)
}
