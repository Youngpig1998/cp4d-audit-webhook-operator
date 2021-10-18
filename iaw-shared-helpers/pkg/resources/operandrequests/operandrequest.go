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
package operandrequests

import (
	odlmv1alpha1 "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// OperandRequest is a wrapper around the odlmv1alpha1.OperandRequest object that meets the
// Reconcileable interface
type OperandRequest struct {
	*odlmv1alpha1.OperandRequest
}

// From returns a new Reconcileable OperandRequest from a odlmv1alpha1.OperandRequest
func From(operandrequest *odlmv1alpha1.OperandRequest) *OperandRequest {
	return &OperandRequest{OperandRequest: operandrequest}
}

// ShouldUpdate returns whether the resource should be updated in Kubernetes and
// the resource to update with
func (or OperandRequest) ShouldUpdate(current client.Object) (bool, client.Object) {
	currentOperandRequest := current.DeepCopyObject().(*odlmv1alpha1.OperandRequest)
	newOperandRequest := currentOperandRequest.DeepCopy()
	resources.MergeMetadata(newOperandRequest, or)
	newOperandRequest.Spec = or.Spec
	return !equality.Semantic.DeepEqual(newOperandRequest, currentOperandRequest), newOperandRequest
}

// GetResource retrieves the resource instance
func (or OperandRequest) GetResource() client.Object {
	return or.OperandRequest
}

// ResourceKind retrieves the string kind of the resource
func (or OperandRequest) ResourceKind() string {
	return "OperandRequest"
}

// ResourceIsNil returns whether or not the resource is nil
func (or OperandRequest) ResourceIsNil() bool {
	return or.OperandRequest == nil
}

// NewResourceInstance returns a new instance of the same resource type
func (or OperandRequest) NewResourceInstance() client.Object {
	return &odlmv1alpha1.OperandRequest{}
}
