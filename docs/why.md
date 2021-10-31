# Why `kubectl-slice`?

- [Why `kubectl-slice`?](#why-kubectl-slice)
    - [Differences with other tools](#differences-with-other-tools)
      - [Losing the original file and its format](#losing-the-original-file-and-its-format)
      - [Naming format and access to data within YAML](#naming-format-and-access-to-data-within-yaml)

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

Performing this task on big installations such as applications coming from Helm is a bit daunting, and a manual task. `kubectl-slice` can help by allowing you to read a single YAML file which holds multiple YAML manifests, parse each one of them, allow you to use their fields as parameters to generate custom names, then rendering those into individual files in a specific folder.

### Differences with other tools

#### Losing the original file and its format

There are other plugins and apps out there that can split your YAML into multiple sub-YAML files like `kubectl-slice`, however, they do it by decoding the YAML, processing it, then re-encode it again, which will lose its original definition. That means that some array pieces, for example, might be encoded to a different output -- while still keeping them as arrays; comments are also lost -- since the decoding to Go, then re-encoding back to YAML will ignore YAML Comments.

`kubectl-slice` will keep the original file, and even when it will still parse it into Go to give you the ability to use any of the fields as part of the template for the name, the original file contents are still preserved with no changes, so your comments and the preference on how you render arrays, for example, will remain exactly the same as the original file.

#### Naming format and access to data within YAML

One of the things you can do too with `kubectl-slice` that you might not be able to do with other tools is the fact that with `kubectl-slice` you can literally access any field from the YAML file. Now, granted, if for example you decide to use an annotation in your YAML as part of the name template, that annotation may exist in _some_ of the YAMLs but perhaps not in all of them, so you have to account for that by providing a [`default`](docs/template_functions.md#default) or using Go Template's `if else` blocks.

Other apps might not allow you to read into the entire YAML, and even more so, they might enforce a convention on some of the fields you are able to access. Resource names, for example, [should follow a Kubernetes standard](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names) which some apps might edit preemptively since they don't make for good or "nice" file names, and as such, replace all dots for underscores. `kubectl-slice` will let you provide a template that might render an invalid file name, that's true, but you have [a plethora of functions](docs/template_functions.md#replace) to modify its behavior yourself to something that fits your design better. Perhaps you prefer dashes rather than underscores, and you can do that.

Upcoming versions will improve this even more by allowing annotation access using positions rather than names, for example.
