package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLessByKind(t *testing.T) {
	tests := []struct {
		name  string
		kindA string
		kindB string
		order []string
		want  bool
	}{
		{
			name:  "both kinds in order, kindA before kindB",
			kindA: "Namespace",
			kindB: "Service",
			order: HelmInstallOrder,
			want:  true, // Namespace comes before Service
		},
		{
			name:  "both kinds in order, kindB before kindA",
			kindA: "Service",
			kindB: "Namespace",
			order: HelmInstallOrder,
			want:  false, // Service comes after Namespace
		},
		{
			name:  "same kinds",
			kindA: "Service",
			kindB: "Service",
			order: HelmInstallOrder,
			want:  false, // Same kinds should maintain order
		},
		{
			name:  "kindA in order, kindB not in order",
			kindA: "Namespace",
			kindB: "UnknownKind",
			order: HelmInstallOrder,
			want:  true, // Known kinds come before unknown
		},
		{
			name:  "kindA not in order, kindB in order",
			kindA: "UnknownKind",
			kindB: "Namespace",
			order: HelmInstallOrder,
			want:  false, // Unknown kinds come after known
		},
		{
			name:  "neither kind in order, alphabetical first",
			kindA: "AAA",
			kindB: "ZZZ",
			order: HelmInstallOrder,
			want:  true, // Alphabetical ordering for unknown kinds
		},
		{
			name:  "neither kind in order, alphabetical second",
			kindA: "ZZZ",
			kindB: "AAA",
			order: HelmInstallOrder,
			want:  false, // Alphabetical ordering for unknown kinds
		},
		{
			name:  "both kinds unknown and equal",
			kindA: "Unknown",
			kindB: "Unknown",
			order: HelmInstallOrder,
			want:  false, // Same kinds should maintain order
		},
		{
			name:  "empty order slice",
			kindA: "Service",
			kindB: "Namespace",
			order: []string{},
			want:  false, // Alphabetical ordering when no order specified (N before S)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lessByKind(tt.kindA, tt.kindB, tt.order)
			require.Equal(t, tt.want, result)
		})
	}
}

func TestSortByKind(t *testing.T) {
	tests := []struct {
		name      string
		manifests []YAMLFile
		want      []string // Expected order of kinds after sorting
	}{
		{
			name: "already sorted manifests",
			manifests: []YAMLFile{
				{
					Filename: "namespace.yaml",
					Meta: &ObjectMeta{
						Kind: "Namespace",
					},
				},
				{
					Filename: "service.yaml",
					Meta: &ObjectMeta{
						Kind: "Service",
					},
				},
				{
					Filename: "deployment.yaml",
					Meta: &ObjectMeta{
						Kind: "Deployment",
					},
				},
			},
			want: []string{"Namespace", "Service", "Deployment"},
		},
		{
			name: "unsorted manifests",
			manifests: []YAMLFile{
				{
					Filename: "deployment.yaml",
					Meta: &ObjectMeta{
						Kind: "Deployment",
					},
				},
				{
					Filename: "namespace.yaml",
					Meta: &ObjectMeta{
						Kind: "Namespace",
					},
				},
				{
					Filename: "service.yaml",
					Meta: &ObjectMeta{
						Kind: "Service",
					},
				},
			},
			want: []string{"Namespace", "Service", "Deployment"},
		},
		{
			name: "with unknown kinds",
			manifests: []YAMLFile{
				{
					Filename: "deployment.yaml",
					Meta: &ObjectMeta{
						Kind: "Deployment",
					},
				},
				{
					Filename: "unknown.yaml",
					Meta: &ObjectMeta{
						Kind: "UnknownKind",
					},
				},
				{
					Filename: "namespace.yaml",
					Meta: &ObjectMeta{
						Kind: "Namespace",
					},
				},
			},
			want: []string{"Namespace", "Deployment", "UnknownKind"},
		},
		{
			name: "multiple of same kind",
			manifests: []YAMLFile{
				{
					Filename: "deployment1.yaml",
					Meta: &ObjectMeta{
						Kind: "Deployment",
					},
				},
				{
					Filename: "deployment2.yaml",
					Meta: &ObjectMeta{
						Kind: "Deployment",
					},
				},
				{
					Filename: "namespace.yaml",
					Meta: &ObjectMeta{
						Kind: "Namespace",
					},
				},
			},
			want: []string{"Namespace", "Deployment", "Deployment"},
		},
		{
			name: "all unknown kinds - alphabetical",
			manifests: []YAMLFile{
				{
					Filename: "c.yaml",
					Meta: &ObjectMeta{
						Kind: "C",
					},
				},
				{
					Filename: "a.yaml",
					Meta: &ObjectMeta{
						Kind: "A",
					},
				},
				{
					Filename: "b.yaml",
					Meta: &ObjectMeta{
						Kind: "B",
					},
				},
			},
			want: []string{"A", "B", "C"},
		},
		{
			name: "empty manifests",
			manifests: []YAMLFile{},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorted := SortByKind(tt.manifests)
			
			// Check if the length matches
			require.Equal(t, len(tt.want), len(sorted))
			
			// Verify the order of kinds
			if len(sorted) > 0 {
				kinds := make([]string, len(sorted))
				for i, manifest := range sorted {
					kinds[i] = manifest.Meta.Kind
				}
				require.Equal(t, tt.want, kinds)
			}
		})
	}
}

// TestHelmInstallOrder verifies that the predefined HelmInstallOrder slice is correctly defined
func TestHelmInstallOrder(t *testing.T) {
	// Verify some key ordering relationships from the Helm install order
	require.Contains(t, HelmInstallOrder, "Namespace")
	require.Contains(t, HelmInstallOrder, "Deployment")
	require.Contains(t, HelmInstallOrder, "Service")
	
	// Namespace should come before Service
	namespaceIndex := -1
	serviceIndex := -1
	
	for i, kind := range HelmInstallOrder {
		if kind == "Namespace" {
			namespaceIndex = i
		}
		if kind == "Service" {
			serviceIndex = i
		}
	}
	
	require.NotEqual(t, -1, namespaceIndex)
	require.NotEqual(t, -1, serviceIndex)
	require.Less(t, namespaceIndex, serviceIndex)
}
