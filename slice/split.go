package slice

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/patrickdappollonio/kubectl-slice/pkg/kubernetes"
	"github.com/patrickdappollonio/kubectl-slice/pkg/logger"
	"github.com/patrickdappollonio/kubectl-slice/pkg/template"
)

// Split is a Kubernetes Split instance. Each instance has its own template
// used to generate the resource names when saving to disk. Because of this,
// avoid reusing the same instance of Split
type Split struct {
	opts     Options
	log      logger.Logger
	template *template.Renderer
	data     *bytes.Buffer

	filesFound []kubernetes.YAMLFile
	fileCount  int
}

// New creates a new Split instance with the options set
func New(opts Options) (*Split, error) {
	s := &Split{
		log: log.New(io.Discard, "[debug] ", log.Lshortfile),
	}

	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}

	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	if opts.DebugMode {
		s.log.SetOutput(opts.Stderr)
	}

	s.opts = opts

	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}
