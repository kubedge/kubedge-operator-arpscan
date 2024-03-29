// Copyright 2019 The Kubedge Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package arpscan

import (
	"context"
	"fmt"
	"reflect"

	arpscanmgr "github.com/kubedge/kubedge-operator-arpscan/pkg/arpscan"
	av1 "github.com/kubedge/kubedge-operator-base/pkg/apis/kubedgeoperators/v1alpha1"
	bcontroller "github.com/kubedge/kubedge-operator-base/pkg/controller/kubedgecontroller"
	bmgr "github.com/kubedge/kubedge-operator-base/pkg/kubedgemanager"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	crthandler "sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	crtpredicate "sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var arpscanlog = logf.Log.WithName("arpscan-controller")

// AddArpscanController creates a new Arpscan Controller and adds it to
// the Manager. The Manager will set fields on the Controller and Start it when
// the Manager is Started.
func AddArpscanController(mgr manager.Manager) error {
	return addArpscan(mgr, newArpscanReconciler(mgr))
}

// newArpscanReconciler returns a new reconcile.Reconciler
func newArpscanReconciler(mgr manager.Manager) reconcile.Reconciler {
	r := &ArpscanReconciler{
		KubedgeBaseReconciler: bcontroller.KubedgeBaseReconciler{
			Client:         mgr.GetClient(),
			Scheme:         mgr.GetScheme(),
			Recorder:       mgr.GetEventRecorderFor("arpscan-recorder"),
			ManagerFactory: arpscanmgr.NewManagerFactory(mgr),
			// reconcilePeriod: flags.ReconcilePeriod,
		},
	}
	return r
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func addArpscan(mgr manager.Manager, r reconcile.Reconciler) error {

	// Create a new controller
	c, err := controller.New("arpscan-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Arpscan
	// EnqueueRequestForObject enqueues a Request containing the Name and Namespace of the object
	// that is the source of the Event. (e.g. the created / deleted / updated objects Name and Namespace).
	// err = c.Watch(&source.Kind{Type: &av1.Arpscan{}}, &crthandler.EnqueueRequestForObject{})
	err = c.Watch(source.Kind(mgr.GetCache(), &av1.Arpscan{}), &crthandler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	dependentPredicate := crtpredicate.Funcs{
		// We don't need to reconcile dependent resource creation events
		// because dependent resources are only ever created during
		// reconciliation. Another.Reconcile would be redundant.
		CreateFunc: func(e event.CreateEvent) bool {
			o := e.Object.(*unstructured.Unstructured)
			arpscanlog.Info("CreateEvent. Filtering", "resource", o.GetName(), "namespace", o.GetNamespace(),
				"apiVersion", o.GroupVersionKind().GroupVersion(), "kind", o.GroupVersionKind().Kind)
			return false
		},

		// Reconcile when a dependent resource is deleted so that it can be
		// recreated.
		DeleteFunc: func(e event.DeleteEvent) bool {
			o := e.Object.(*unstructured.Unstructured)
			arpscanlog.Info("DeleteEvent. Triggering", "resource", o.GetName(), "namespace", o.GetNamespace(),
				"apiVersion", o.GroupVersionKind().GroupVersion(), "kind", o.GroupVersionKind().Kind)
			return true
		},

		// Reconcile when a dependent resource is updated, so that it can
		// be patched back to the resource managed by the Argo workflow, if
		// necessary. Ignore updates that only change the status and
		// resourceVersion.
		UpdateFunc: func(e event.UpdateEvent) bool {
			old := e.ObjectOld.(*unstructured.Unstructured).DeepCopy()
			new := e.ObjectNew.(*unstructured.Unstructured).DeepCopy()

			delete(old.Object, "status")
			delete(new.Object, "status")
			old.SetResourceVersion("")
			new.SetResourceVersion("")

			if reflect.DeepEqual(old.Object, new.Object) {
				arpscanlog.Info("UpdateEvent. Filtering", "resource", new.GetName(), "namespace", new.GetNamespace(),
					"apiVersion", new.GroupVersionKind().GroupVersion(), "kind", new.GroupVersionKind().Kind)
				return false
			} else {
				arpscanlog.Info("UpdateEvent. Triggering", "resource", new.GetName(), "namespace", new.GetNamespace(),
					"apiVersion", new.GroupVersionKind().GroupVersion(), "kind", new.GroupVersionKind().Kind)
				return true
			}
		},
	}

	// Watch for changes to secondary resource (described in the yaml files) and requeue the owner Arpscan
	// EnqueueRequestForOwner enqueues Requests for the Owners of an object. E.g. the object
	// that created the object that was the source of the Event
	if racr, isArpscanReconciler := r.(*ArpscanReconciler); isArpscanReconciler {
		// The enqueueRequestForOwner is not actually done here since we don't know yet the
		// content of the yaml files. The tools wait for the yaml files to be parse. The manager
		// then add the "OwnerReference" to the content of the yaml files. It then invokes the EnqueueRequestForOwner
		owner := av1.NewArpscanVersionKind("", "")
		racr.DepResourceWatchUpdater = bmgr.BuildDependentResourceWatchUpdater(mgr, owner, c, dependentPredicate)
	} else if rrf, isReconcileFunc := r.(*reconcile.Func); isReconcileFunc {
		// Unit test issue
		log.Info("UnitTests", "ReconfileFunc", rrf)
	}

	return nil
}

var _ reconcile.Reconciler = &ArpscanReconciler{}

// ArpscanReconciler.Reconciles custom resources as Argo workflows.
type ArpscanReconciler struct {
	bcontroller.KubedgeBaseReconciler
}

const (
	finalizerArpscan = "uninstall-arpscan-resource"
)

// Reconcile reads that state of the cluster for an Arpscan object and
// makes changes based on the state read and what is in the Arpscan.Spec
//
// Note: The Controller will requeue the Request to be processed again if the
// returned error is non-nil or Result.Requeue is true, otherwise upon
// completion it will remove the work from the queue.
func (r *ArpscanReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	reclog := arpscanlog.WithValues("namespace", request.Namespace, "arpscan", request.Name)
	reclog.Info("Received a request")

	instance := &av1.Arpscan{}
	instance.SetNamespace(request.Namespace)
	instance.SetName(request.Name)

	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	instance.Init()

	if apierrors.IsNotFound(err) {
		// We are working asynchronously. By the time we receive the event,
		// the object could already be gone
		return reconcile.Result{}, nil
	}

	if err != nil {
		reclog.Error(err, "Failed to lookup Arpscan")
		return reconcile.Result{}, err
	}

	mgr := r.ManagerFactory.NewArpscanManager(instance)
	reclog = reclog.WithValues("arpscan", mgr.ResourceName())

	var shouldRequeue bool
	if shouldRequeue, err = r.updateFinalizers(instance); shouldRequeue {
		// Need to requeue because finalizer update does not change metadata.generation
		return reconcile.Result{Requeue: true}, err
	}

	if err := r.ensureSynced(mgr, instance); err != nil {
		if !instance.IsDeleted() {
			// TODO(jeb): Changed the behavior to stop only if we are not
			// in a delete phase.
			return reconcile.Result{}, err
		}
	}

	if instance.IsDeleted() {
		if shouldRequeue, err = r.deleteArpscan(mgr, instance); shouldRequeue {
			// Need to requeue because finalizer update does not change metadata.generation
			return reconcile.Result{Requeue: true}, err
		}
		return reconcile.Result{}, err
	}

	if instance.IsSatisfied() {
		reclog.Info("Already satisfied; skipping")
		err = r.updateResource(instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		err = r.Client.Status().Update(context.TODO(), instance)
		return reconcile.Result{}, err
	}

	hrc := av1.KubedgeCondition{
		Type:   av1.ConditionInitialized,
		Status: av1.ConditionStatusTrue,
	}
	instance.Status.SetCondition(hrc, instance.Spec.TargetState)

	switch {
	case !mgr.IsInstalled():
		if shouldRequeue, err = r.installArpscan(mgr, instance); shouldRequeue {
			return reconcile.Result{RequeueAfter: r.ReconcilePeriod}, err
		}
		return reconcile.Result{}, err
	case mgr.IsUpdateRequired():
		if shouldRequeue, err = r.updateArpscan(mgr, instance); shouldRequeue {
			return reconcile.Result{RequeueAfter: r.ReconcilePeriod}, err
		}
		return reconcile.Result{}, err
	}

	if err := r.reconcileArpscan(mgr, instance); err != nil {
		return reconcile.Result{}, err
	}

	reclog.Info("Reconciled Arpscan")
	err = r.updateResourceStatus(instance)
	return reconcile.Result{RequeueAfter: r.ReconcilePeriod}, err
}

// logAndRecordFailure adds a failure event to the recorder
func (r ArpscanReconciler) logAndRecordFailure(instance *av1.Arpscan, hrc *av1.KubedgeCondition, err error) {
	reclog := arpscanlog.WithValues("namespace", instance.Namespace, "arpscan", instance.Name)
	reclog.Error(err, fmt.Sprintf("%s. ErrorCondition", hrc.Type.String()))
	r.Recorder.Event(instance, corev1.EventTypeWarning, hrc.Type.String(), hrc.Reason.String())
}

// logAndRecordSuccess adds a success event to the recorder
func (r ArpscanReconciler) logAndRecordSuccess(instance *av1.Arpscan, hrc *av1.KubedgeCondition) {
	reclog := arpscanlog.WithValues("namespace", instance.Namespace, "arpscan", instance.Name)
	reclog.Info(fmt.Sprintf("%s. SuccessCondition", hrc.Type.String()))
	r.Recorder.Event(instance, corev1.EventTypeNormal, hrc.Type.String(), hrc.Reason.String())
}

// updateResource updates the Resource object in the cluster
func (r ArpscanReconciler) updateResource(instance *av1.Arpscan) error {
	return r.Client.Update(context.TODO(), instance)
}

// updateResourceStatus updates the the Status field of the Resource object in the cluster
func (r ArpscanReconciler) updateResourceStatus(instance *av1.Arpscan) error {
	reclog := arpscanlog.WithValues("namespace", instance.Namespace, "arpscan", instance.Name)

	helper := av1.KubedgeConditionListHelper{Items: instance.Status.Conditions}
	instance.Status.Conditions = helper.InitIfEmpty()

	// JEB: Be sure to have update status subresources in the CRD.yaml
	// JEB: Look for kubebuilder subresources in the _types.go
	err := r.Client.Status().Update(context.TODO(), instance)
	if err != nil {
		reclog.Error(err, "Failure to update status. Ignoring")
		err = nil
	}

	return err
}

// ensureSynced checks that the KubedgeResourceManager is in sync with the cluster
func (r ArpscanReconciler) ensureSynced(mgr bmgr.KubedgeResourceManager, instance *av1.Arpscan) error {
	if err := mgr.Sync(context.TODO()); err != nil {
		hrc := av1.KubedgeCondition{
			Type:    av1.ConditionIrreconcilable,
			Status:  av1.ConditionStatusTrue,
			Reason:  av1.ReasonReconcileError,
			Message: err.Error(),
		}
		instance.Status.SetCondition(hrc, instance.Spec.TargetState)
		r.logAndRecordFailure(instance, &hrc, err)
		_ = r.updateResourceStatus(instance)
		return err
	}
	instance.Status.RemoveCondition(av1.ConditionIrreconcilable)
	return nil
}

// updateFinalizers asserts that the finalizers match what is expected based on
// whether the instance is currently being deleted or not. It returns true if
// the finalizers were changed, false otherwise
func (r ArpscanReconciler) updateFinalizers(instance *av1.Arpscan) (bool, error) {
	pendingFinalizers := instance.GetFinalizers()
	if !instance.IsDeleted() && !r.Contains(pendingFinalizers, finalizerArpscan) {
		finalizers := append(pendingFinalizers, finalizerArpscan)
		instance.SetFinalizers(finalizers)
		err := r.updateResource(instance)

		return true, err
	}
	return false, nil
}

// watchDependentResources updates all resources which are dependent on this one
func (r ArpscanReconciler) watchDependentResources(resource *av1.SubResourceList) error {
	if r.DepResourceWatchUpdater != nil {
		if err := r.DepResourceWatchUpdater(resource.GetDependentResources()); err != nil {
			return err
		}
	}
	return nil
}

// deleteArpscan deletes an instance of an Arpscan. It returns true if the reconciler should be re-enqueueed
func (r ArpscanReconciler) deleteArpscan(mgr bmgr.KubedgeResourceManager, instance *av1.Arpscan) (bool, error) {
	reclog := arpscanlog.WithValues("namespace", instance.Namespace, "arpscan", instance.Name)
	reclog.Info("Deleting")

	pendingFinalizers := instance.GetFinalizers()
	if !r.Contains(pendingFinalizers, finalizerArpscan) {
		reclog.Info("Arpscan is terminated, skipping reconciliation")
		return false, nil
	}

	uninstalledResource, err := mgr.UninstallResource(context.TODO())
	if err != nil && err != bmgr.ErrNotFound {
		hrc := av1.KubedgeCondition{
			Type:         av1.ConditionFailed,
			Status:       av1.ConditionStatusTrue,
			Reason:       av1.ReasonUninstallError,
			Message:      err.Error(),
			ResourceName: uninstalledResource.GetName(),
		}
		instance.Status.SetCondition(hrc, instance.Spec.TargetState)
		r.logAndRecordFailure(instance, &hrc, err)

		_ = r.updateResourceStatus(instance)
		return false, err
	}
	instance.Status.RemoveCondition(av1.ConditionFailed)

	if err == bmgr.ErrNotFound {
		reclog.Info("Resource already uninstalled, Removing finalizer")
	} else {
		hrc := av1.KubedgeCondition{
			Type:   av1.ConditionDeployed,
			Status: av1.ConditionStatusFalse,
			Reason: av1.ReasonUninstallSuccessful,
		}
		instance.Status.SetCondition(hrc, instance.Spec.TargetState)
		r.logAndRecordSuccess(instance, &hrc)
	}
	if err := r.updateResourceStatus(instance); err != nil {
		return false, err
	}

	finalizers := []string{}
	for _, pendingFinalizer := range pendingFinalizers {
		if pendingFinalizer != finalizerArpscan {
			finalizers = append(finalizers, pendingFinalizer)
		}
	}
	instance.SetFinalizers(finalizers)
	err = r.updateResource(instance)

	return true, err
}

// installArpscan attempts to install instance. It returns true if the reconciler should be re-enqueueed
func (r ArpscanReconciler) installArpscan(mgr bmgr.KubedgeResourceManager, instance *av1.Arpscan) (bool, error) {
	reclog := arpscanlog.WithValues("namespace", instance.Namespace, "arpscan", instance.Name)
	reclog.Info("Installing`")

	installedResource, err := mgr.InstallResource(context.TODO())
	if err != nil {
		hrc := av1.KubedgeCondition{
			Type:    av1.ConditionFailed,
			Status:  av1.ConditionStatusTrue,
			Reason:  av1.ReasonInstallError,
			Message: err.Error(),
		}
		instance.Status.SetCondition(hrc, instance.Spec.TargetState)
		r.logAndRecordFailure(instance, &hrc, err)

		_ = r.updateResourceStatus(instance)
		return false, err
	}
	instance.Status.RemoveCondition(av1.ConditionFailed)

	if err := r.watchDependentResources(installedResource); err != nil {
		reclog.Error(err, "Failed to update watch on dependent resources")
		return false, err
	}

	hrc := av1.KubedgeCondition{
		Type:            av1.ConditionDeployed,
		Status:          av1.ConditionStatusTrue,
		Reason:          av1.ReasonInstallSuccessful,
		Message:         installedResource.GetNotes(),
		ResourceName:    installedResource.GetName(),
		ResourceVersion: installedResource.GetVersion(),
	}
	instance.Status.SetCondition(hrc, instance.Spec.TargetState)
	r.logAndRecordSuccess(instance, &hrc)

	err = r.updateResourceStatus(instance)
	return true, err
}

// updateArpscan attempts to update instance. It returns true if the reconciler should be re-enqueueed
func (r ArpscanReconciler) updateArpscan(mgr bmgr.KubedgeResourceManager, instance *av1.Arpscan) (bool, error) {
	reclog := arpscanlog.WithValues("namespace", instance.Namespace, "arpscan", instance.Name)
	reclog.Info("Updating`")

	previousResource, updatedResource, err := mgr.UpdateResource(context.TODO())
	if previousResource != nil && updatedResource != nil {
		reclog.Info("UpdateResource", "Previous", previousResource.GetName(), "Updated", updatedResource.GetName())
	}
	if err != nil {
		hrc := av1.KubedgeCondition{
			Type:         av1.ConditionFailed,
			Status:       av1.ConditionStatusTrue,
			Reason:       av1.ReasonUpdateError,
			Message:      err.Error(),
			ResourceName: updatedResource.GetName(),
		}
		instance.Status.SetCondition(hrc, instance.Spec.TargetState)
		r.logAndRecordFailure(instance, &hrc, err)

		_ = r.updateResourceStatus(instance)
		return false, err
	}
	instance.Status.RemoveCondition(av1.ConditionFailed)

	if err := r.watchDependentResources(updatedResource); err != nil {
		reclog.Error(err, "Failed to update watch on dependent resources")
		return false, err
	}

	hrc := av1.KubedgeCondition{
		Type:            av1.ConditionDeployed,
		Status:          av1.ConditionStatusTrue,
		Reason:          av1.ReasonUpdateSuccessful,
		Message:         updatedResource.GetNotes(),
		ResourceName:    updatedResource.GetName(),
		ResourceVersion: updatedResource.GetVersion(),
	}
	instance.Status.SetCondition(hrc, instance.Spec.TargetState)
	r.logAndRecordSuccess(instance, &hrc)

	err = r.updateResourceStatus(instance)
	return true, err
}

// reconcileArpscan reconciles the yaml files with the cluster
func (r ArpscanReconciler) reconcileArpscan(mgr bmgr.KubedgeResourceManager, instance *av1.Arpscan) error {
	reclog := arpscanlog.WithValues("namespace", instance.Namespace, "arpscan", instance.Name)
	reclog.Info("Reconciling Arpscan")

	expectedResource, err := mgr.ReconcileResource(context.TODO())
	if err != nil {
		hrc := av1.KubedgeCondition{
			Type:         av1.ConditionIrreconcilable,
			Status:       av1.ConditionStatusTrue,
			Reason:       av1.ReasonReconcileError,
			Message:      err.Error(),
			ResourceName: expectedResource.GetName(),
		}
		instance.Status.SetCondition(hrc, instance.Spec.TargetState)
		r.logAndRecordFailure(instance, &hrc, err)

		_ = r.updateResourceStatus(instance)
		return err
	}
	instance.Status.RemoveCondition(av1.ConditionIrreconcilable)
	if err := r.watchDependentResources(expectedResource); err != nil {
		reclog.Error(err, "Failed to update watch on dependent resources")
		return err
	}
	return nil
}
