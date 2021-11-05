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
package certificates

import (
	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Certificate is a wrapper around the certmanager.Certificate object that meets the
// Reconcileable interface
type Certificate struct {
	*certmanager.Certificate
}

// From returns a new Reconcileable Certificate from a certmanager.Certificate
func From(certificate *certmanager.Certificate) *Certificate {
	return &Certificate{Certificate: certificate}
}

// ShouldUpdate returns whether the resource should be updated in Kubernetes and
// the resource to update with
func (c Certificate) ShouldUpdate(current client.Object) (bool, client.Object) {
	currentCertificate := current.DeepCopyObject().(*certmanager.Certificate)
	newCertificate := currentCertificate.DeepCopy()
	resources.MergeMetadata(newCertificate, c)
	newCertificate.Spec = c.Spec
	return !equality.Semantic.DeepEqual(newCertificate, currentCertificate), newCertificate
}

// GetResource retrieves the resource instance
func (c Certificate) GetResource() client.Object {
	return c.Certificate
}

// ResourceKind retrieves the string kind of the resource
func (c Certificate) ResourceKind() string {
	return "Certificate"
}

// ResourceIsNil returns whether or not the resource is nil
func (c Certificate) ResourceIsNil() bool {
	return c.Certificate == nil
}

// NewResourceInstance returns a new instance of the sme resource type
func (c Certificate) NewResourceInstance() client.Object {
	return &certmanager.Certificate{}
}
