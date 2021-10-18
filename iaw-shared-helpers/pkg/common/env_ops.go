package common

import (
	"sort"

	corev1 "k8s.io/api/core/v1"
)

// EnvSort ensures consistent ordering of ENVVAR slices
func EnvSort(unsorted []corev1.EnvVar) []corev1.EnvVar {
	sorted := make(envVarlist, len(unsorted))
	i := 0
	for _, entry := range unsorted {
		sorted[i] = entry
		i++
	}
	sort.Sort(sort.Reverse(sorted))
	return sorted
}

type envVarlist []corev1.EnvVar

func (e envVarlist) Len() int           { return len(e) }
func (e envVarlist) Less(i, j int) bool { return e[i].Name > e[j].Name }
func (e envVarlist) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
