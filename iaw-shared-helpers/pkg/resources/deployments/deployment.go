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
package deployments

import (
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Deployment is a wrapper around the appsv1.Deployment object that meets the
// Reconcileable interface
type Deployment struct {
	*appsv1.Deployment
}

// From returns a new Reconcileable Deployment from a appsv1.Deployment
func From(deployment *appsv1.Deployment) *Deployment {
	return &Deployment{Deployment: deployment}
}

// ShouldUpdate returns whether the resource should be updated in Kubernetes and
// the resource to update with
func (d Deployment) ShouldUpdate(current client.Object) (bool, client.Object) {
	newDeployment := current.DeepCopyObject().(*appsv1.Deployment)
	resources.MergeMetadata(newDeployment, d)
	resources.MergeMetadata(&newDeployment.Spec.Template, &d.Spec.Template)
	mergedTemplate := newDeployment.Spec.Template
	newDeployment.Spec = d.Spec
	newDeployment.Spec.Template.ObjectMeta = mergedTemplate.ObjectMeta
	return !equality.Semantic.DeepEqual(newDeployment, current), newDeployment
}

// GetResource retrieves the resource instance
func (d Deployment) GetResource() client.Object {
	return d.Deployment
}

// ResourceKind retrieves the string kind of the resource
func (d Deployment) ResourceKind() string {
	return "Deployment"
}

// ResourceIsNil returns whether or not the resource is nil
func (d Deployment) ResourceIsNil() bool {
	return d.Deployment == nil
}

// NewResourceInstance returns a new instance of the same resource type
func (d Deployment) NewResourceInstance() client.Object {
	return &appsv1.Deployment{}
}
