# Including and excluding items

- [Including and excluding items](#including-and-excluding-items)
  - [Including items](#including-items)
  - [Excluding items](#excluding-items)
  - [Filtering by labels](#filtering-by-labels)
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

## Filtering by labels

The `--selector` flag (or `-l`) keeps only the resources whose `metadata.labels` match a Kubernetes label selector. It uses the same syntax and rules as `kubectl get -l`, so all of these work:

```bash
kubectl-slice -f foo.yaml --stdout -l app=web
kubectl-slice -f foo.yaml --stdout -l 'app=web,env in (prod,staging)'
kubectl-slice -f foo.yaml --stdout -l 'env notin (dev),!canary'
```

The selector can be combined with any of the include and exclude flags above. A resource has to pass all of them to be kept.

Selectors like `!canary` or `env!=prod` match resources that have no labels at all, the same way they do in `kubectl`. That includes YAML files that aren't Kubernetes manifests, so pair them with `--skip-non-k8s` if your input has any.

Kubernetes only accepts text as a label value. When a value isn't written as text, `kubectl-slice` handles it like this:

| Label value | Without `--skip-non-k8s` | With `--skip-non-k8s` |
| --- | --- | --- |
| Text, like `app: web` or `version: "1.10"` | Matched as written | Matched as written |
| A number or boolean, like `version: 1.10` or `enabled: true` | Matched against the text as written in the file, so `version: 1.10` matches `version=1.10` | The resource is skipped |
| Empty, like `version:` or `version: ~` | Matched as an empty value | Matched as an empty value |
| A list or a map | Error | The resource is skipped |

Labels are only read when `--selector` is used. Without it, `--skip-non-k8s` does not look at labels.

## Excluding non Kubernetes manifests

In some cases, you might provide to `kubectl-slice` a list of YAML files that might not actually be Kubernetes manifests. The flag `--skip-non-k8s` can be used to skip these files that do not have an `apiVersion`, `kind` and `metadata.name`.

Be aware there are no attempts to validate whether these fields are correct. If you're expecting to exclude a Kubernetes manifest with a nonexistent API version with this, it won't work. This flag is only meant to skip files that do not have the fields mentioned before, and it will perform no API calls to your Kubernetes cluster to check if the `apiVersion` and `kind` fields are valid objects or CRDs in your Kubernetes cluster.
