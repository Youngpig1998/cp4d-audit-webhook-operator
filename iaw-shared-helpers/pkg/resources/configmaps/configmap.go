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
package configmaps

import (
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConfigMap is a wrapper around the corev1.ConfigMap object that meets the
// Reconcileable interface
type ConfigMap struct {
	*corev1.ConfigMap
}

// From returns a new Reconcileable ConfigMap from a corev1.ConfigMap
func From(configMap *corev1.ConfigMap) *ConfigMap {
	return &ConfigMap{ConfigMap: configMap}
}

// ShouldUpdate returns whether the resource should be updated in Kubernetes and
// the resource to update with
func (cm ConfigMap) ShouldUpdate(currentObject client.Object) (bool, client.Object) {
	newConfigMap := currentObject.DeepCopyObject().(*corev1.ConfigMap)
	resources.MergeMetadata(newConfigMap, cm)
	newConfigMap.Data = cm.Data
	return !equality.Semantic.DeepEqual(newConfigMap, currentObject), newConfigMap
}

// GetResource retrieves the resource instance
func (cm ConfigMap) GetResource() client.Object {
	return cm.ConfigMap
}

// ResourceKind retrieves the string kind of the resource
func (cm ConfigMap) ResourceKind() string {
	return "ConfigMap"
}

// ResourceIsNil returns whether or not the resource is nil
func (cm ConfigMap) ResourceIsNil() bool {
	return cm.ConfigMap == nil
}

// NewResourceInstance returns a new instance of the same resource type
func (cm ConfigMap) NewResourceInstance() client.Object {
	return &corev1.ConfigMap{}
}
