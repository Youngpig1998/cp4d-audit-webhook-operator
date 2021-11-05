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
package networkpolicies

import (
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NetworkPolicy is a wrapper around the networking.NetworkPolicy object that meets the
// Reconcileable interface
type NetworkPolicy struct {
	*networking.NetworkPolicy
}

// DefaultProtocol The default protocol set by Kube
var DefaultProtocol corev1.Protocol = corev1.ProtocolTCP

// From returns a new Reconcileable NetworkPolicy from a networking.NetworkPolicy
func From(networkPolicy *networking.NetworkPolicy) *NetworkPolicy {
	return &NetworkPolicy{NetworkPolicy: networkPolicy}
}

// ShouldUpdate returns whether the resource should be updated in Kubernetes and
// the resource to update with. Ensure you have set protocol
func (netpol NetworkPolicy) ShouldUpdate(currentObject client.Object) (bool, client.Object) {
	currentNetworkPolicy := currentObject.DeepCopyObject().(*networking.NetworkPolicy)
	newNetworkPolicy := currentNetworkPolicy.DeepCopy()
	resources.MergeMetadata(newNetworkPolicy, netpol)
	newNetworkPolicy.Spec = netpol.Spec
	return !equality.Semantic.DeepEqual(newNetworkPolicy, currentNetworkPolicy), newNetworkPolicy
}

// GetResource retrieves the resource instance
func (netpol NetworkPolicy) GetResource() client.Object {
	return netpol.NetworkPolicy
}

// ResourceKind retrieves the string kind of the resource
func (netpol NetworkPolicy) ResourceKind() string {
	return "NetworkPolicy"
}

// ResourceIsNil returns whether or not the resource is nil
func (netpol NetworkPolicy) ResourceIsNil() bool {
	return netpol.NetworkPolicy == nil
}

// NewResourceInstance returns a new instance of the same resource type
func (netpol NetworkPolicy) NewResourceInstance() client.Object {
	return &networking.NetworkPolicy{}
}
