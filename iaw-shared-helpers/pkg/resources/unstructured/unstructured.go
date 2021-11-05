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
package unstructured

import (
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources"
	"k8s.io/apimachinery/pkg/api/equality"
	unstructuredv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Unstructured is a wrapper around the unstructured.Unstructured object that meets the
// Reconcileable interface.
// ! You must have set the GroupVersionKind !
// ! Only works for resources with a spec field !
type Unstructured struct {
	*unstructuredv1.Unstructured
}

// From returns a new Reconcileable Unstructured from a unstructured.Unstructured
func From(unstructured *unstructuredv1.Unstructured) *Unstructured {
	return &Unstructured{Unstructured: unstructured}
}

// ShouldUpdate returns whether the resource should be updated in Kubernetes and
// the resource to update with
func (u Unstructured) ShouldUpdate(current client.Object) (bool, client.Object) {
	newUnstructured := current.DeepCopyObject().(*unstructuredv1.Unstructured)
	resources.MergeMetadata(newUnstructured, u)
	newUnstructured.Object["spec"] = u.Object["spec"]
	return !equality.Semantic.DeepEqual(newUnstructured, current), newUnstructured
}

// GetResource retrieves the resource instance
func (u Unstructured) GetResource() client.Object {
	return u.Unstructured
}

// ResourceKind retrieves the string kind of the resource
func (u Unstructured) ResourceKind() string {
	return u.GetKind()
}

// ResourceIsNil returns whether or not the resource is nil
func (u Unstructured) ResourceIsNil() bool {
	_, hasSpec := u.Unstructured.Object["spec"]
	return !hasSpec
}

// NewResourceInstance returns a new instance of the sme resource type
func (u Unstructured) NewResourceInstance() client.Object {
	newUnstructured := &unstructuredv1.Unstructured{}
	newUnstructured.SetGroupVersionKind(u.GroupVersionKind())
	return newUnstructured
}
