// ------------------------------------------------------ {COPYRIGHT-TOP} ---
// IBM Confidential
// Automated Tests
// Copyright IBM Corp. 2021
// ------------------------------------------------------ {COPYRIGHT-END} ---
package testing

import (
	"context"
	"fmt"
	"reflect"

	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateNamespaceName is an optional helper method to return a NamespacedName
func CreateNamespaceName(name string, namespace string) types.NamespacedName {
	return types.NamespacedName{Name: name, Namespace: namespace}
}

// CreateObject creates the kube object and then retrieves it to ensure it was created
func CreateObject(client client.Client, obj client.Object, namespacedName types.NamespacedName) {
	gomega.Expect(client.Create(context.Background(), obj)).Should(gomega.Succeed())
	GetObject(client, obj, namespacedName)
}

// GetObject retrieves the resource with the specified name. time options can be provided. The first specifies the
// total time to retry for in seconds (default 10), the second the interval between each attempt in seconds (default 1)
func GetObject(client client.Client, obj client.Object, namespacedName types.NamespacedName, timeOptions ...int) {
	retryTime := 10
	frequency := 1
	if len(timeOptions) > 0 {
		retryTime = timeOptions[0]
	}
	if len(timeOptions) > 1 {
		frequency = timeOptions[1]
	}
	gomega.Eventually(func() error {
		return client.Get(context.Background(), namespacedName, obj)
	}, retryTime, frequency).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Did not find the expected object in time, "+
		"looked for %s: %s. Have you checked the name and namespace matches up "+
		"with the call to this function?", reflect.TypeOf(obj), namespacedName))
}

// GetObjectsMatchingLabels retrieves resources matching labels. time options can be provided. The first specifies the
// total time to retry for in seconds (default 10), the second the interval between each attempt in seconds (default 1)
func GetObjectsMatchingLabels(k8sClient client.Client, obj client.ObjectList, matchLabels map[string]string, timeOptions ...int) {
	retryTime := 10
	frequency := 1
	if len(timeOptions) > 0 {
		retryTime = timeOptions[0]
	}
	if len(timeOptions) > 1 {
		frequency = timeOptions[1]
	}
	gomega.Eventually(func() error {
		return k8sClient.List(context.Background(), obj, &client.ListOptions{LabelSelector: labels.SelectorFromSet(matchLabels)})
	}, retryTime, frequency).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Did not find the expected objects in time, "+
		"looked for objects of type %s with labels: %s. Have you checked that the labels match up?", reflect.TypeOf(obj), matchLabels))
}

// DeleteObject deletes the kube resource then waits for the resource to be gone
func DeleteObject(client client.Client, obj client.Object, namespacedName types.NamespacedName) {
	gomega.Expect(client.Delete(context.Background(), obj)).Should(gomega.Succeed())
	EventuallyDeleted(client, obj, namespacedName)
}

// EventuallyDeleted waits for the resource to be deleted
func EventuallyDeleted(client client.Client, obj client.Object, namespacedName types.NamespacedName) {
	gomega.Eventually(func() bool {
		err := client.Get(context.Background(), namespacedName, obj)
		if err != nil {
			if errors.IsNotFound(err) {
				return true
			}
		}
		return false
	}, 10, 1).Should(gomega.BeTrue())
}

// UpdateObject updates the kube resource of the same name in the cluster, then retrieves the updated
// resource
func UpdateObject(client client.Client, obj client.Object, namespacedName types.NamespacedName) {
	gomega.Expect(client.Update(context.Background(), obj)).Should(gomega.Succeed())
	GetObject(client, obj, namespacedName)
}

// UpdateStatus updates the status of the kube resources of the same name in the cluster, then retrieves the updated
// resource
func UpdateStatus(client client.Client, obj client.Object, namespacedName types.NamespacedName) {
	gomega.Expect(client.Status().Update(context.Background(), obj)).Should(gomega.Succeed())
	GetObject(client, obj, namespacedName)
}

// UpdateCertificateReadiness updates the status of a set of certificate manager certificates to
// be ready, this can be helpful to get past waiting for certificates to be ready
func UpdateCertificateReadiness(client client.Client, namespacedNames ...types.NamespacedName) {
	readyCondition := certmanager.CertificateCondition{
		Type:   certmanager.CertificateConditionReady,
		Status: certmanager.ConditionTrue,
	}
	for _, namespacedName := range namespacedNames {
		cert := &certmanager.Certificate{}
		GetObject(client, cert, namespacedName)
		cert.Status.Conditions = []certmanager.CertificateCondition{readyCondition}
		UpdateStatus(client, cert, namespacedName)
	}
}

// DeleteAllObjects deletes all the resources of the given type in the namespace
func DeleteAllObjects(k8sclient client.Client, obj client.Object, namespace string) {
	gomega.Expect(k8sclient.DeleteAllOf(context.Background(), obj, client.InNamespace(namespace)))
}
