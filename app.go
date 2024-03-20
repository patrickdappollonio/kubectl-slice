package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/patrickdappollonio/kubectl-slice/slice"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
	"kubectl-slice -f foo.yaml --include Pod/* --stdout",
	"kubectl-slice -f foo.yaml --exclude deployment/kube* --stdout",
	"kubectl-slice --config config.yaml",
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

func root() *cobra.Command {
	opts := slice.Options{}
	var configFile string

	rootCommand := &cobra.Command{
		Use:           "kubectl-slice",
		Short:         helpShort,
		Long:          helpLong,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Example:       generateExamples(examples),

		PreRunE: func(cmd *cobra.Command, args []string) error {
			return bindCobraAndViper(cmd, configFile)
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			// Bind to the appropriate stdout/stderr
			opts.Stdout = cmd.OutOrStdout()
			opts.Stderr = cmd.ErrOrStderr()

			// If no input file has been provided or it's "-", then
			// point the app to stdin
			if opts.InputFile == "" || opts.InputFile == "-" {
				opts.InputFile = os.Stdin.Name()

				// Check if we're receiving data from the terminal
				// or from piped content. Users from piped content
				// won't see this message. Users that might have forgotten
				// setting the flags correctly will see this message.
				if !opts.Quiet {
					if fi, err := os.Stdin.Stat(); err == nil && fi.Mode()&os.ModeNamedPipe == 0 {
						fmt.Fprintln(opts.Stderr, "Receiving data from the terminal. Press CTRL+D when you're done typing or CTRL+C")
						fmt.Fprintln(opts.Stderr, "to exit without processing the content. If you're seeing this by mistake, make")
						fmt.Fprintln(opts.Stderr, "sure the command line flags, environment variables or config file are correct.")
					}
				}
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
	rootCommand.Flags().StringSliceVar(&opts.Included, "include", nil, "resource name to include in the output (format <kind>/<name>, case insensitive, glob supported)")
	rootCommand.Flags().StringSliceVar(&opts.Excluded, "exclude", nil, "resource name to exclude in the output (format <kind>/<name>, case insensitive, glob supported)")
	rootCommand.Flags().BoolVarP(&opts.StrictKubernetes, "skip-non-k8s", "s", false, "if enabled, any YAMLs that don't contain at least an \"apiVersion\", \"kind\" and \"metadata.name\" will be excluded from the split")
	rootCommand.Flags().BoolVar(&opts.SortByKind, "sort-by-kind", false, "if enabled, resources are sorted by Kind, a la Helm, before saving them to disk")
	rootCommand.Flags().BoolVar(&opts.OutputToStdout, "stdout", false, "if enabled, no resource is written to disk and all resources are printed to stdout instead")
	rootCommand.Flags().StringVarP(&configFile, "config", "c", "", "path to the config file")
	rootCommand.Flags().BoolVar(&opts.AllowEmptyKinds, "allow-empty-kinds", false, "if enabled, resources with empty kinds don't produce an error when filtering")
	rootCommand.Flags().BoolVar(&opts.AllowEmptyNames, "allow-empty-names", false, "if enabled, resources with empty names don't produce an error when filtering")
	rootCommand.Flags().BoolVar(&opts.IncludeTripleDash, "include-triple-dash", false, "if enabled, the typical \"---\" YAML separator is included at the beginning of resources sliced")
	rootCommand.Flags().BoolVar(&opts.PruneOutputDir, "prune", false, "if enabled, the output directory will be pruned before writing the files")

	_ = rootCommand.Flags().MarkHidden("debug")
	return rootCommand
}

// envVarPrefix is the prefix used for environment variables.
// Using underscores to ensure compatibility with the shell.
const envVarPrefix = "KUBECTL_SLICE"

// skippedFlags is a list of flags that are not bound through
// Viper. These include things like "help", "version", and of
// course, "config", since it doesn't make sense to say where
// the config file is located in the config file itself.
var skippedFlags = [...]string{
	"help",
	"version",
	"config",
}

// bindCobraAndViper binds the settings loaded by Viper
// to the flags defined in Cobra.
func bindCobraAndViper(cmd *cobra.Command, configFileLocation string) error {
	v := viper.New()

	// If a configuration file has been passed...
	if cmd.Flags().Lookup("config").Changed {
		// ... then set it as the configuration file
		v.SetConfigFile(configFileLocation)

		// then read the configuration file
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read configuration file: %w", err)
		}
	}

	// Handler for potential error
	var err error

	// Recurse through all the variables
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		// Skip the flags that are not bound through Viper
		for _, v := range skippedFlags {
			if v == flag.Name {
				return
			}
		}

		// Normalize key names with underscores instead of dashes
		nameUnderscored := strings.ReplaceAll(flag.Name, "-", "_")
		envVarName := strings.ToUpper(fmt.Sprintf("%s_%s", envVarPrefix, nameUnderscored))

		// Bind the flag to the environment variable
		if val, found := os.LookupEnv(envVarName); found {
			v.Set(nameUnderscored, val)
		}

		// If the CLI flag hasn't been changed, but the value is set in
		// the configuration file, then set the CLI flag to the value
		// from the configuration file
		if !flag.Changed && v.IsSet(nameUnderscored) {
			// Type check for all the supported types
			switch val := v.Get(nameUnderscored).(type) {

			case string:
				_ = cmd.Flags().Set(flag.Name, val)

			case []interface{}:
				var stringified []string
				for _, v := range val {
					stringified = append(stringified, fmt.Sprintf("%v", v))
				}
				_ = cmd.Flags().Set(flag.Name, strings.Join(stringified, ","))

			case bool:
				_ = cmd.Flags().Set(flag.Name, fmt.Sprintf("%t", val))

			case int:
				_ = cmd.Flags().Set(flag.Name, fmt.Sprintf("%d", val))

			default:
				err = fmt.Errorf("unsupported type %T for flag %q", val, nameUnderscored)
				return
			}
		}
	})

	// If an error occurred, return it
	return err
}
