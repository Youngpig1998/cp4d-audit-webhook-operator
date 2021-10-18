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
package secrets

import (
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Secret is a wrapper around the corev1.Secret object that meets the
// Reconcileable interface
type Secret struct {
	*corev1.Secret
}

// From returns a new Reconcileable Secret from a corev1.Secret
func From(secret *corev1.Secret) *Secret {
	return &Secret{Secret: secret}
}

// ShouldUpdate returns whether the resource should be updated in Kubernetes and
// the resource to update with
func (s Secret) ShouldUpdate(current client.Object) (bool, client.Object) {
	currentSecret := current.DeepCopyObject().(*corev1.Secret)
	newSecret := currentSecret.DeepCopy()
	resources.MergeMetadata(newSecret, s)
	newSecret.Data = s.Data
	newSecret.StringData = s.StringData
	return !equality.Semantic.DeepEqual(newSecret, currentSecret), newSecret
}

// GetResource retrieves the resource instance
func (s Secret) GetResource() client.Object {
	return s.Secret
}

// ResourceKind retrieves the string kind of the resource
func (s Secret) ResourceKind() string {
	return "Secret"
}

// ResourceIsNil returns whether or not the resource is nil
func (s Secret) ResourceIsNil() bool {
	return s.Secret == nil
}

// NewResourceInstance returns a new instance of the same resource type
func (s Secret) NewResourceInstance() client.Object {
	return &corev1.Secret{}
}
