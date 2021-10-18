/*
Copyright 2021 Fan Zhang.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AuditWebhookSpec defines the desired state of AuditWebhook
type AuditWebhookSpec struct {
	// The mirror image corresponding to the business service, including the name: tag
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`
	// The mirror image corresponding to the business service, including the dockerregistryprefix
	DockerRegistryPrefix string `json:"dockerRegistryPrefix"`
}

// AuditWebhookStatus defines the observed state of AuditWebhook
type AuditWebhookStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Nodes []string `json:"nodes"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AuditWebhook is the Schema for the auditwebhooks API
type AuditWebhook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AuditWebhookSpec   `json:"spec,omitempty"`
	Status AuditWebhookStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AuditWebhookList contains a list of AuditWebhook
type AuditWebhookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AuditWebhook `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AuditWebhook{}, &AuditWebhookList{})
}
