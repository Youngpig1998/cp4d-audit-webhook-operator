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
package commonservices

import (
	"context"
	"fmt"
	"time"

	odlmv1alpha1 "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/resources/operandrequests"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheme                     = runtime.NewScheme()
	logger                     = ctrl.Log.WithName("init-common-services")
	operandRequestGroupVersion = "operator.ibm.com/v1alpha1"
)

// Client is a client to help initialise common services via an OperandRequest
type Client struct {
	TimeoutSeconds  int
	IntervalSeconds int
	Scheme          *runtime.Scheme
	Owner           metav1.Object
	DiscoveryClient *discovery.DiscoveryClient
	KubeClient      client.Client
	Context         context.Context
	maxRetries      int
	retry           int
}

// InitialiseCommonServices dependencies. This waits until the OperandRequest API is available, then creates or updates
// the OperandRequest, then waits till that OperandRequest is ready. Returns an error if something has gone wrong, indicating
// the operator should restart
func (c Client) InitialiseCommonServices(operandRequestName types.NamespacedName, operandRequest *odlmv1alpha1.OperandRequest) chan error {
	done := make(chan error)
	go func() {
		c.maxRetries = c.TimeoutSeconds / c.IntervalSeconds
		c.retry = 0
		logger.Info("Waiting for the OperandRequest Kind to be available")
		err := c.waitForOperandRequestToBeAvailable()
		if err != nil {
			done <- err
			return
		}
		logger.Info("OperandRequest Kind is available")

		logger.Info("Creating or Updating the OperandRequest for the operator")
		err = c.createOrUpdateOperandRequest(operandRequestName, operandRequest)
		if err != nil {
			done <- err
			return
		}
		logger.Info("OperandRequest created", "OperandRequest", operandRequestName)

		logger.Info("Waiting for the OperandRequest to be ready", "OperandRequest", operandRequestName)
		err = c.waitForOperandRequestToBeReady(operandRequestName)
		if err != nil {
			done <- err
			return
		}
		logger.Info("OperandRequest is ready", "OperandRequest", operandRequestName)

		done <- nil
	}()
	return done
}

func (c Client) waitForOperandRequestToBeAvailable() error {
	for ; c.retry <= c.maxRetries; c.retry++ {
		_, resources, err := c.DiscoveryClient.ServerGroupsAndResources()
		if err != nil {
			if !discovery.IsGroupDiscoveryFailedError(err) {
				logger.Error(err, "Failed to get server groups and resources", "Attempt", c.retry, "Max Retries", c.maxRetries)
				time.Sleep(time.Duration(c.IntervalSeconds) * time.Second)
				continue
			}
			logger.Info("Failed to get all server groups and resources", "Failed Groups", err.Error())
		}
		for _, resourceList := range resources {
			if resourceList.GroupVersion == operandRequestGroupVersion {
				for _, resource := range resourceList.APIResources {
					if resource.Kind == "OperandRequest" {
						logger.Info("OperandRequest Kind has been found")
						return nil
					}
				}
			}
		}
		logger.Info("OperandRequest Kind not yet available", "Attempt", c.retry, "Max Retries", c.maxRetries)
		time.Sleep(time.Duration(c.IntervalSeconds) * time.Second)
	}
	return fmt.Errorf("More than %d seconds have passed and the OperandRequest Kind is not available", c.TimeoutSeconds)
}

func (c Client) createOrUpdateOperandRequest(operandRequestName types.NamespacedName, operandRequest *odlmv1alpha1.OperandRequest) error {

	ctrl.SetControllerReference(c.Owner, operandRequest, c.Scheme)

	resourceClient := resources.Reconciler{
		Client:       c.KubeClient,
		Ctx:          c.Context,
		Log:          logger,
		MissingKinds: map[string]struct{}{},
	}
	_, _, err := resourceClient.Reconcile(operandRequestName, operandrequests.From(operandRequest))
	return err
}

func (c Client) waitForOperandRequestToBeReady(operandRequestName types.NamespacedName) error {
	for ; c.retry <= c.maxRetries; c.retry++ {
		current := &odlmv1alpha1.OperandRequest{}
		err := c.KubeClient.Get(c.Context, operandRequestName, current)
		if err != nil {
			logger.Error(err, "Failed to get OperandRequest", "OperandReqest", operandRequestName, "Attempt", c.retry, "Max Retries", c.maxRetries)
			time.Sleep(time.Duration(c.IntervalSeconds) * time.Second)
			continue
		}
		if len(current.Status.Members) == 0 {
			logger.Info("OperandRequest: not yet ready, members have no status", "OperandRequest", operandRequestName, "Attempt", c.retry, "Max Retries", c.maxRetries)
			time.Sleep(time.Duration(c.IntervalSeconds) * time.Second)
			continue
		}
		allDepsReady := true
		for _, member := range current.Status.Members {
			if member.Phase.OperatorPhase == "" && member.Phase.OperandPhase == "" {
				logger.Info("Dependency not ready, no phase information available", "Dependency", member.Name, "Attempt", c.retry, "Max Retries", c.maxRetries)
				allDepsReady = false
				continue
			}
			if member.Phase.OperatorPhase != "" && member.Phase.OperatorPhase != odlmv1alpha1.OperatorRunning {
				logger.Info("Dependency not ready", "Dependency", member.Name, "Operator Phase", member.Phase, "Attempt", c.retry, "Max Retries", c.maxRetries)
				allDepsReady = false
				continue
			}
			if member.Phase.OperandPhase != "" && member.Phase.OperandPhase != odlmv1alpha1.ServiceRunning {
				logger.Info("Dependency not ready", "Dependency", member.Name, "Operand Phase", member.Phase, "Attempt", c.retry, "Max Retries", c.maxRetries)
				allDepsReady = false
			}
		}
		if allDepsReady {
			logger.Info("All dependencies ready")
			return nil
		}
		time.Sleep(time.Duration(c.IntervalSeconds) * time.Second)
	}
	return fmt.Errorf("Operand request %s, has taken more than %d seconds to become ready", operandRequestName, c.TimeoutSeconds)
}
