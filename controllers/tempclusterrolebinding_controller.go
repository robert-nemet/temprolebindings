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
		log.Info("[TRB] cound not find TempRoleBinding")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if tempClusterRoleBinding.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&tempClusterRoleBinding, finalizerName) {
			controllerutil.AddFinalizer(&tempClusterRoleBinding, finalizerName)
			if err := r.Update(ctx, &tempClusterRoleBinding); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(&tempClusterRoleBinding, finalizerName) {
			log.Info("Deleting external Resorces")

			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(ctx, tempClusterRoleBinding); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(&tempClusterRoleBinding, finalizerName)
			if err := r.Update(ctx, &tempClusterRoleBinding); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// calculate next status
	currentStatus, nextStatus := calculateNextStatusCluster(tempClusterRoleBinding)
	// excute translation to next status
	result, err := r.executeTransition(ctx, tempClusterRoleBinding, currentStatus, nextStatus)
	if err != nil {
		log.Error(err, "[TRB] cound not find TempRoleBinding")
		return ctrl.Result{}, err
	}
	// save status
	if err := r.reconcileStatus(ctx, tempClusterRoleBinding, nextStatus); err != nil {
		log.Error(err, "[TRB] cound not save Status")
		return ctrl.Result{}, err
	}

	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TempClusterRoleBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &rbac.ClusterRoleBinding{}, ownerKey, func(rawObj client.Object) []string {
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

func (r *TempClusterRoleBindingReconciler) deleteExternalResources(ctx context.Context, tcrb tmprbacv1.TempClusterRoleBinding) error {
	log := log.FromContext(ctx)

	var roleBindingList rbac.ClusterRoleBindingList
	err := r.List(ctx, &roleBindingList, client.InNamespace("*"), client.MatchingFields{ownerKey: tcrb.Name})
	if err != nil {
		log.Error(err, "[TmpClusterRoleBinding] Error getting ClusterRoleBindings when TempClusterRoleBinding is deleted")
		return err
	}

	for _, rb := range roleBindingList.Items {
		log.Info("[TmpClusterRoleBinding] Deleting ClusterRoleBinding " + rb.Name)
		err = r.Delete(ctx, &rb)
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, fmt.Sprintf("[TempClusterRoleBinding] Error deleting ClusterRoleBinding for name: %s", tcrb.Name))
			return err
		}
	}
	return nil
}

// executeTransition execute transition from one to another status
func (r *TempClusterRoleBindingReconciler) executeTransition(ctx context.Context, tcrb tmprbacv1.TempClusterRoleBinding, status tmprbacv1.TempClusterRoleBindingStatus, next tmprbacv1.TempClusterRoleBindingStatus) (ctrl.Result, error) {

	log := log.FromContext(ctx)
	log.Info(fmt.Sprintf("[TRB] Executing transaction from %s to %s", status.Phase, next.Phase))
	// nothing happen
	if status.Phase == next.Phase {
		log.Info("[TRB] Current and next Phase are the same. No Transition.")
		return ctrl.Result{}, nil
	}

	result, ok := switchFromPendingToApprovedCluster(tcrb, status, next)
	if ok {
		return result, nil
	}

	result, ok = switchFromPendingToDeclinedCluster(tcrb, status, next)
	if ok {
		return result, nil
	}

	result, ok, err := r.switchFromAppliedToExpired(ctx, tcrb, status, next)
	if ok {
		return result, err
	}

	result, ok = switchFromApprovedToHoldCluster(tcrb, status, next)
	if ok {
		return result, nil
	}

	result, ok, err = r.switchFromApprovedToApplied(ctx, tcrb, status, next)
	if ok {
		return result, err
	}

	return ctrl.Result{}, errors.NewBadRequest(fmt.Sprintf("Invalid Transition for TempRoleBinding from %v to %v", status.Phase, next.Phase))
}

// reconcileStatus save TempRoleBindingStatus
func (r *TempClusterRoleBindingReconciler) reconcileStatus(ctx context.Context, trb tmprbacv1.TempClusterRoleBinding, status tmprbacv1.TempClusterRoleBindingStatus) error {
	log := log.FromContext(ctx)
	if trb.Status.Phase != status.Phase {
		log.Info("Status phases are different saving status")
		oldPhase := trb.Status.Phase
		trb.Status = status
		if err := r.Status().Update(ctx, &trb); err != nil {
			log.Error(err, fmt.Sprintf("[TRB] error updating TempRoleBinding status from %s to %v", oldPhase, status))
			return err
		}
	}
	return nil
}

func (r *TempClusterRoleBindingReconciler) switchFromAppliedToExpired(ctx context.Context, trb tmprbacv1.TempClusterRoleBinding, status tmprbacv1.TempClusterRoleBindingStatus, next tmprbacv1.TempClusterRoleBindingStatus) (ctrl.Result, bool, error) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusApplied && next.Phase == tmprbacv1.TempRoleBindingStatusExpired {
		// delete RoleBindings
		err := r.deleteExternalResources(ctx, trb)
		return ctrl.Result{}, true, err
	}
	return ctrl.Result{}, false, nil
}

func (r *TempClusterRoleBindingReconciler) switchFromApprovedToApplied(ctx context.Context, trb tmprbacv1.TempClusterRoleBinding, status tmprbacv1.TempClusterRoleBindingStatus, next tmprbacv1.TempClusterRoleBindingStatus) (ctrl.Result, bool, error) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusApproved && next.Phase == tmprbacv1.TempRoleBindingStatusApplied {
		// create new role bindings
		result, err := r.setTempRoleBindingApplied(ctx, trb)
		return result, true, err
	}
	return ctrl.Result{}, false, nil
}

// setTempRoleBindingApplied TODO: what is this?
func (r *TempClusterRoleBindingReconciler) setTempRoleBindingApplied(ctx context.Context, req tmprbacv1.TempClusterRoleBinding) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info(fmt.Sprintf("[TempRoleBinding] approved Name %s Namespace %s", req.Name, req.Namespace))

	// Gate when status is changed
	if req.Status.Phase != tmprbacv1.TempRoleBindingStatusApproved {
		return ctrl.Result{}, nil
	}

	log.Info("[TempRoleBonding] Creating RobeBinding")
	var roleBinding rbac.RoleBinding
	err := r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: req.Namespace}, &roleBinding)
	if err != nil && errors.IsNotFound(err) {
		// Making new RoleBinding
		log.Info(fmt.Sprintf("[TempRoleBindig] RoleBindig do not exist, create one: %v", req.Name))

		reschedule, err := r.reconcileRoleBinding(ctx, req)
		if err != nil {
			return ctrl.Result{}, err
		}

		log.Info("[TempRoleBonding] RoleBinding ->Applied")
		return ctrl.Result{RequeueAfter: *reschedule}, nil
	}

	if err != nil {
		log.Error(err, "[TempRoleBonding] Approved but can not create RoleBinding")
	}
	return ctrl.Result{}, err
}

// reconcileRoleBinding prepare RoleBinding and duration
func (r *TempClusterRoleBindingReconciler) reconcileRoleBinding(ctx context.Context, trb tmprbacv1.TempClusterRoleBinding) (*time.Duration, error) {
	log := log.FromContext(ctx)
	duration := trb.Spec.Duration.Duration

	roleBinding := rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:        trb.Name,
			Namespace:   trb.Namespace,
			Labels:      trb.Labels,
			Annotations: trb.Annotations,
		},
		Subjects: trb.Spec.Subjects,
		RoleRef:  trb.Spec.RoleRef,
	}

	if err := ctrl.SetControllerReference(&trb, &roleBinding, r.Scheme); err != nil {
		return nil, err
	}

	err := r.Create(ctx, &roleBinding)
	if err != nil {
		log.Error(err, "[TempRoleBinding] unable to create ClusterRoleBinding")
		return nil, err
	}

	return &duration, nil
}
