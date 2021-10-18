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
package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler is a struct containing the necessary objects to allow
// a Client to reconcile objects in Kubernetes
type Reconciler struct {
	client.Client
	Ctx          context.Context
	Log          logr.Logger
	MissingKinds map[string]struct{}
}

// Reconcileable is a reconcileable kubernetes object
type Reconcileable interface {
	metav1.Object
	ShouldUpdate(client.Object) (bool, client.Object)
	GetResource() client.Object
	ResourceKind() string
	NewResourceInstance() client.Object
	ResourceIsNil() bool
}

type reconcileOptions struct {
	exitOnChange bool
}

func defaultReconcileOptions() *reconcileOptions {
	return &reconcileOptions{
		exitOnChange: false,
	}
}

// ReconcileOption options that can effect the reconcile of a resource
type ReconcileOption func(*reconcileOptions)

// SetExitOnChange Instructs the Reconcile to exit from the reconcile loop when the resource is thought to have changed. This is
// tempramental as sometimes the resource is defaulted in Kube
func SetExitOnChange(ro *reconcileOptions) {
	ro.exitOnChange = true
}

// Reconcile reconciles the provided Reconcileable object with the equivalent Object in Kubernetes
// Creating, Updating or Deleting the resource as necessary
func (r *Reconciler) Reconcile(namespacedName types.NamespacedName, desired Reconcileable, options ...ReconcileOption) (result ctrl.Result, exit bool, err error) {
	reconcileOptions := defaultReconcileOptions()
	for _, option := range options {
		option(reconcileOptions)
	}

	kind := desired.ResourceKind()
	if _, missing := r.MissingKinds[kind]; missing {
		r.Log.Info("Kind not available", "Kind", kind, "NamespacedName", namespacedName)
		return ctrl.Result{}, false, nil
	}
	r.Log.Info("Reconciling", "Kind", kind, "NamespacedName", namespacedName)
	current := desired.NewResourceInstance()
	err = r.Get(r.Ctx, namespacedName, current)
	if err != nil && errors.IsNotFound(err) {
		current = nil
	} else if err != nil {
		return ctrl.Result{}, true, fmt.Errorf("Failed to get %s: %s", kind, err)
	}

	switch {
	case desired.ResourceIsNil() && current == nil:
		r.Log.V(1).Info("Already removed", "Kind", kind, "NamespacedName", namespacedName)
	case desired.ResourceIsNil() && current != nil:
		return r.delete(kind, namespacedName, current, reconcileOptions.exitOnChange)
	case !desired.ResourceIsNil() && current == nil:
		return r.create(kind, namespacedName, desired.GetResource(), reconcileOptions.exitOnChange)
	case !desired.ResourceIsNil() && current != nil:
		updated, new := desired.ShouldUpdate(current)
		if updated {
			return r.update(kind, namespacedName, new, reconcileOptions.exitOnChange)
		}
	}
	r.Log.V(1).Info("No action required", "Kind", kind, "NamespacedName", namespacedName)
	return ctrl.Result{}, false, nil
}

// update an instance of resourceType in Kubernetes. If the object is successfully updated returns the value of exitOnChange which indicates whether the
// reconcile loop should exit. If the resource is being watched a new reconcile will be triggered by the update
func (r *Reconciler) update(resourceType string, namespacedName types.NamespacedName, updated client.Object, exitOnChange bool) (result ctrl.Result, exit bool, err error) {
	r.Log.V(1).Info("Updating", "resource type", resourceType, "NamespacedName", namespacedName)
	err = r.Update(r.Ctx, updated)
	if err != nil && errors.IsConflict(err) {
		// Object has been updated underneath us, this is likely to occur some time, requeue so it will be
		// sorted in the future
		r.Log.V(1).Info("Requeue due to update conflict", "namespacedName", namespacedName)
		return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, true, nil
	}
	if err != nil {
		return ctrl.Result{}, true, fmt.Errorf("Failed to update %s %s: %s", resourceType, namespacedName, err)
	}
	// return with false indicating that the reconcile should continue this may cause multiple concurrent reconciliations
	return ctrl.Result{}, exitOnChange, nil
}

// create an instance of resourceType in Kube. If the object is successfully created returns the value of exitOnChange which indicates whether the
// reconcile loop should exit. If the resource is being watched a new reconcile will be triggered by the creation
func (r *Reconciler) create(resourceType string, namespacedName types.NamespacedName, created client.Object, exitOnChange bool) (result ctrl.Result, exit bool, err error) {
	r.Log.V(1).Info("Creating", "resource type", resourceType, "NamespacedName", namespacedName)
	err = r.Create(r.Ctx, created)
	if err != nil && errors.IsAlreadyExists(err) {
		// Object has already been created, this is likely to occur some time, requeue so it will be
		// handled by update in the next reconcile
		r.Log.V(1).Info("Requeue due to already exists", "namespacedName", namespacedName)
		return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, true, nil
	}
	if err != nil {
		return ctrl.Result{}, true, fmt.Errorf("Failed to create %s %s: %s", resourceType, namespacedName, err)
	}
	return ctrl.Result{}, exitOnChange, nil
}

// delete an instance of resourceType in Kube. If the object is successfully deleted returns the value of exitOnChange which indicates whether the
// reconcile loop should exit. If the resource is being watched a new reconcile will be triggered by the deletion
func (r *Reconciler) delete(resourceType string, namespacedName types.NamespacedName, deleted client.Object, exitOnChange bool) (result ctrl.Result, exit bool, err error) {
	r.Log.V(1).Info("Deleting", "resource type", resourceType, "NamespacedName", namespacedName)
	err = r.Delete(r.Ctx, deleted)
	if err != nil && errors.IsNotFound(err) {
		// Already deleted, carry ononfigmap
		return ctrl.Result{}, false, nil
	}
	if err != nil {
		return ctrl.Result{}, true, fmt.Errorf("Failed to delete %s %s: %s", resourceType, namespacedName, err)
	}
	return ctrl.Result{}, exitOnChange, nil
}
