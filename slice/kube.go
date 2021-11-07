package slice

import "sort"

type yamlFile struct {
	name string
	kind string
	data []byte
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

func sortYAMLsByKind(manifests []yamlFile) []yamlFile {
	sort.SliceStable(manifests, func(i, j int) bool {
		return lessByKind(manifests[i], manifests[j], manifests[i].kind, manifests[j].kind, helmInstallOrder)
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
