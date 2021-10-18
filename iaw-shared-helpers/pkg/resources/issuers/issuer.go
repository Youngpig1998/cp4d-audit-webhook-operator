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
package issuers

import (
	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Issuer is a wrapper around the certmanager.Issuer object that meets the
// Reconcileable interface
type Issuer struct {
	*certmanager.Issuer
}

// From returns a new Reconcileable Issuer from a certmanager.Issuer
func From(issuer *certmanager.Issuer) *Issuer {
	return &Issuer{Issuer: issuer}
}

// ShouldUpdate returns whether the resource should be updated in Kubernetes and
// the resource to update with
func (i Issuer) ShouldUpdate(current client.Object) (bool, client.Object) {
	currentIssuer := current.DeepCopyObject().(*certmanager.Issuer)
	newIssuer := currentIssuer.DeepCopy()
	resources.MergeMetadata(newIssuer, i)
	newIssuer.Spec = i.Spec
	return !equality.Semantic.DeepEqual(newIssuer, currentIssuer), newIssuer
}

// GetResource retrieves the resource instance
func (i Issuer) GetResource() client.Object {
	return i.Issuer
}

// ResourceKind retrieves the string kind of the resource
func (i Issuer) ResourceKind() string {
	return "Issuer"
}

// ResourceIsNil returns whether or not the resource is nil
func (i Issuer) ResourceIsNil() bool {
	return i.Issuer == nil
}

// NewResourceInstance returns a new instance of the sme resource type
func (i Issuer) NewResourceInstance() client.Object {
	return &certmanager.Issuer{}
}
