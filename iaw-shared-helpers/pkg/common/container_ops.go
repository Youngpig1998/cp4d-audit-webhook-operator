package common

import (
	"fmt"
	"reflect"

	"github.com/imdario/mergo"
	corev1 "k8s.io/api/core/v1"
)

// GetContainerIndexByName returns the index of the given container name or error if not found
func GetContainerIndexByName(containers []corev1.Container, name string) (int, error) {
	for i := range containers {
		if containers[i].Name == name {
			return i, nil
		}
	}
	return -1, fmt.Errorf("Container named \"%s\" not found", name)
}

// ApplyContainerOverrides searches the list of overriding containers for a container matching the name of the original container.
// if one is found the values from the overriding container are merged into the original container, overriding values and performing
// unique merges on arrays
func ApplyContainerOverrides(orignal *corev1.Container, overridingContainers []corev1.Container) {
	idx, err := GetContainerIndexByName(overridingContainers, orignal.Name)
	if err == nil {
		override := overridingContainers[idx]
		mergo.Merge(orignal,
			override,
			mergo.WithTransformers(NamedTransformer{
				Types: []reflect.Type{
					reflect.TypeOf([]corev1.EnvVar{}),
					reflect.TypeOf([]corev1.EnvFromSource{}),
					reflect.TypeOf([]corev1.ContainerPort{}),
					reflect.TypeOf([]corev1.VolumeMount{}),
					reflect.TypeOf([]corev1.VolumeDevice{}),
				},
			}),
			mergo.WithOverride,
		)
	}
}
