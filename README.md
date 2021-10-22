# `kubectl-split`: split Kubernetes YAMLs into files

![Tests passing](https://img.shields.io/github/workflow/status/patrickdappollonio/kubectl-split/Testing/master?logo=github&style=flat-square) [![Releasing](https://img.shields.io/github/downloads/patrickdappollonio/kubectl-split/latest/total?label=Downloads&style=social)](https://github.com/patrickdappollonio/kubectl-split/releases)

`kubectl-split` is a neat tool that allows you to split a single multi-YAML Kubernetes manifest into multiple subfiles using a naming convention you choose. This is done by parsing the YAML code and giving you the option to access any key from the YAML object [using Go Templates](https://pkg.go.dev/text/template).

By default, `kubectl-split` will split your files into multiple subfiles following this naming convention:

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
$ kubectl-split --input-file=example.yaml
Wrote pod-nginx-ingress.yaml -- 57 bytes.
Wrote namespace-production.yaml -- 60 bytes.
2 files generated.
```

You can customize the file name to your liking, by using the `--template` flag.

## Usage

```text
kubectl-split allows you to split a YAML into multiple subfiles using a pattern.
For documentation, available functions, and more, visit: https://github.com/patrickdappollonio/kubectl-split.

Usage:
  kubectl-split [flags]

Examples:
kubectl-split -f foo.yaml -o ./ -i Pod,Namespace

Flags:
      --dry-run                if true, no files are created, but the potentially generated files will be printed as the command output
  -e, --exclude-kind strings   kinds to exclude in the output (singular, case insensitive); if empty, all Kubernetes object kinds are excluded
  -h, --help                   help for kubectl-split
  -i, --include-kind strings   kinds to include in the output (singular, case insensitive); if empty, all Kubernetes object kinds are included
  -f, --input-file string      the input file used to read the initial macro YAML file; if empty or "-", stdin is used
  -o, --output-dir string      the output directory used to output the splitted files (default ".")
  -s, --skip-non-k8s           if enabled, any YAMLs that don't contain at least an "apiVersion", "kind" and "metadata.name" will be excluded from the split
  -t, --template string        go template used to generate the file name when creating the resource files in the output directory (default "{{.kind | lower}}-{{.metadata.name}}.yaml")
  -v, --version                version for kubectl-split
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
    * You can use any of the built-in [Template Functions](#template-functions) to your advantage.
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

## Why `kubectl-split`?

Multiple services and applications to do GitOps require you to provide a folder similar to this:

```text
.
├── cluster/
│   ├── foo-cluster-role-binding.yaml
│   ├── foo-cluster-role.yaml
│   └── ...
└── namespaces/
    ├── kube-system/
    │   └── ...
    ├── prometheus-monitoring/
    │   └── ...
    └── production/
        ├── foo-role-binding.yaml
        ├── foo-service-account.yaml
        └── foo-deployment.yaml
```

Where resources that are globally scoped live in the `cluster/` folder -- or the folder designated by the service or application -- and namespace-specific resources live inside `namespaces/$NAME/`.

Performing this task on big installations such as applications coming from Helm is a bit daunting, and a manual task. `kubectl-split` can help by allowing you to read a single YAML file which holds multiple YAML manifests, parse each one of them, allow you to use their fields as parameters to generate custom names, then rendering those into individual files in a specific folder.

### Differences with other tools

#### Losing the original file and its format

There are other plugins and apps out there that can split your YAML into multiple sub-YAML files like `kubectl-split`, however, they do it by decoding the YAML, processing it, then re-encode it again, which will lose its original definition. That means that some array pieces, for example, might be encoded to a different output -- while still keeping them as arrays; comments are also lost -- since the decoding to Go, then re-encoding back to YAML will ignore YAML Comments.

`kubectl-split` will keep the original file, and even when it will still parse it into Go to give you the ability to use any of the fields as part of the template for the name, the original file contents are still preserved with no changes, so your comments and the preference on how you render arrays, for example, will remain exactly the same as the original file.

#### Naming format and access to data within YAML

One of the things you can do too with `kubectl-split` that you might not be able to do with other tools is the fact that with `kubectl-split` you can literally access any field from the YAML file. Now, granted, if for example you decide to use an annotation in your YAML as part of the name template, that annotation may exist in _some_ of the YAMLs but perhaps not in all of them, so you have to account for that by providing a [`default`](#default) or using Go Template's `if else` blocks.

Other apps might not allow you to read into the entire YAML, and even more so, they might enforce a convention on some of the fields you are able to access. Resource names, for example, [should follow a Kubernetes standard](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names) which some apps might edit preemptively since they don't make for good or "nice" file names, and as such, replace all dots for underscores. `kubectl-split` will let you provide a template that might render an invalid file name, that's true, but you have [a plethora of functions](#replace) to modify its behavior yourself to something that fits your design better. Perhaps you prefer dashes rather than underscores, and you can do that.

Upcoming versions will improve this even more by allowing annotation access using positions rather than names, for example.

## `kubectl` plugin

Since the application name is `kubectl-split`, adding it to any folder in your `$PATH` will allow you to run it either via its real command name, `kubectl-split`, or as a `kubectl` plugin: `kubectl split`.

`kubectl-split` does not use any configuration from `kubectl`, and it can be used standalone, even without a `$KUBECONFIG`.

## Resources with no namespace

It's very common that Helm charts or even plain YAMLs found online might not contain the namespace, and because of that, the field isn't available in the YAML. Since this tool was created to fit a specific criteria [as seen above](#why-kubectl-split), there's no need to implement this here. However, you can use `kustomize` to quickly add the namespace to your manifest, then run it through `kubectl-split`.

First, create a `kustomization.yaml` file:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: my-namespace
resources:
  - my-file.yaml
```

Replace `my-namespace` for the namespace you want to set, and `my-file.yaml` for the name of the actual file that has no namespaces declared. Then run:

```
kustomize build
```

This will render your new file, namespaces included, to `stdout`. You can pipe this as-is to `kubectl-split`:

```
kustomize build | kubectl-split
```

Keep in mind that is recommended to **not add namespaces** to your YAML resources, to allow users to install in any destination they choose. For example, a namespaceless file called `foo.yaml` can be installed to the namespace `bar` by using:

```
kubectl apply -n bar -f foo.yaml
```

## Accessing fields by name

Any field from the YAML file can be used, however, non-existent fields will render an empty string. This is very common for situations such as rendering a Helm template [where the namespace shouldn't be defined](#resources-with-no-namespace).

If you would rather fail executing `kubectl-split` if a field was not found, consider using the `required` Go Template function. The following template will make `kubectl-split` fail with a non-zero exit code if the namespace of any of the resources is not defined.

```handlebars
{{.metadata.namespace | required}}.yaml
```

If you would rather provide a default for those resources without, say, a namespace, you can use the `default` function:

```handlebars
{{.metadata.namespace | default "global"}}.yaml
```

This will render any resource without a namespace with the name `global.yaml`.

## Conflicting file names

Since it's possible to provide a Go Template for a file name that might be the same for multiple resources, `kubectl-split` will append any YAML that matches by file name to the given file using the `---` separator.

For example, considering the following file name:

```handlebars
{{.metadata.namespace | default "global"}}.yaml
```

Any cluster-scoped resource will be appended into `global.yaml`, while any resource in the namespace `production` will be appended to `production.yaml`.

## Windows `CRLF`

`kubectl-split` does not support Windows line breaks with `CRLF` -- also known as `\r\n`. Consider using Unix line breaks `\n`.

## String conversion

Since `kubectl-split` is built in Go, there's only a handful of primitives that can be read from the YAML manifest. All of these have been hand-picked to be stringified automatically -- in fact, multiple template functions will accept them, and yes, that means you can `lowercase` a number 😅

The decision is intentional and it's due to the fact that's impossible to map any potential Kubernetes resource, given the fact that you can teach Kubernetes new objects using Custom Resource Definitions. Because of that, resources are read as untyped and converted to strings when possible.

The following YAML untyped values are handled, [in accordance with the `json.Unmarshal` documentation](https://pkg.go.dev/encoding/json#Unmarshal):

* `bool`, for JSON booleans
* `float64`, for JSON numbers
* `string`, for JSON strings

## Template Functions

The following template functions are available, with some functions having aliases for convenience:

### `lower`, `lowercase`

Converts the value to string as stated in [String conversion](#string-conversion), then lowercases it.

```handlebars
{{ "Namespace" | lower }}
namespace
```

### `upper`, `uppercase`

Converts the value to string as stated in [String conversion](#string-conversion), then uppercases it.

```handlebars
{{ "Namespace" | upper }}
NAMESPACE
```

### `title`

Converts the value to string as stated in [String conversion](#string-conversion), then capitalize the first character of each word.

```handlebars
{{ "hello world" | title }}
Hello World
```

While available, it's use is discouraged for file names.

### `sprintf`, `printf`

Alias of Go's `fmt.Sprintf`.

```handlebars
{{ printf "number-%d" 20 }}
number-20
```

### `trim`

Converts the value to string as stated in [String conversion](#string-conversion), then removes any whitespace at the beginning or end of the string.

```handlebars
{{ "   hello world    " | trim }}
hello world
```

### `trimPrefix`, `trimSuffix`

Converts the value to string as stated in [String conversion](#string-conversion), then removes either the prefix or the suffix.

Do note that the parameters are flipped from Go's `strings.TrimPrefix` and `strings.TrimSuffix`: here, the first parameter is the prefix, rather than being the last parameter. This is to allow piping one output to another:

```handlebars
{{ "   foo" | trimPrefix " " }}
foo
```

### `default`

If the value is set, return it, otherwise, a default value is used.

```handlebars
{{ "" | default "bar" }}
bar
```

### `required`

If the argument renders to an empty string, the application fails and exits with non-zero status code.

```handlebars
{{ "" | required }}
<!-- argument is marked as required, but it was not found in the YAML data -->
```

### `env`

Fetch an environment variable to be printed. If the environment variable is mandatory, consider using `required`. If the environment variable might be empty, consider using `default`.

`env` allows the key to be case-insensitive: it will be uppercased internally.

```handlebars
{{ env "user" }}
patrick
```

### `sha1sum`, `sha256sum`

Renders a `sha1sum` or `sha256sum` of a given value. The value is converted first to their YAML representation, with comments removed, then the `sum` is performed. This is to ensure that the "behavior" can stay the same, even when the file might have multiple comments that might change.

Primitives such as `string`, `bool` and `float64` are converted as-is.

While not recommended, you can use this to always generate a new name if the YAML declaration drifts. The following snippet uses `.`, which represents the entire YAML file -- on a multi-YAML file, each `.` represents a single file:

```handlebars
{{ . | sha1sum }}
f502bbf15d0988a9b28b73f8450de47f75179f5c
```

### `str`

Converts any primitive as stated in [String conversion](#string-conversion), to string:

```handlebars
{{ false | str }}
false
```

### `replace`

Converts the value to a string as stated in [String conversion](#string-conversion), then replaces all ocurrences of a string with another:

```handlebars
{{ "hello.dev" | replace "." "_" }}
hello_dev
```

### `alphanumify`, `alphanumdash`

Converts the value to a string as stated in [String conversion](#string-conversion), and keeps from the original string only alphanumeric characters -- for `alphanumify` -- or alphanumeric plus dashes and underscores -- like URLs, for `alphanumdash`:

```handlebars
{{ "secret-foo.dev" | alphanumify }}
secretsfoodev
```

```handlebars
{{ "secret-foo.dev" | alphanumdash }}
secrets-foodev
```

### `dottodash`, `dottounder`

Converts the value to a string as stated in [String conversion](#string-conversion), and replaces all dots to either dashes or underscores:

```handlebars
{{ "secret-foo.dev" | dottodash }}
secrets-foo-dev
```

```handlebars
{{ "secret-foo.dev" | dottounder }}
secrets-foo_dev
```

Particularly useful for Kubernetes FQDNs needed to be used as filenames.

## Example split for Tekton

Tekton Pipelines is a powerful tool that's available through a Helm Chart from the [cd.foundation](https://cd.foundation). We can grab it from their Helm repository and render it locally, then use `kubectl-split` to split it into multiple files.

We'll use the following filename template so there's one folder for each Kubernetes resource `kind`, so all `Secrets` for example are in the same folder, then we will use the resource name as defined in `metadata.name`. We'll also modify the name, since some of the Tekton resources have an FQDN for a name, like `tekton.pipelines.dev`, with the `dottodash` template function:

```handlebars
{{.kind|lower}}/{{.metadata.name|dottodash}}.yaml
```

We will render the Helm Chart locally to `stdout` with:

```bash
helm template tekton cdf/tekton-pipeline
```

Then we can pipe that output directly to `kubectl-split`:

```bash
helm template tekton cdf/tekton-pipeline | kubectl-split --template '{{.kind|lower}}/{{.metadata.name|dottodash}}.yaml'
```

Which will render the following output:

```text
Wrote rolebinding/tekton-pipelines-info.yaml -- 590 bytes.
Wrote service/tekton-pipelines-controller.yaml -- 1007 bytes.
Wrote podsecuritypolicy/tekton-pipelines.yaml -- 1262 bytes.
Wrote configmap/config-registry-cert.yaml -- 906 bytes.
Wrote configmap/feature-flags.yaml -- 646 bytes.
Wrote clusterrole/tekton-pipelines-controller-tenant-access.yaml -- 1035 bytes.
Wrote clusterrolebinding/tekton-pipelines-webhook-cluster-access.yaml -- 565 bytes.
Wrote role/tekton-pipelines-info.yaml -- 592 bytes.
Wrote service/tekton-pipelines-webhook.yaml -- 1182 bytes.
Wrote deployment/tekton-pipelines-webhook.yaml -- 3645 bytes.
Wrote serviceaccount/tekton-bot.yaml -- 883 bytes.
Wrote configmap/config-defaults.yaml -- 2424 bytes.
Wrote configmap/config-logging.yaml -- 1596 bytes.
Wrote customresourcedefinition/runs-tekton-dev.yaml -- 2308 bytes.
Wrote role/tekton-pipelines-leader-election.yaml -- 495 bytes.
Wrote rolebinding/tekton-pipelines-webhook.yaml -- 535 bytes.
Wrote customresourcedefinition/clustertasks-tekton-dev.yaml -- 2849 bytes.
Wrote customresourcedefinition/pipelineresources-tekton-dev.yaml -- 1874 bytes.
Wrote clusterrole/tekton-aggregate-view.yaml -- 1133 bytes.
Wrote role/tekton-pipelines-webhook.yaml -- 1152 bytes.
Wrote rolebinding/tekton-pipelines-webhook-leaderelection.yaml -- 573 bytes.
Wrote validatingwebhookconfiguration/validation-webhook-pipeline-tekton-dev.yaml -- 663 bytes.
Wrote serviceaccount/tekton-pipelines-webhook.yaml -- 317 bytes.
Wrote configmap/config-leader-election.yaml -- 985 bytes.
Wrote configmap/pipelines-info.yaml -- 1137 bytes.
Wrote clusterrolebinding/tekton-pipelines-controller-cluster-access.yaml -- 1163 bytes.
Wrote role/tekton-pipelines-controller.yaml -- 1488 bytes.
Wrote deployment/tekton-pipelines-controller.yaml -- 5203 bytes.
Wrote configmap/config-observability.yaml -- 2429 bytes.
Wrote customresourcedefinition/tasks-tekton-dev.yaml -- 2824 bytes.
Wrote mutatingwebhookconfiguration/webhook-pipeline-tekton-dev.yaml -- 628 bytes.
Wrote validatingwebhookconfiguration/config-webhook-pipeline-tekton-dev.yaml -- 742 bytes.
Wrote namespace/tekton-pipelines.yaml -- 808 bytes.
Wrote secret/webhook-certs.yaml -- 959 bytes.
Wrote customresourcedefinition/pipelineruns-tekton-dev.yaml -- 3801 bytes.
Wrote serviceaccount/tekton-pipelines-controller.yaml -- 908 bytes.
Wrote configmap/config-artifact-pvc.yaml -- 977 bytes.
Wrote customresourcedefinition/conditions-tekton-dev.yaml -- 1846 bytes.
Wrote clusterrolebinding/tekton-pipelines-controller-tenant-access.yaml -- 816 bytes.
Wrote rolebinding/tekton-pipelines-controller.yaml -- 1133 bytes.
Wrote rolebinding/tekton-pipelines-controller-leaderelection.yaml -- 585 bytes.
Wrote horizontalpodautoscaler/tekton-pipelines-webhook.yaml -- 1518 bytes.
Wrote configmap/config-artifact-bucket.yaml -- 1408 bytes.
Wrote customresourcedefinition/pipelines-tekton-dev.yaml -- 2840 bytes.
Wrote customresourcedefinition/taskruns-tekton-dev.yaml -- 3785 bytes.
Wrote clusterrole/tekton-aggregate-edit.yaml -- 1274 bytes.
Wrote clusterrole/tekton-pipelines-controller-cluster-access.yaml -- 1886 bytes.
Wrote clusterrole/tekton-pipelines-webhook-cluster-access.yaml -- 2480 bytes.
48 files generated.
```

We can navigate the folders:

```bash
$ tree -d
.
├── clusterrole
├── clusterrolebinding
├── configmap
├── customresourcedefinition
├── deployment
├── horizontalpodautoscaler
├── mutatingwebhookconfiguration
├── namespace
├── podsecuritypolicy
├── role
├── rolebinding
├── secret
├── service
├── serviceaccount
└── validatingwebhookconfiguration

15 directories
```

And poking into a single directory, for example, `configmap`:

```bash
$ tree configmap
configmap
├── config-artifact-bucket.yaml
├── config-artifact-pvc.yaml
├── config-defaults.yaml
├── config-leader-election.yaml
├── config-logging.yaml
├── config-observability.yaml
├── config-registry-cert.yaml
├── feature-flags.yaml
└── pipelines-info.yaml

0 directories, 9 files
```

## Contributing & Roadmap

Pull requests are welcomed! So far, looking for help with the following items, which are also part of the roadmap:

* Adding unit tests
* Improving the YAML file-by-file parser, right now it works by buffering line by line
* Adding support to install through `brew` or `krew`, for `kubectl`
* Functions to allow accessing labels and annotations on a way different than the dot-notation from Go templates
* Allowing only a subset of resources to be included or excluded, for example, extract only `Deployment` and `Service` type YAMLs (or extract all resources but not `Deployment` and `Service`)
