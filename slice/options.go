package slice

import "io"

// Options configures how the Split operation processes and outputs Kubernetes resources.
// It controls input sources, output destinations, filtering criteria, and formatting options.
type Options struct {
	// Stdout is the writer used for standard output
	Stdout io.Writer
	// Stderr is the writer used for error and debug output
	Stderr io.Writer

	InputFile         string   // the name of the input file to be read
	InputFolder       string   // the name of the input folder to be read
	InputFolderExt    []string // the extensions of the files to be read
	Recurse           bool     // if true, the input folder will be read recursively
	OutputDirectory   string   // the path to the directory where the files will be stored
	PruneOutputDir    bool     // if true, the output directory will be pruned before writing the files
	OutputToStdout    bool     // if true, the output will be written to stdout instead of a file
	GoTemplate        string   // the go template code to render the file names
	DryRun            bool     // if true, no files are created
	DebugMode         bool     // enables debug mode
	Quiet             bool     // disables all writing to stdout/stderr
	IncludeTripleDash bool     // include the "---" separator on resources sliced

	// IncludedKinds is a list of Kubernetes kinds to include (all others will be excluded)
	IncludedKinds []string
	// ExcludedKinds is a list of Kubernetes kinds to exclude (all others will be included)
	ExcludedKinds []string
	// IncludedNames is a list of resource names to include (all others will be excluded)
	IncludedNames []string
	// ExcludedNames is a list of resource names to exclude (all others will be included)
	ExcludedNames []string
	// Included is a list of "kind/name" combinations to include
	Included []string
	// Excluded is a list of "kind/name" combinations to exclude
	Excluded []string

	// StrictKubernetes when enabled, any YAMLs that don't contain at least an "apiVersion", "kind" and "metadata.name" are excluded
	StrictKubernetes bool

	// SortByKind enables sorting of resources by Kubernetes kind importance (follows Helm install order)
	SortByKind bool
	// RemoveFileComments removes auto-generated comments from output files
	RemoveFileComments bool

	// AllowEmptyNames permits resources without a metadata.name field
	AllowEmptyNames bool
	// AllowEmptyKinds permits resources without a kind field
	AllowEmptyKinds bool

	// IncludedGroups is a list of API groups to include (all others will be excluded)
	IncludedGroups []string
	// ExcludedGroups is a list of API groups to exclude (all others will be included)
	ExcludedGroups []string
}
