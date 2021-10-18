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
package bootstrap

import (
	"context"
	"fmt"
	"github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
	auditv1beta1 "github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/api/v1beta1"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/commonservices"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources"
	//appsv1 "k8s.io/api/apps/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client is a Kubernetes bootstrap client for an operator
type Client struct {
	DiscoveryClient *discovery.DiscoveryClient
	kubeClient      client.Client
	Owner           *auditv1beta1.AuditWebhook
	resourceClient  resources.Reconciler
	context         context.Context
	scheme          *runtime.Scheme
	namespace       string
}

var (
	logger = ctrl.Log.WithName("bootstrap-operator")
)

// NewClient creates a new bootstrap client to be used at operator install time.
// It instantiates relevant clients to be used and sets up an owner, context, scheme
// and install namespace to be referenced.
func NewClient(config *rest.Config, scheme *runtime.Scheme, controllerManagerName string,owner *auditv1beta1.AuditWebhook) (*Client, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	kubeClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	// TODO: I believe this is one of our other requirements to include
	// OPERATOR_NAMESPACE in our resources (explictly) along with WATCH_NAMESPACE
	// (even if WATHC_NAMESPACE is empty by default)

	//namespace, exists := os.LookupEnv("OPERATOR_NAMESPACE")
	//if !exists {
	//	return nil, fmt.Errorf("Operator namespace not set")
	//}
	namespace := owner.Namespace

	context := context.Background()
	//owner := &appsv1.Deployment{             //这边创建的是operator运行的那个deployment
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      controllerManagerName,
	//		Namespace: namespace,
	//	},
	//}

	//Alreday checkout the owner
	//err = kubeClient.Get(context, types.NamespacedName{Name: owner.Name, Namespace: owner.Namespace}, owner)
	//if err != nil {
	//	return nil, err
	//}

	resourceClient := resources.Reconciler{
		Client:       kubeClient,
		Ctx:          context,
		Log:          logger,
		MissingKinds: map[string]struct{}{},
	}

	return &Client{
		DiscoveryClient: discoveryClient,
		kubeClient:      kubeClient,
		Owner:           owner,
		resourceClient:  resourceClient,
		context:         context,
		scheme:          scheme,
		namespace:       namespace,
	}, nil
}

// CreateResource facilitates the generic creation of any resource to be created with
// and managed by the Operator.
func (c Client) CreateResource(name string, resource resources.Reconcileable) error {

	resourceNamespacedName := types.NamespacedName{Name: name, Namespace: c.namespace}
	if !resource.ResourceIsNil() {
		resource.SetNamespace(c.namespace)
	}

	ctrl.SetControllerReference(c.Owner, resource, c.scheme)

	_, _, err := c.resourceClient.Reconcile(resourceNamespacedName, resource)
	return err
}



// InitialiseCommonServices is a wrapper around the commonServicesClient.InitialiseCommonServices method to allow for
// a custom OperandRequest and name to be easily provided and created in the install namespace
func (c Client) InitialiseCommonServices(operandRequestName string, operandRequest *v1alpha1.OperandRequest) chan error {
	commonServicesClient := &commonservices.Client{
		TimeoutSeconds:  600,
		IntervalSeconds: 10,
		KubeClient:      c.kubeClient,
		Scheme:          c.scheme,
		Context:         c.context,
		Owner:           c.Owner,
		DiscoveryClient: c.DiscoveryClient,
	}
	operandRequestNamespacedName := types.NamespacedName{Name: operandRequestName, Namespace: c.namespace}
	if operandRequest != nil {
		operandRequest.ObjectMeta.Namespace = c.namespace
	}
	done := commonServicesClient.InitialiseCommonServices(operandRequestNamespacedName, operandRequest)
	return done
}



// CheckAPIGroups that are currently supported against those that are expected to exist. If they
// are found, return an empty map, otherwise return any kinds that are missing.
// Example input map: {"route.openshift.io": {"Route"}, "operator.ibm.com": {"OperandRequest"}}
func (c Client) CheckAPIGroups(optionalAPIGroups map[string][]string, requiredAPIGroups map[string][]string) (map[string]struct{}, error) {
	apiGroups, _, err := c.DiscoveryClient.ServerGroupsAndResources()
	if err != nil {
		if !discovery.IsGroupDiscoveryFailedError(err) {
			return nil, err
		}
		logger.Info("Failed to get all server groups and resources", "Failed Groups", err.Error())
	}

	for _, apiGroup := range apiGroups {
		_, apiGroupIsPresent := optionalAPIGroups[apiGroup.Name]
		if apiGroupIsPresent {
			delete(optionalAPIGroups, apiGroup.Name)
		}
		_, apiGroupIsPresent = requiredAPIGroups[apiGroup.Name]
		if apiGroupIsPresent {
			delete(requiredAPIGroups, apiGroup.Name)
		}
	}

	missingKinds := map[string]struct{}{}
	for _, kinds := range optionalAPIGroups {
		for _, kind := range kinds {
			missingKinds[kind] = struct{}{}
		}
	}

	if len(requiredAPIGroups) != 0 {
		return missingKinds, fmt.Errorf("Some required API Groups are missing: %s", requiredAPIGroups)
	}

	return missingKinds, nil
}
