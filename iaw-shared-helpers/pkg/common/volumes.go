// ------------------------------------------------------ {COPYRIGHT-TOP} ---
// IBM Confidential
// OCO Source Materials
// 5900-AEO
//
// Copyright IBM Corp. 2021
//
// The source code for this program is not published or otherwise
// divested of its trade secrets, irrespective of what has been
// deposited with the U.S. Copyright Office.
// ------------------------------------------------------ {COPYRIGHT-END} ---
package common

import (
	"sort"

	corev1 "k8s.io/api/core/v1"
)

// VolumesUniqueMerge merges two slices of Volume structs
// duplicates overridden in favour of the initial slice over the additions
func VolumesUniqueMerge(initial []corev1.Volume, additions ...corev1.Volume) []corev1.Volume {
	result := initial
	for i := 0; i < len(additions); i++ {
		exists := false
		for j := 0; j < len(initial); j++ {
			if additions[i].Name == initial[j].Name {
				exists = true
				break
			}
		}
		if !exists {
			result = append(result, additions[i])
		}
	}
	return result
}

// VolumeMountsUniqueMerge merges two slices of VolumeMount structs
// duplicates overridden in favour of the initial slice over the additions
func VolumeMountsUniqueMerge(initial []corev1.VolumeMount, additions ...corev1.VolumeMount) []corev1.VolumeMount {
	result := initial
	for i := 0; i < len(additions); i++ {
		exists := false
		for j := 0; j < len(initial); j++ {
			if additions[i].Name == initial[j].Name {
				exists = true
				break
			}
		}
		if !exists {
			result = append(result, additions[i])
		}
	}
	return result
}

// VolumeSort ensures consistent ordering of Volume slices
func VolumeSort(unsorted []corev1.Volume) []corev1.Volume {
	sorted := make(volumelist, len(unsorted))
	i := 0
	for _, entry := range unsorted {
		sorted[i] = entry
		i++
	}
	sort.Sort(sort.Reverse(sorted))
	return sorted
}

type volumelist []corev1.Volume

func (v volumelist) Len() int           { return len(v) }
func (v volumelist) Less(i, j int) bool { return v[i].Name > v[j].Name }
func (v volumelist) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
