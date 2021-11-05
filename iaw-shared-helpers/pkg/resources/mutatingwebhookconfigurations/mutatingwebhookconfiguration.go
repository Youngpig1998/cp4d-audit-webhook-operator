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
package mutatingwebhookconfigurations

import (
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Certificate is a wrapper around the certmanager.Certificate object that meets the
// Reconcileable interface
type MutatingwebhookConfiguration struct {
	*admissionregistrationv1.MutatingWebhookConfiguration
}

// From returns a new Reconcileable Certificate from a certmanager.Certificate
func From(mutatingwebhookconfiguration *admissionregistrationv1.MutatingWebhookConfiguration) *MutatingwebhookConfiguration {
	return &MutatingwebhookConfiguration{
		MutatingWebhookConfiguration: mutatingwebhookconfiguration,
	}
}



// ShouldUpdate returns whether the resource should be updated in Kubernetes and
// the resource to update with
func (c MutatingwebhookConfiguration) ShouldUpdate(current client.Object) (bool, client.Object) {
	currentMutatingwebhookConfiguration := current.DeepCopyObject().(*admissionregistrationv1.MutatingWebhookConfiguration)
	newMutatingwebhookConfiguration := currentMutatingwebhookConfiguration.DeepCopy()
	resources.MergeMetadata(newMutatingwebhookConfiguration, c)
	newMutatingwebhookConfiguration.Webhooks = c.Webhooks
	return !equality.Semantic.DeepEqual(newMutatingwebhookConfiguration, currentMutatingwebhookConfiguration), newMutatingwebhookConfiguration
}

// GetResource retrieves the resource instance
func (c MutatingwebhookConfiguration) GetResource() client.Object {
	return c.MutatingWebhookConfiguration
}

// ResourceKind retrieves the string kind of the resource
func (c MutatingwebhookConfiguration) ResourceKind() string {
	return "MutatingWebhookConfiguration"
}

// ResourceIsNil returns whether or not the resource is nil
func (c MutatingwebhookConfiguration) ResourceIsNil() bool {
	return c.MutatingWebhookConfiguration == nil
}

// NewResourceInstance returns a new instance of the sme resource type
func (c MutatingwebhookConfiguration) NewResourceInstance() client.Object {
	return &admissionregistrationv1.MutatingWebhookConfiguration{}
}
