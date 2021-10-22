package main

import (
	"fmt"
	"os"

	"github.com/patrickdappollonio/kubectl-split/split"
	"github.com/spf13/cobra"
)

var version = "development"

const (
	helpShort = "kubectl-split allows you to split a YAML into multiple subfiles using a pattern."

	helpLong = `kubectl-split allows you to split a YAML into multiple subfiles using a pattern.
For documentation, available functions, and more, visit: https://github.com/patrickdappollonio/kubectl-split.`
)

func root() *cobra.Command {
	opts := split.Options{}

	rootCommand := &cobra.Command{
		Use:           "kubectl-split",
		Short:         helpShort,
		Long:          helpLong,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, args []string) error {
			// If no input file has been provided or it's "-", then
			// point the app to stdin
			if opts.InputFile == "" || opts.InputFile == "-" {
				opts.InputFile = os.Stdin.Name()
			}

			// Create a new instance. This will also perform a basic validation.
			instance, err := split.New(opts)
			if err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			return instance.Execute()
		},
	}

	rootCommand.Flags().StringVarP(&opts.InputFile, "input-file", "f", "", "the input file used to read the initial macro YAML file. If empty or \"-\", stdin is used")
	rootCommand.Flags().StringVarP(&opts.OutputDirectory, "output-dir", "o", ".", "the output directory used to output the splitted files")
	rootCommand.Flags().StringVarP(&opts.GoTemplate, "template", "t", "{{.kind | lower}}-{{.metadata.name}}.yaml", "go template used to generate the file name when creating the resource files in the output directory")
	rootCommand.Flags().BoolVar(&opts.DryRun, "dry-run", false, "if true, no files are created, but the potentially generated files will be printed as the command output")
	rootCommand.Flags().BoolVar(&opts.DebugMode, "debug", false, "enable debug mode")
	rootCommand.Flags().BoolVarP(&opts.StrictKubernetes, "skip-non-k8s", "s", false, "if enabled, any YAMLs that don't contain at least an \"apiVersion\", \"kind\" and \"metadata.name\" will be excluded from the split")

	rootCommand.Flags().MarkHidden("debug")

	return rootCommand
}
