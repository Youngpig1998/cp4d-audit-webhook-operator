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
package services

import (
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Service is a wrapper around the corev1.Service object that meets the
// Reconcileable interface
type Service struct {
	*corev1.Service
}

// From returns a new Reconcileable Service from a corev1.Service
func From(service *corev1.Service) *Service {
	return &Service{Service: service}
}

// ShouldUpdate returns whether the resource should be updated in Kubernetes and
// the resource to update with
func (s Service) ShouldUpdate(current client.Object) (bool, client.Object) {
	newService := current.DeepCopyObject().(*corev1.Service)
	resources.MergeMetadata(newService, s)
	newService.Spec = s.Spec

	// Check the TargetPort as it can cause inequality issues. Please specify TargetPort even
	// though Kube will accept a Service without TargetPort. Kube will add TargetPort which
	// will get set to the value of Port, but the unspecified TargetPort will have value 0
	// Causing the equality check to fail. Handle unset TargetPort here just in case
	for i, port := range newService.Spec.Ports {
		if port.TargetPort == intstr.FromInt(0) {
			newService.Spec.Ports[i].TargetPort = intstr.FromInt(int(port.Port))
		}
	}

	// ClusterIP is immutable so keep current value
	currentService := current.DeepCopyObject().(*corev1.Service)
	newService.Spec.ClusterIP = currentService.Spec.ClusterIP

	return !equality.Semantic.DeepEqual(newService, current), newService
}

// GetResource retrieves the resource instance
func (s Service) GetResource() client.Object {
	return s.Service
}

// ResourceKind retrieves the string kind of the resource
func (s Service) ResourceKind() string {
	return "Service"
}

// ResourceIsNil returns whether or not the resource is nil
func (s Service) ResourceIsNil() bool {
	return s.Service == nil
}

// NewResourceInstance returns a new instance of the sme resource type
func (s Service) NewResourceInstance() client.Object {
	return &corev1.Service{}
}
