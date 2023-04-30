# Including and excluding items

- [Including and excluding items](#including-and-excluding-items)
  - [Including items](#including-items)
  - [Excluding items](#excluding-items)
  - [Excluding non Kubernetes manifests](#excluding-non-kubernetes-manifests)

`kubectl-slice` supports including and excluding items from the list of resources to be processed. You can achieve this by using the `--include` and `--exclude` flags and their extensions.

## Including items

The following flags will allow you to include items by name or kind or both:

* `--include-name`: include items by name. For example, on a pod named `foo`, you can use `--include-name=foo` to include it.
* `--include-kind`: include items by kind. For example, on a pod, you can use `--include-kind=Pod` to include it.
* `--include`: include items by kind and name, using the format `<kind>/<name>`. For example, on a pod named `foo`, you can use `--include=Pod/foo` to include it.

Globs are supported on all of the above flags so that you can use `--include-name=foo*` to include all resources with names starting with `foo`. For the `--include` flag, globs are supported on both the `<kind>` and `<name>` parts so that you can use `--include=Pod/foo*` to include all pods with names starting with `foo`.

## Excluding items

The following flags will allow you to exclude items by name or kind, or both:

* `--exclude-name`: exclude items by name. For example, on a pod named `foo`, you can use `--exclude-name=foo` to exclude it.
* `--exclude-kind`: exclude items by kind. For example, on a pod, you can use `--exclude-kind=Pod` to exclude it.
* `--exclude`: exclude items by kind and name, using the format `<kind>/<name>`. For example, on a pod named `foo`, you can use `--exclude=Pod/foo` to exclude it.

## Excluding non Kubernetes manifests

In some cases, you might provide to `kubectl-slice` a list of YAML files that might not actually be Kubernetes manifests. The flag `--skip-non-k8s` can be used to skip these files that do not have an `apiVersion`, `kind` and `metadata.name`.

Be aware there are no attempts to validate whether these fields are correct. If you're expecting to exclude a Kubernetes manifest with a nonexistent API version with this, it won't work. This flag is only meant to skip files that do not have the fields mentioned before, and it will perform no API calls to your Kubernetes cluster to check if the `apiVersion` and `kind` fields are valid objects or CRDs in your Kubernetes cluster.
