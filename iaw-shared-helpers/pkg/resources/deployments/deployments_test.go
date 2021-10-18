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
package deployments_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources/deployments"
)

var _ = Describe("Deployments", func() {
	Describe("ShouldUpdate", func() {
		It("Correctly indicates an update is required and performs the update", func() {
			deployment1 := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"label1": "value1",
					},
					Annotations: map[string]string{
						"annotation1": "value1",
					},
					Finalizers: []string{
						"finalizer1",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Paused: true,
				},
			}

			deployment2 := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"label2": "value2",
					},
					Annotations: map[string]string{
						"annotation2": "value2",
					},
					Finalizers: []string{
						"finalizer2",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Paused: false,
				},
			}

			update, newDeploymentObj := From(deployment2).ShouldUpdate(deployment1)
			newDeployment := newDeploymentObj.(*appsv1.Deployment)
			Expect(update).To(BeTrue())
			Expect(newDeployment.Spec.Paused).To(BeFalse())
			Expect(newDeployment.Labels).To(HaveLen(2))
			Expect(newDeployment.Annotations).To(HaveLen(2))
			Expect(newDeployment.Finalizers).To(HaveLen(2))
		})

		It("Correctly indicates no update is required when current tempalte metadata differs", func() {
			deployment1 := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"label1": "value1",
					},
					Annotations: map[string]string{
						"annotation1": "value1",
					},
					Finalizers: []string{
						"finalizer1",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"templatelabel1": "value1",
								"templatelabel2": "value2",
							},
							Annotations: map[string]string{
								"templateannotation1": "value1",
							},
							Finalizers: []string{
								"templatefinalizer1",
							},
						},
					},
					Paused: true,
				},
			}

			deployment2 := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"label1": "value1",
					},
					Annotations: map[string]string{
						"annotation1": "value1",
					},
					Finalizers: []string{
						"finalizer1",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"templatelabel1": "value1",
							},
							Annotations: map[string]string{
								"templateannotation1": "value1",
							},
							Finalizers: []string{
								"templatefinalizer1",
							},
						},
					},
					Paused: true,
				},
			}

			update, result := From(deployment2).ShouldUpdate(deployment1)
			newDeployment := result.DeepCopyObject().(*appsv1.Deployment)
			Expect(update).To(BeFalse())
			Expect(len(newDeployment.Spec.Template.Labels)).To(Equal(2))
			Expect(len(newDeployment.Spec.Template.Annotations)).To(Equal(1))
			Expect(len(newDeployment.Spec.Template.Finalizers)).To(Equal(1))
		})

		It("Correctly indicates no update is required", func() {
			deployment1 := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"label1": "value1",
					},
					Annotations: map[string]string{
						"annotation1": "value1",
					},
					Finalizers: []string{
						"finalizer1",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Paused: true,
				},
			}

			deployment2 := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"label1": "value1",
					},
					Annotations: map[string]string{
						"annotation1": "value1",
					},
					Finalizers: []string{
						"finalizer1",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Paused: true,
				},
			}

			update, _ := From(deployment2).ShouldUpdate(deployment1)
			Expect(update).To(BeFalse())
		})
	})

	Describe("GetResource", func() {
		It("Returns the correct resource", func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-deployment",
				},
				Spec: appsv1.DeploymentSpec{
					Paused: true,
				},
			}
			Expect(From(deployment).GetResource()).To(Equal(deployment))
		})
	})

	Describe("ResourceKind", func() {
		It("Returns the correct resource", func() {
			Expect(From(nil).ResourceKind()).To(Equal("Deployment"))
		})
	})

	Describe("ResourceIsNil", func() {
		It("Returns the correct resource", func() {
			Expect(From(nil).ResourceIsNil()).To(BeTrue())
			Expect(From(&appsv1.Deployment{}).ResourceIsNil()).To(BeFalse())
		})
	})

	Describe("NewResourceInstance", func() {
		It("Returns the correct resource", func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-deployment",
				},
				Spec: appsv1.DeploymentSpec{
					Paused: true,
				},
			}
			newInstance := From(deployment).NewResourceInstance()
			Expect(newInstance).NotTo(Equal(deployment))
		})
	})
})
