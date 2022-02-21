package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/muesli/coral"
	"github.com/patrickdappollonio/kubectl-slice/slice"
)

var version = "development"

const (
	helpShort = "kubectl-slice allows you to split a YAML into multiple subfiles using a pattern."

	helpLong = `kubectl-slice allows you to split a YAML into multiple subfiles using a pattern.
For documentation, available functions, and more, visit: https://github.com/patrickdappollonio/kubectl-slice.`
)

var examples = []string{
	"kubectl-slice -f foo.yaml -o ./ --include-kind Pod,Namespace",
	"kubectl-slice -f foo.yaml -o ./ --exclude-kind Pod",
	"kubectl-slice -f foo.yaml -o ./ --exclude-name *-svc",
	"kubectl-slice -f foo.yaml --exclude-name *-svc --stdout",
}

func generateExamples([]string) string {
	var s bytes.Buffer
	for pos, v := range examples {
		s.WriteString(fmt.Sprintf("  %s", v))

		if pos != len(examples)-1 {
			s.WriteString("\n")
		}
	}

	return s.String()
}

func root() *coral.Command {
	opts := slice.Options{}

	rootCommand := &coral.Command{
		Use:           "kubectl-slice",
		Short:         helpShort,
		Long:          helpLong,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Example:       generateExamples(examples),
		RunE: func(_ *coral.Command, args []string) error {
			// If no input file has been provided or it's "-", then
			// point the app to stdin
			if opts.InputFile == "" || opts.InputFile == "-" {
				opts.InputFile = os.Stdin.Name()
			}

			// Create a new instance. This will also perform a basic validation.
			instance, err := slice.New(opts)
			if err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			return instance.Execute()
		},
	}

	rootCommand.Flags().StringVarP(&opts.InputFile, "input-file", "f", "", "the input file used to read the initial macro YAML file; if empty or \"-\", stdin is used")
	rootCommand.Flags().StringVarP(&opts.OutputDirectory, "output-dir", "o", "", "the output directory used to output the splitted files")
	rootCommand.Flags().StringVarP(&opts.GoTemplate, "template", "t", slice.DefaultTemplateName, "go template used to generate the file name when creating the resource files in the output directory")
	rootCommand.Flags().BoolVar(&opts.DryRun, "dry-run", false, "if true, no files are created, but the potentially generated files will be printed as the command output")
	rootCommand.Flags().BoolVar(&opts.DebugMode, "debug", false, "enable debug mode")
	rootCommand.Flags().BoolVarP(&opts.Quiet, "quiet", "q", false, "if true, no output is written to stdout/err")
	rootCommand.Flags().StringSliceVar(&opts.IncludedKinds, "include-kind", nil, "resource kind to include in the output (singular, case insensitive, glob supported)")
	rootCommand.Flags().StringSliceVar(&opts.ExcludedKinds, "exclude-kind", nil, "resource kind to exclude in the output (singular, case insensitive, glob supported)")
	rootCommand.Flags().StringSliceVar(&opts.IncludedNames, "include-name", nil, "resource name to include in the output (singular, case insensitive, glob supported)")
	rootCommand.Flags().StringSliceVar(&opts.ExcludedNames, "exclude-name", nil, "resource name to exclude in the output (singular, case insensitive, glob supported)")
	rootCommand.Flags().BoolVarP(&opts.StrictKubernetes, "skip-non-k8s", "s", false, "if enabled, any YAMLs that don't contain at least an \"apiVersion\", \"kind\" and \"metadata.name\" will be excluded from the split")
	rootCommand.Flags().BoolVar(&opts.SortByKind, "sort-by-kind", false, "if enabled, resources are sorted by Kind, a la Helm, before saving them to disk")
	rootCommand.Flags().BoolVar(&opts.OutputToStdout, "stdout", false, "if enabled, no resource is written to disk and all resources are printed to stdout instead")

	rootCommand.Flags().MarkHidden("debug")

	return rootCommand
}
