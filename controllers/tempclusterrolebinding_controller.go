/*
Copyright 2022.

Licensed under the Apache License, VersionAnnotation 2.0 (the "License");
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
	"fmt"
	"time"

	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tmprbacv1 "rnemet.dev/temprolebindings/api/v1"
	"rnemet.dev/temprolebindings/base"
)

const (
	finalizerNameCluster = "tcrb.tmprbac.rnemet.dev/finalizer"
)

var (
	ownerKeyCluster = ".metadata.controller.tcrb"
)

// TempClusterRoleBindingReconciler reconciles a TempClusterRoleBinding object
type TempClusterRoleBindingReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tmprbac.rnemet.dev,resources=tempclusterrolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tmprbac.rnemet.dev,resources=tempclusterrolebindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tmprbac.rnemet.dev,resources=tempclusterrolebindings/finalizers,verbs=update
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TempClusterRoleBinding object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *TempClusterRoleBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var tempClusterRoleBinding tmprbacv1.TempClusterRoleBinding

	if err := r.Get(ctx, req.NamespacedName, &tempClusterRoleBinding); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if tempClusterRoleBinding.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&tempClusterRoleBinding, finalizerNameCluster) {
			controllerutil.AddFinalizer(&tempClusterRoleBinding, finalizerNameCluster)
			return ctrl.Result{}, r.Update(ctx, &tempClusterRoleBinding)
		}
	} else {
		if controllerutil.ContainsFinalizer(&tempClusterRoleBinding, finalizerNameCluster) {
			log.Info("Deleting external Resorces")

			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(ctx, tempClusterRoleBinding); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(&tempClusterRoleBinding, finalizerNameCluster)
			if err := r.Update(ctx, &tempClusterRoleBinding); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// calculate next status
	currentStatus, nextStatus := base.GetCurrentAndNextStatus(tempClusterRoleBinding.Status.ToBaseStatus(), tempClusterRoleBinding.Annotations, tmprbacv1.BaseSpec(tempClusterRoleBinding.Spec))
	// excute translation to next status
	result, err := r.executeTransition(ctx, tempClusterRoleBinding, currentStatus, nextStatus)
	if err != nil {
		log.Error(err, "[TCRB] cound not find TempClusterRoleBinding")
		return ctrl.Result{}, err
	}
	// save status
	if err := r.reconcileStatus(ctx, tempClusterRoleBinding, nextStatus); err != nil {
		log.Error(err, "[TCRB] cound not save Status")
		return ctrl.Result{}, err
	}

	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TempClusterRoleBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &rbac.ClusterRoleBinding{}, ownerKeyCluster, func(rawObj client.Object) []string {
		// grab the job object, extract the owner...
		rb := rawObj.(*rbac.ClusterRoleBinding)
		owner := metav1.GetControllerOf(rb)
		if owner == nil {
			return nil
		}
		// ...make sure it's a TempRoleBinding...
		if owner.APIVersion != apiGVStr || owner.Kind != "TempClusterRoleBinding" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&tmprbacv1.TempClusterRoleBinding{}).
		Owns(&rbac.ClusterRoleBinding{}).
		Complete(r)
}

// executeTransition execute transition from one to another status
func (r *TempClusterRoleBindingReconciler) executeTransition(ctx context.Context, tcrb tmprbacv1.TempClusterRoleBinding, status tmprbacv1.BaseStatus, next tmprbacv1.BaseStatus) (ctrl.Result, error) {

	log := log.FromContext(ctx)
	log.Info(fmt.Sprintf("[TCRB] Executing transaction from %s to %s", status.Phase, next.Phase))
	// nothing happen
	if status.Phase == next.Phase {
		log.Info("[TCRB] Current and next Phase are the same. No Transition.")
		if status.Phase == tmprbacv1.TempRoleBindingStatusExpired {
			log.Info("Exired. Deleting external resorces")
			return ctrl.Result{}, r.deleteExternalResources(ctx, tcrb)
		}
		return ctrl.Result{}, nil
	}

	result, ok := base.SwitchFromPendingToApproved(status, next)
	if ok {
		return result, nil
	}

	result, ok = base.SwitchFromPendingToDeclined(status, next)
	if ok {
		return result, nil
	}

	result, ok = base.SwitchFromAppliedToExpired(status, next)
	if ok {
		return result, nil
	}

	result, ok = base.SwitchFromApprovedToHold(status, next)
	if ok {
		return result, nil
	}

	result, ok, err := r.switchFromApprovedToApplied(ctx, tcrb, status, next)
	if ok {
		return result, err
	}

	return ctrl.Result{}, errors.NewBadRequest(fmt.Sprintf("Invalid Transition for TempRoleBinding from %v to %v", status.Phase, next.Phase))
}

// reconcileStatus save TempRoleBindingStatus
func (r *TempClusterRoleBindingReconciler) reconcileStatus(ctx context.Context, trb tmprbacv1.TempClusterRoleBinding, status tmprbacv1.BaseStatus) error {
	log := log.FromContext(ctx)
	if trb.Status.Phase != status.Phase {
		log.Info(fmt.Sprintf("Status phases are different saving status %s -> %s", trb.Status.Phase, status.Phase))
		oldPhase := trb.Status.Phase
		trb.Status = tmprbacv1.TempClusterRoleBindingStatus(status)
		if err := r.Status().Update(ctx, &trb); err != nil {
			log.Error(err, fmt.Sprintf("[TCRB] error updating TempRoleBinding status from %s to %v", oldPhase, status))
			return err
		}
	}
	return nil
}

func (r *TempClusterRoleBindingReconciler) switchFromApprovedToApplied(ctx context.Context, trb tmprbacv1.TempClusterRoleBinding, status tmprbacv1.BaseStatus, next tmprbacv1.BaseStatus) (ctrl.Result, bool, error) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusApproved && next.Phase == tmprbacv1.TempRoleBindingStatusApplied {
		// create new role bindings
		result, err := r.setTempClusterRoleBindingApplied(ctx, trb)
		return result, true, err
	}
	return ctrl.Result{}, false, nil
}

// setTempClusterRoleBindingApplied TODO: what is this?
func (r *TempClusterRoleBindingReconciler) setTempClusterRoleBindingApplied(ctx context.Context, req tmprbacv1.TempClusterRoleBinding) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info(fmt.Sprintf("approved ClusterRoleBinding Name %s", req.Name))

	var rb rbac.ClusterRoleBinding
	err := r.Get(ctx, types.NamespacedName{Name: req.Namespace}, &rb)
	if err != nil && errors.IsNotFound(err) {
		// Making new ClusterRoleBinding
		log.Info(fmt.Sprintf("CLusterRoleBindig do not exist, create one: %v", req.Name))

		reschedule, err := r.reconcileClusterRoleBinding(ctx, req)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: *reschedule}, nil
	}

	if err != nil {
		log.Error(err, "Approved but can not create ClusterRoleBinding")
	}
	return ctrl.Result{}, err
}

// reconcileRoleBinding prepare RoleBinding and duration
func (r *TempClusterRoleBindingReconciler) reconcileClusterRoleBinding(ctx context.Context, tcrb tmprbacv1.TempClusterRoleBinding) (*time.Duration, error) {
	log := log.FromContext(ctx)
	duration := tcrb.Spec.Duration.Duration

	crb := rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:        tcrb.Name,
			Labels:      tcrb.Labels,
			Annotations: tcrb.Annotations,
		},
		Subjects: tcrb.Spec.Subjects,
		RoleRef:  tcrb.Spec.RoleRef,
	}

	if err := ctrl.SetControllerReference(&tcrb, &crb, r.Scheme); err != nil {
		return nil, err
	}

	err := r.Create(ctx, &crb)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Error(err, "unable to create ClusterRoleBinding")
		return nil, err
	}

	return &duration, nil
}

func (r *TempClusterRoleBindingReconciler) deleteExternalResources(ctx context.Context, tcrb tmprbacv1.TempClusterRoleBinding) error {
	log := log.FromContext(ctx)

	var bindingList rbac.ClusterRoleBindingList
	err := r.List(ctx, &bindingList, client.MatchingFields{ownerKeyCluster: tcrb.Name})
	if err != nil {
		log.Error(err, "Error getting ClusterRoleBindings when TempClusterRoleBinding is deleted")
		return err
	}

	for _, crb := range bindingList.Items {
		log.Info("[TmpClusterRoleBinding] Deleting ClusterRoleBinding " + crb.Name)
		err = r.Delete(ctx, &crb)
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, fmt.Sprintf("[TempClusterRoleBinding] Error deleting ClusterRoleBinding for name: %s", tcrb.Name))
			return err
		}
	}
	return nil
}
