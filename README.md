# `kubectl-slice`: split Kubernetes YAMLs into files

[![Downloads](https://img.shields.io/github/downloads/patrickdappollonio/kubectl-slice/total?color=blue&logo=github&style=flat-square)](https://github.com/patrickdappollonio/kubectl-slice/releases)

- [`kubectl-slice`: split Kubernetes YAMLs into files](#kubectl-slice-split-kubernetes-yamls-into-files)
  - [Installation](#installation)
    - [Using `krew`](#using-krew)
    - [Download and install manually](#download-and-install-manually)
  - [Usage](#usage)
  - [Why `kubectl-slice`?](#why-kubectl-slice)
  - [Passing configuration options to `kubectl-slice`](#passing-configuration-options-to-kubectl-slice)
  - [Including and excluding manifests from the output](#including-and-excluding-manifests-from-the-output)
  - [Examples](#examples)
  - [Contributing \& Roadmap](#contributing--roadmap)

`kubectl-slice` is a neat tool that allows you to split a single multi-YAML Kubernetes manifest into multiple subfiles using a naming convention you choose. This is done by parsing the YAML code and allowing you to access any key from the YAML object [using Go Templates](https://pkg.go.dev/text/template).

By default, `kubectl-slice` will split your files into multiple subfiles following this naming convention that you can configure to your liking:

```handlebars
{{.kind | lower}}-{{.metadata.name}}.yaml
```

That is, the Kubernetes kind -- in this case, the value `Namespace` -- lowercased, followed by a dash, followed by the resource name -- in this case, the value `production`:

```text
namespace-production.yaml
```

If your YAML includes multiple files, for example:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-ingress
---
apiVersion: v1
kind: Namespace
metadata:
  name: production
```

Then the following files will be created:

```text
$ kubectl-slice --input-file=input.yaml --output-dir=.
Wrote pod-nginx-ingress.yaml -- 58 bytes.
Wrote namespace-production.yaml -- 61 bytes.
2 files generated.
```

You can customize the file name to your liking, by using the `--template` flag.

## Installation

`kubectl-slice` can be used as a standalone tool or through `kubectl`, as a plugin.

### Using `krew`

`kubectl-slice` is available as a [krew plugin](https://krew.sigs.k8s.io/docs/user-guide/installing-plugins/).

To install, use the following command:

```bash
kubectl krew install slice
```

### Download and install manually

Download the latest release for your platform from the [Releases page](https://github.com/patrickdappollonio/kubectl-slice/releases), then extract and move the `kubectl-slice` binary to any place in your `$PATH`. If you have `kubectl` installed, you can use both `kubectl-slice` and `kubectl slice` (note in the later the absence of the `-`).

## Usage

```text
kubectl-slice allows you to split a YAML into multiple subfiles using a pattern.
For documentation, available functions, and more, visit: https://github.com/patrickdappollonio/kubectl-slice.

Usage:
  kubectl-slice [flags]

Examples:
  kubectl-slice -f foo.yaml -o ./ --include-kind Pod,Namespace
  kubectl-slice -f foo.yaml -o ./ --exclude-kind Pod
  kubectl-slice -f foo.yaml -o ./ --exclude-name *-svc
  kubectl-slice -f foo.yaml --exclude-name *-svc --stdout
  kubectl-slice -f foo.yaml --include Pod/* --stdout
  kubectl-slice -f foo.yaml --exclude deployment/kube* --stdout
  kubectl-slice --config config.yaml

Flags:
  -c, --config string          path to the config file
      --dry-run                if true, no files are created, but the potentially generated files will be printed as the command output
      --exclude strings        resource name to exclude in the output (format <kind>/<name>, case insensitive, glob supported)
      --exclude-kind strings   resource kind to exclude in the output (singular, case insensitive, glob supported)
      --exclude-name strings   resource name to exclude in the output (singular, case insensitive, glob supported)
  -h, --help                   help for kubectl-slice
      --include strings        resource name to include in the output (format <kind>/<name>, case insensitive, glob supported)
      --include-kind strings   resource kind to include in the output (singular, case insensitive, glob supported)
      --include-name strings   resource name to include in the output (singular, case insensitive, glob supported)
  -f, --input-file string      the input file used to read the initial macro YAML file; if empty or "-", stdin is used
  -o, --output-dir string      the output directory used to output the splitted files
  -q, --quiet                  if true, no output is written to stdout/err
  -s, --skip-non-k8s           if enabled, any YAMLs that don't contain at least an "apiVersion", "kind" and "metadata.name" will be excluded from the split
      --sort-by-kind           if enabled, resources are sorted by Kind, a la Helm, before saving them to disk
      --stdout                 if enabled, no resource is written to disk and all resources are printed to stdout instead
  -t, --template string        go template used to generate the file name when creating the resource files in the output directory (default "{{.kind | lower}}-{{.metadata.name}}.yaml")
  -v, --version                version for kubectl-slice
```

## Why `kubectl-slice`?

See [why `kubectl-slice`?](docs/why.md) for more information.

## Passing configuration options to `kubectl-slice`

Besides command-line flags, you can also use environment variables and a YAML configuration file to pass options to `kubectl-slice`. See [the documentation for configuration options](docs/configuring-cli.md) for details about both, including precedence.

## Including and excluding manifests from the output

Including or excluding manifests from the output via `metadata.name` or `kind` is possible. Globs are supported in both cases. See [the documentation for including and excluding items](docs/including-excluding-items.md) for more information.

## Examples

See [examples](docs/examples.md) for more information.

## Contributing & Roadmap

Pull requests are welcomed! So far, looking for help with the following items, which are also part of the roadmap:

* Adding unit tests
* Improving the YAML file-by-file parser, right now it works by buffering line by line
* Adding support to install through `brew`
* [Adding new features marked as `enhancement`](//github.com/patrickdappollonio/kubectl-slice/issues?q=is%3Aissue+is%3Aopen+label%3Aenhancement)
