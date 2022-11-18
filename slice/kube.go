package slice

import "sort"

type yamlFile struct {
	filename string
	meta     kubeObjectMeta
	data     []byte
}

type kubeObjectMeta struct {
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
}

func newKubeObjectMeta(meta map[string]interface{}) *kubeObjectMeta {
	var k8smeta kubeObjectMeta

	if v, found := meta["apiVersion"]; found {
		if s, ok := v.(string); ok {
			k8smeta.APIVersion = s
		}
	}

	if v, found := meta["kind"]; found {
		if s, ok := v.(string); ok {
			k8smeta.Kind = s
		}
	}

	if v, found := meta["metadata"]; found {
		if m, ok := v.(map[string]interface{}); ok {
			if v, found := m["name"]; found {
				if s, ok := v.(string); ok {
					k8smeta.Name = s
				}
			}

			if v, found := m["namespace"]; found {
				if s, ok := v.(string); ok {
					k8smeta.Namespace = s
				}
			}
		}
	}

	return &k8smeta
}

func (k *kubeObjectMeta) findMissingField() string {
	if k.APIVersion == "" {
		return "apiVersion"
	}

	if k.Kind == "" {
		return "kind"
	}

	if k.Name == "" {
		return "metadata.name"
	}

	return ""
}

// from: https://github.com/helm/helm/blob/v3.7.1/pkg/releaseutil/kind_sorter.go#L31
var helmInstallOrder = []string{
	"Namespace",
	"NetworkPolicy",
	"ResourceQuota",
	"LimitRange",
	"PodSecurityPolicy",
	"PodDisruptionBudget",
	"ServiceAccount",
	"Secret",
	"SecretList",
	"ConfigMap",
	"StorageClass",
	"PersistentVolume",
	"PersistentVolumeClaim",
	"CustomResourceDefinition",
	"ClusterRole",
	"ClusterRoleList",
	"ClusterRoleBinding",
	"ClusterRoleBindingList",
	"Role",
	"RoleList",
	"RoleBinding",
	"RoleBindingList",
	"Service",
	"DaemonSet",
	"Pod",
	"ReplicationController",
	"ReplicaSet",
	"Deployment",
	"HorizontalPodAutoscaler",
	"StatefulSet",
	"Job",
	"CronJob",
	"Ingress",
	"APIService",
}

// from: https://github.com/helm/helm/blob/v3.7.1/pkg/releaseutil/kind_sorter.go#L111
func sortYAMLsByKind(manifests []yamlFile) []yamlFile {
	sort.SliceStable(manifests, func(i, j int) bool {
		return lessByKind(manifests[i], manifests[j], manifests[i].meta.Kind, manifests[j].meta.Kind, helmInstallOrder)
	})

	return manifests
}

// from: https://github.com/helm/helm/blob/v3.7.1/pkg/releaseutil/kind_sorter.go#L131
func lessByKind(a interface{}, b interface{}, kindA string, kindB string, o []string) bool {
	ordering := make(map[string]int, len(o))
	for v, k := range o {
		ordering[k] = v
	}

	first, aok := ordering[kindA]
	second, bok := ordering[kindB]

	if !aok && !bok {
		// if both are unknown then sort alphabetically by kind, keep original order if same kind
		if kindA != kindB {
			return kindA < kindB
		}
		return first < second
	}
	// unknown kind is last
	if !aok {
		return false
	}
	if !bok {
		return true
	}
	// sort different kinds, keep original order if same priority
	return first < second
}
