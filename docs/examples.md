# Examples

- [Examples](#examples)
  - [Slicing the Tekton manifest](#slicing-the-tekton-manifest)

The following examples demonstrate the capabilities of `kubectl-slice`.

## Slicing the Tekton manifest

[Tekton Pipelines](https://tekton.dev/) is a powerful tool that's available through a Helm Chart from the [cd.foundation](https://cd.foundation). We can grab it from their Helm repository and render it locally, then use `kubectl-slice` to split it into multiple files.

We'll use the following filename template so there's one folder for each Kubernetes resource `kind`, so all `Secrets` for example are in the same folder, then we will use the resource name as defined in `metadata.name`. We'll also modify the name, since some of the Tekton resources have an FQDN for a name, like `tekton.pipelines.dev`, with the `dottodash` template function:

```handlebars
{{.kind|lower}}/{{.metadata.name|dottodash}}.yaml
```

We will render the Helm Chart locally to `stdout` with:

```bash
helm repo add cdf https://cdfoundation.github.io/tekton-helm-chart/
helm template tekton cdf/tekton-pipeline
```

Then we can pipe that output directly to `kubectl-slice`:

```bash
helm template tekton cdf/tekton-pipeline | kubectl-slice --template '{{.kind|lower}}/{{.metadata.name|dottodash}}.yaml'
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
