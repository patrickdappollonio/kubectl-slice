# `kubectl-slice`: split Kubernetes YAMLs into files

[![Tests passing](https://img.shields.io/github/workflow/status/patrickdappollonio/kubectl-slice/Testing/master?logo=github&style=flat-square)](https://github.com/patrickdappollonio/kubectl-slice/actions)
[![Downloads](https://img.shields.io/github/downloads/patrickdappollonio/kubectl-slice/total?color=blue&logo=github&style=flat-square)](https://github.com/patrickdappollonio/kubectl-slice/releases)

- [`kubectl-slice`: split Kubernetes YAMLs into files](#kubectl-slice-split-kubernetes-yamls-into-files)
  - [Installation](#installation)
    - [Using `krew`](#using-krew)
    - [Download and install manually](#download-and-install-manually)
  - [Usage](#usage)
    - [Flags](#flags)
  - [Why `kubectl-slice`?](#why-kubectl-slice)
  - [Examples](#examples)
  - [Contributing & Roadmap](#contributing--roadmap)

`kubectl-slice` is a neat tool that allows you to split a single multi-YAML Kubernetes manifest into multiple subfiles using a naming convention you choose. This is done by parsing the YAML code and giving you the option to access any key from the YAML object [using Go Templates](https://pkg.go.dev/text/template).

By default, `kubectl-slice` will split your files into multiple subfiles following this naming convention:

```handlebars
{{.kind | lower}}-{{.metadata.name}}.yaml
```

That is, the Kubernets kind -- say, `Namespace` -- lowercased, followed by a dash, followed by the resource name -- say, `production`:

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
$ kubectl-slice --input-file=example.yaml
Wrote pod-nginx-ingress.yaml -- 57 bytes.
Wrote namespace-production.yaml -- 60 bytes.
2 files generated.
```

You can customize the file name to your liking, by using the `--template` flag.

## Installation

`kubectl-slice` can be used as a standalone tool or through `kubectl`, as a plugin.

### Using `krew`

`kubectl-slice` is available as a [krew plugin](https://krew.sigs.k8s.io/docs/user-guide/installing-plugins/).

To install, use the following command:

```bash
kubectl krew install kubectl-slice
```

### Download and install manually

Download the latest release for your platform from the [Releases page](https://github.com/patrickdappollonio/kubectl-slice/releases), then extract and move the `kubectl-slice` binary to any place in your `$PATH`. If you have `kubectl` installed, you can use both `kubectl-slice` and `kubectl split` (note in the later the absence of the `-`).

## Usage

```text
kubectl-slice allows you to split a YAML into multiple subfiles using a pattern.
For documentation, available functions, and more, visit: https://github.com/patrickdappollonio/kubectl-slice.

Usage:
  kubectl-slice [flags]

Examples:
kubectl-slice -f foo.yaml -o ./ -i Pod,Namespace

Flags:
      --dry-run                if true, no files are created, but the potentially generated files will be printed as the command output
  -e, --exclude-kind strings   kinds to exclude in the output (singular, case insensitive); if empty, all Kubernetes object kinds are excluded
  -h, --help                   help for kubectl-slice
  -i, --include-kind strings   kinds to include in the output (singular, case insensitive); if empty, all Kubernetes object kinds are included
  -f, --input-file string      the input file used to read the initial macro YAML file; if empty or "-", stdin is used
  -o, --output-dir string      the output directory used to output the splitted files (default ".")
  -s, --skip-non-k8s           if enabled, any YAMLs that don't contain at least an "apiVersion", "kind" and "metadata.name" will be excluded from the split
  -t, --template string        go template used to generate the file name when creating the resource files in the output directory (default "{{.kind | lower}}-{{.metadata.name}}.yaml")
  -v, --version                version for kubectl-slice
```

### Flags

* `--dry-run`:
  * Allows the program to execute but not save anything to files. The output will show what potential files would be created.
* `--input-file`:
  * The input file to read as YAML multi-file. If this value is empty or set to `-`, `stdin` is used instead. Even after processing, the original file is preserved as much as possible, and that includes comments, YAML arrays, and formatting.
* `--output-dir`:
  * The output directory where the files must be saved. By default is set to the current directory. You can use this in combination with `--template` to control where your files will land once split. If the folder does not exist, it will be created.
* `--template`:
  * A Go Text Template used to generate the splitted file names. You can access any field from your YAML files -- even fields that don't exist, although they will render as `""` -- and use this to your advantage. Consider the following:
    * There's a check to validate that, after rendering the file name, there's at least a file name.
    * Unix linebreaks (`\n`) are removed from the generated file name, thus allowing you to use multiline Go Templates if needed.
    * You can use any of the built-in [Template Functions](docs/template_functions.md#template-functions) to your advantage.
    * If multiple files from your YAML generate the same file name, all YAMLs that match this file name will be appended.
    * If the rendered file name includes a path separator, subfolders under `--output-dir` will be created.
    * If a file already exists in `--output-directory` under this generated file name, their contents will be replaced.
* `--exclude-kind`:
  * A case-insensitive, comma-separated list of Kubernetes object kinds to exclude from the output.
  * You can also repeat the parameter multiple times to achieve the same effect (`--exclude-kind pod --exclude-kind deployment`)
* `--include-kind`:
  * A case-insensitive, comma-separated list of Kubernetes object kinds to include in the output. Any other Kubernetes object kinds will be excluded.
  * You can also repeat the parameter multiple times to achieve the same effect (`--include-kind pod --include-kind deployment`)
* `--skip-non-k8s`:
  * If enabled, any YAMLs that don't contain at least an `apiVersion`, `kind` and `metadata.name` will be excluded from the split
  * There are no attempts to validate how correct these fields are. For example, there's no check to validate that `apiVersion` exists in a Kubernetes cluster, or whether this `apiVersion` is valid: `"example\foo"`.
    * It's useful, however, if alongside the original YAML you suspect there might be some non Kubernetes YAMLs being generated.

## Why `kubectl-slice`?

See [why `kubectl-slice`?](docs/why.md) for more information.

## Examples

See [examples](docs/examples.md) for more information.

## Contributing & Roadmap

Pull requests are welcomed! So far, looking for help with the following items, which are also part of the roadmap:

* Adding unit tests
* Improving the YAML file-by-file parser, right now it works by buffering line by line
* Adding support to install through `brew`
* Functions to allow accessing labels and annotations on a way different than the dot-notation from Go templates
