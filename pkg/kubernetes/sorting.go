package kubernetes

import (
	"sort"
)

// YAMLFile represents a Kubernetes YAML file with associated metadata and content.
// It's used throughout the application for storing and processing YAML resources.
type YAMLFile struct {
	Filename string
	Meta     *ObjectMeta
	Data     []byte
}

// HelmInstallOrder defines the order in which Kubernetes resources should be installed
// from: https://github.com/helm/helm/blob/v3.11.1/pkg/releaseutil/kind_sorter.go#LL31-L67C2
var HelmInstallOrder = []string{
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
	"IngressClass",
	"Ingress",
	"APIService",
}

// SortByKind sorts a slice of YAMLFile according to Kubernetes resource kind ordering
func SortByKind(manifests []YAMLFile) []YAMLFile {
	sort.SliceStable(manifests, func(i, j int) bool {
		return lessByKind(
			manifests[i].Meta.Kind,
			manifests[j].Meta.Kind,
			HelmInstallOrder,
		)
	})

	return manifests
}

// lessByKind compares two kinds and determines their relative order
// from: https://github.com/helm/helm/blob/v3.11.1/pkg/releaseutil/kind_sorter.go#L133-L158
func lessByKind(kindA, kindB string, order []string) bool {
	ordering := make(map[string]int, len(order))
	for v, k := range order {
		ordering[k] = v
	}

	first, aok := ordering[kindA]
	second, bok := ordering[kindB]

	if !aok && !bok {
		// if both are unknown then sort alphabetically by kind
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
