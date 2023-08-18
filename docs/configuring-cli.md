# Configuring the CLI

- [Configuring the CLI](#configuring-the-cli)
  - [Using a configuration file](#using-a-configuration-file)
  - [Using environment variables](#using-environment-variables)
  - [Using command-line flags](#using-command-line-flags)

There are three ways to configure the CLI. Providing configuration to it has the following processing order:

1. Configuration file
2. Environment variables
3. Command-line flags

The order of precedence dictates how `kubectl-slice` will handle being provided with configuration from multiple sources.

The following example can illustrate this by providing `kubectl-slice` with an input file to process via three different ways:

```bash
KUBECTL_SLICE_INPUT_FILE=2.yaml kubectl-slice -f 3.yaml --config $(echo "input_file: 1.yaml">>config.yaml && echo "config.yaml")
```

You'll notice the error message you get is that the file `3.yaml` doesn't exist. From the configuration file (first precedence), to the environment variable (second precedence), to the command-line flag (third precedence), `kubectl-slice` used `3.yaml`.

Removing now the `-f 3.yaml` you'll see that `kubectl-slice` will use `2.yaml` as the input file, coming from the environment variable. Deleting the environment variable will load the setting from the configuration file.

The order of precedence is useful if you want to provide a default configuration file and then override some of the settings using environment variables or command-line flags.

## Using a configuration file

The configuration file is a YAML file that contains the settings for `kubectl-slice`. The configuration file uses the same format expected by the CLI flags, with the names of the flags being the keys of the YAML file and dashes replaced with underscores.

For example, the `--input-file` flag becomes `input_file:` in the configuration file.

The following is an example of a configuration file with the types defined:

```yaml
input_file: string
output_dir: string
template: string
dry_run: boolean
debug: boolean
quiet: boolean
include_kind: [string]
exclude_kind: [string]
include_name: [string]
exclude_name: [string]
include: [string]
exclude: [string]
skip_non_k8s: bool
sort_by_kind: bool
stdout: bool
```

You can use this file to provide more complex templates by using multiline strings without having to escape special characters, for example:

```yaml
template: >
  {{ .kind | lower }}/{{ .metadata.name | dottodash | replace ":" "-" }}.yaml
```

## Using environment variables

Similarly to what happens with YAML configuration files, we use the same format for environment variables, with the names of the flags being the keys of the environment variable and dashes replaced with underscores. The environment variable's name is also prefixed with `KUBECTL_SLICE`, and the entire key is uppercased.

Here are a few examples of environment variables and their corresponding flags:

| Environment variable       | Flag           |
| -------------------------- | -------------- |
| `KUBECTL_SLICE_INPUT_FILE` | `--input-file` |
| `KUBECTL_SLICE_OUTPUT_DIR` | `--output-dir` |
| `KUBECTL_SLICE_TEMPLATE`   | `--template`   |
| `KUBECTL_SLICE_DRY_RUN`    | `--dry-run`    |
| `KUBECTL_SLICE_DEBUG`      | `--debug`      |

The same values as the YAML counterpart apply. In the case of booleans, `true` or `false` are valid values. In the case of arrays, the values are comma-separated.

## Using command-line flags

The command-line flags are the most straightforward way to configure `kubectl-slice`. You can get an up-to-date list of the available flags by running `kubectl-slice --help`.
