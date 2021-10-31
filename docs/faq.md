# Frequently Asked Questions

- [Frequently Asked Questions](#frequently-asked-questions)
  - [I want to exclude or include certain Kubernetes resource types, how do I do it?](#i-want-to-exclude-or-include-certain-kubernetes-resource-types-how-do-i-do-it)
  - [Some of the code in my YAML is an entire YAML file commented out, how do I skip it?](#some-of-the-code-in-my-yaml-is-an-entire-yaml-file-commented-out-how-do-i-skip-it)
  - [How to add namespaces to YAML resources with no namespace?](#how-to-add-namespaces-to-yaml-resources-with-no-namespace)
  - [How do I access YAML fields by name?](#how-do-i-access-yaml-fields-by-name)
  - [Two files will generate the same file name, what do I do?](#two-files-will-generate-the-same-file-name-what-do-i-do)
  - [This app doesn't seem to work with Windows `CRLF`](#this-app-doesnt-seem-to-work-with-windows-crlf)
  - [How are string conversions handled?](#how-are-string-conversions-handled)
  - [I keep getting `file name template parse failed: bad character`, how do I fix it?](#i-keep-getting-file-name-template-parse-failed-bad-character-how-do-i-fix-it)

## I want to exclude or include certain Kubernetes resource types, how do I do it?

`kubectl-slice` has two available flags: `--exclude-kind` and `--include-kind`. They can be used to exclude or include specific resources. For example, to exclude all `Deployment` resources, you can use `--exclude-kind=Deployment`:

```bash
kubectl-slice -f manifest.yaml --exclude-kind=Deployment
```

Both arguments can be used comma-separated or by calling them multiple times. As such, these two invocations are the same:

```bash
kubectl-slice --exclude-kind=Deployment,ReplicaSet,DaemonSet
kubectl-slice --exclude-kind=Deployment --exclude-kind=ReplicaSet --exclude-kind=DaemonSet
```

## Some of the code in my YAML is an entire YAML file commented out, how do I skip it?

By default, `kubectl-slice` will also slice out commented YAML file sections. If you would rather want to ensure only Kubernetes resources are sliced from the original YAML file, then there's two options:

* Use `--include-kind` to only include Kubernetes resources by kind; or
* Use `--skip-non-k8s` to skip any non-Kubernetes resources

`--include-kind` can be used so you control your entire output by specifying only the resources you want. For example, if you want to only slice out `Deployment` resources, you can use `--include-kind=Deployment`.

`--skip-non-k8s`, on the other hand, works by ensuring that your YAML contains the following fields: `apiVersion`, `kind`, and `metadata.name`. Keep in mind that it won't check if these fields are empty, it will just ensure that those fields exist within each one of the YAML processed. If one of them does not contain these fields, it will be skipped.

## How to add namespaces to YAML resources with no namespace?

It's very common that Helm charts or even plain YAMLs found online might not contain the namespace, and because of that, the field isn't available in the YAML. Since this tool was created to fit a specific criteria [as seen above](README.md#why-kubectl-slice), there's no need to implement this here. However, you can use `kustomize` to quickly add the namespace to your manifest, then run it through `kubectl-slice`.

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

This will render your new file, namespaces included, to `stdout`. You can pipe this as-is to `kubectl-slice`:

```
kustomize build | kubectl-slice
```

Keep in mind that is recommended to **not add namespaces** to your YAML resources, to allow users to install in any destination they choose. For example, a namespaceless file called `foo.yaml` can be installed to the namespace `bar` by using:

```
kubectl apply -n bar -f foo.yaml
```

## How do I access YAML fields by name?

Any field from the YAML file can be used, however, non-existent fields will render an empty string. This is very common for situations such as rendering a Helm template [where the namespace shouldn't be defined](#how-to-add-namespaces-to-yaml-resources-with-no-namespace).

If you would rather fail executing `kubectl-slice` if a field was not found, consider using the `required` Go Template function. The following template will make `kubectl-slice` fail with a non-zero exit code if the namespace of any of the resources is not defined.

```handlebars
{{.metadata.namespace | required}}.yaml
```

If you would rather provide a default for those resources without, say, a namespace, you can use the `default` function:

```handlebars
{{.metadata.namespace | default "global"}}.yaml
```

This will render any resource without a namespace with the name `global.yaml`.

## Two files will generate the same file name, what do I do?

Since it's possible to provide a Go Template for a file name that might be the same for multiple resources, `kubectl-slice` will append any YAML that matches by file name to the given file using the `---` separator.

For example, considering the following file name:

```handlebars
{{.metadata.namespace | default "global"}}.yaml
```

Any cluster-scoped resource will be appended into `global.yaml`, while any resource in the namespace `production` will be appended to `production.yaml`.

## This app doesn't seem to work with Windows `CRLF`

`kubectl-slice` does not support Windows line breaks with `CRLF` -- also known as `\r\n`. Consider using Unix line breaks `\n`.

## How are string conversions handled?

Since `kubectl-slice` is built in Go, there's only a handful of primitives that can be read from the YAML manifest. All of these have been hand-picked to be stringified automatically -- in fact, multiple template functions will accept them, and yes, that means you can `lowercase` a number 😅

The decision is intentional and it's due to the fact that's impossible to map any potential Kubernetes resource, given the fact that you can teach Kubernetes new objects using Custom Resource Definitions. Because of that, resources are read as untyped and converted to strings when possible.

The following YAML untyped values are handled, [in accordance with the `json.Unmarshal` documentation](https://pkg.go.dev/encoding/json#Unmarshal):

* `bool`, for JSON booleans
* `float64`, for JSON numbers
* `string`, for JSON strings

## I keep getting `file name template parse failed: bad character`, how do I fix it?

If you're receiving this error, chances are you're attempting to access a field from the YAML whose name is not limited to alphanumeric characters, such as annotations or labels, like `app.kubernetes.io/name`.

To fix it, use the [`index` function](docs/functions.md#index) to access the field by index. For example, if you want to access the `app.kubernetes.io/name` field, you can use the following template:

```handlebars
{{ index "app.kubernetes.io/name" .metadata.labels }}
```
