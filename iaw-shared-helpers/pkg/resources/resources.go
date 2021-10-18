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
package resources

import "github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/common"

// MetadataUpdatableResource is any resource that can have its labels, annotations and finlizers
// retrieved and set
type MetadataUpdatableResource interface {
	SetLabels(labels map[string]string)
	SetAnnotations(annotations map[string]string)
	SetFinalizers(finalizers []string)
	GetLabels() map[string]string
	GetAnnotations() map[string]string
	GetFinalizers() []string
}

// MergeMetadata Merges the metadata of the desired resource into the metadata of the new resource by updating the new resource
func MergeMetadata(newResource MetadataUpdatableResource, desiredResource MetadataUpdatableResource) {
	newLabels := common.CombineStringStringMaps(newResource.GetLabels(), desiredResource.GetLabels())
	// If checks stop probelms where labels: map{} is not equal to labels not set at all
	if len(newLabels) != 0 {
		newResource.SetLabels(newLabels)
	}
	newAnnotations := common.CombineStringStringMaps(newResource.GetAnnotations(), desiredResource.GetAnnotations())
	if len(newAnnotations) != 0 {
		newResource.SetAnnotations(newAnnotations)
	}
	newFinalizers := common.CombineStringSlices(newResource.GetFinalizers(), desiredResource.GetFinalizers())
	if len(newFinalizers) != 0 {
		newResource.SetFinalizers(newFinalizers)
	}
}
