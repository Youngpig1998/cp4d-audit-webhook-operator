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

package controllers

import (
	"context"
	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/iaw-shared-helpers/pkg/bootstrap"
	"github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/internal/operator"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkpolicy "k8s.io/api/networking/v1"
	"k8s.io/client-go/rest"
	"time"

	//"fmt"
	//"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	auditv1beta1 "github.ibm.com/watson-foundation-services/cp4d-audit-webhook-operator/api/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"

	//"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/util/intstr"
	//"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	//"strings"
	//"time"
)

var (
	controllerManagerName = "cp4d-audit-webhook-operator-controller-manager"
)




// AuditWebhookReconciler reconciles a AuditWebhook object
type AuditWebhookReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Config *rest.Config
}

//+kubebuilder:rbac:groups=audit.watson.ibm.com,resources=auditwebhooks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=audit.watson.ibm.com,resources=auditwebhooks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=audit.watson.ibm.com,resources=auditwebhooks/finalizers,verbs=update
//+kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.ibm.com,resources=operandrequests,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AuditWebhook object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *AuditWebhookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {


	log := r.Log.WithValues("auditwebhook", req.NamespacedName)


	log.Info("1. start reconcile logic")
	// Instantialize the data structure
	instance := &auditv1beta1.AuditWebhook{}

	//First,query the webhook instance
	err := r.Get(ctx, req.NamespacedName, instance)

	if err != nil {
		// If there is no instance, an empty result is returned, so that the Reconcile method will not be called immediately
		if errors.IsNotFound(err) {
			log.Info("Instance not found, maybe removed")
			return reconcile.Result{}, nil
		}
		log.Error(err, "query action happens error")
		// Return error message
		return ctrl.Result{}, err
	}



	//Set the bootstrapClient's owner value as the webhook,so the resources we create then will be set reference to the webhook
	//when the webhook cr is deleted,the resources(such as deployment.configmap,issuer...) we create will be deleted too
	bootstrapClient, err := bootstrap.NewClient(r.Config,r.Scheme, controllerManagerName,instance)
	if err != nil {
		log.Error(err, "failed to initialise bootstrap client")
		return ctrl.Result{}, err
	}


	//operandRequestName, operandRequest := operator.OperandRequest()
	//
	//done := bootstrapClient.InitialiseCommonServices(operandRequestName, operandRequest)
	//err = <-done
	//if err != nil {
	//	log.Error(err, "failed to initialise common services")
	//	return ctrl.Result{}, err
	//}


	//We create networkpolicy first
	networkPolicyName, networkPolicy := operator.NetworkPolicy()
	err = bootstrapClient.CreateResource(networkPolicyName, networkPolicy)
	if err != nil {
		log.Error(err, "failed to create operator NetworkPolicy", "Name", networkPolicyName)
		return ctrl.Result{}, err
	}

	issuerName, issuer := operator.Issuer()
	err = bootstrapClient.CreateResource(issuerName, issuer)
	if err != nil {
		log.Error(err, "failed to create operator issuer", "Name", issuerName)
		return ctrl.Result{}, err
	}

	certificateName, certificate := operator.Certificate(instance)
	err = bootstrapClient.CreateResource(certificateName, certificate)
	if err != nil {
		log.Error(err, "failed to create operator certificate", "Name", certificateName)
		return ctrl.Result{}, err
	}


	secretData,err := r.getSecretData(ctx,req)
	if err != nil {
		log.Error(err, "failed to get  secret's data")
		return ctrl.Result{}, err
	}

	secretName, secret := operator.Secret(secretData)
	err = bootstrapClient.CreateResource(secretName, secret)
	if err != nil {
		log.Error(err, "failed to create operator secret", "Name", secretName)
		return ctrl.Result{}, err
	}


	configMapName, configMap := operator.ConfigMap(instance)
	err = bootstrapClient.CreateResource(configMapName, configMap)
	if err != nil {
		log.Error(err, "failed to create operator configMap", "Name", configMapName)
		return ctrl.Result{}, err
	}


	serviceName, service := operator.Service()
	err = bootstrapClient.CreateResource(serviceName, service)
	if err != nil {
		log.Error(err, "failed to create operator service", "Name", serviceName)
		return ctrl.Result{}, err
	}

	deploymentName, deployment := operator.Deployment(instance)
	err = bootstrapClient.CreateResource(deploymentName, deployment)
	if err != nil {
		log.Error(err, "failed to create operator Deployment", "Name", deploymentName)
		return ctrl.Result{}, err
	}

	mcName, mc := operator.MutatingWebhookConfiguration(instance)
	err = bootstrapClient.CreateResource(mcName, mc)
	if err != nil {
		log.Error(err, "failed to create operator Mc", "Name", mcName)
		return ctrl.Result{}, err
	}


	return ctrl.Result{}, nil
}




// SetupWithManager sets up the controller with the Manager.
func (r *AuditWebhookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&auditv1beta1.AuditWebhook{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&networkpolicy.NetworkPolicy{}).
		Owns(&certmanager.Issuer{}).
		Owns(&certmanager.Certificate{}).
		Owns(&admissionregistrationv1.MutatingWebhookConfiguration{}).
		Complete(r)
}



// Fetch the secretData From the secret created by certmanager
func (r *AuditWebhookReconciler) getSecretData(ctx context.Context, req ctrl.Request) (map[string][]byte,error) {
	log := r.Log
	log.WithValues("func", "getSecretData")
	log.Info("start to get secret's data")

	//Wait For certmanager creating secret
	time.Sleep(5000 * time.Millisecond)

	secret := &corev1.Secret{}

	// Query secrets through client tools
	err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: "audit-webhook-tls-secret"}, secret)


	//If the secret exists
	if err == nil {
		log.Info("secret exists")
		// Critical step
		log.Info("set reference")

		//Save the secret's data
		var secretData = secret.Data

		return secretData,nil
	}else {
		return nil,err
	}

}
