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
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tmprbacv1 "rnemet.dev/temprolebindings/api/v1"
)

// TempRoleBindingReconciler reconciles a TempRoleBinding object
type TempRoleBindingReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tmprbac.rnemet.dev,resources=temprolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tmprbac.rnemet.dev,resources=temprolebindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tmprbac.rnemet.dev,resources=temprolebindings/finalizers,verbs=update
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Modify the Reconcile function to compare the state specified by
// the TempRoleBinding object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *TempRoleBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var tempRoleBinding tmprbacv1.TempRoleBinding

	if err := r.Get(ctx, req.NamespacedName, &tempRoleBinding); err != nil {
		log.Info("[TRB] cound not find TempRoleBinding")
		// Do deletion in finaliser
		// if errors.IsNotFound(err) {
		// 	log.Info("[TempRoleBinding] TempRoleBinding is deleted, delete RoleBinding")
		// 	return r.cleanRoleBinding(ctx, req.NamespacedName)
		// }
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// get current TempRoleBindingStatus
	currentStatus := r.calculateCurrentStatus(tempRoleBinding)
	// calculate next status
	nextStatus := r.calculateNextStatus(tempRoleBinding)
	// excute translation to next status
	result, err := r.executeTransition(ctx, tempRoleBinding, currentStatus, nextStatus)
	if err != nil {
		log.Error(err, "[TRB] cound not find TempRoleBinding")
		return ctrl.Result{}, err
	}
	// save status
	if err := r.reconcileStatus(ctx, tempRoleBinding, nextStatus); err != nil {
		log.Error(err, "[TRB] cound not save Status")
		return ctrl.Result{}, err
	}

	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TempRoleBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmprbacv1.TempRoleBinding{}).
		Complete(r)
}

// calculateCurrentStatus for TempRoleBindingStatus
func (r *TempRoleBindingReconciler) calculateCurrentStatus(trb tmprbacv1.TempRoleBinding) tmprbacv1.TempRoleBindingStatus {
	if len(trb.Status.Phase) == 0 {
		return tmprbacv1.TempRoleBindingStatus{
			Conditions: []tmprbacv1.Condition{
				{
					TransitionTime: metav1.Now(),
					Status:         true,
					Type:           tmprbacv1.TempRoleBindingStatusPending,
				},
				{
					Status: false,
					Type:   tmprbacv1.TempRoleBindingStatusDeclined,
				},
				{
					Status: false,
					Type:   tmprbacv1.TempRoleBindingStatusApproved,
				},
				{
					Status: false,
					Type:   tmprbacv1.TempRoleBindingStatusApplied,
				},
				{
					Status: false,
					Type:   tmprbacv1.TempRoleBindingStatusExpired,
				},
				{
					Status: false,
					Type:   tmprbacv1.TempRoleBindingStatusError,
				},
			},
			Phase: tmprbacv1.TempRoleBindingStatusPending,
		}
	}
	return trb.Status
}

// calculateNextStatus ...
func (r *TempRoleBindingReconciler) calculateNextStatus(trb tmprbacv1.TempRoleBinding) tmprbacv1.TempRoleBindingStatus {
	curentStatus := r.calculateCurrentStatus(trb)
	if curentStatus.Phase == tmprbacv1.TempRoleBindingStatusExpired {
		return curentStatus
	}
	if curentStatus.Phase == tmprbacv1.TempRoleBindingStatusApplied {
		if r.isTempTempRoleBindingExpired(trb) {
			next := tmprbacv1.TempRoleBindingStatusExpired
			for i, v := range curentStatus.Conditions {
				if v.Type == next {
					curentStatus.Conditions[i].Status = true
					curentStatus.Conditions[i].TransitionTime = metav1.Now()
					curentStatus.Phase = next
					return curentStatus
				}
			}
		}
		return curentStatus
	}
	if curentStatus.Phase == tmprbacv1.TempRoleBindingStatusApproved {
		// next should be applied
		next := tmprbacv1.TempRoleBindingStatusApplied
		for i, v := range curentStatus.Conditions {
			if v.Type == next {
				curentStatus.Conditions[i].Status = true
				curentStatus.Conditions[i].TransitionTime = metav1.Now()
				curentStatus.Phase = next
				return curentStatus
			}
		}
		return curentStatus
	}
	if status, ok := trb.Annotations[tmprbacv1.StatusAnnotation]; ok {
		// set one from annotation
		next := tmprbacv1.RoleBindingStatus(status)
		for i, v := range curentStatus.Conditions {
			if v.Type == next {
				curentStatus.Conditions[i].Status = true
				curentStatus.Conditions[i].TransitionTime = metav1.Now()
				curentStatus.Phase = next
				return curentStatus
			}
		}
	}
	return curentStatus
}

// reconcileStatus save TempRoleBindingStatus
func (r *TempRoleBindingReconciler) reconcileStatus(ctx context.Context, trb tmprbacv1.TempRoleBinding, status tmprbacv1.TempRoleBindingStatus) error {
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

// executeTransition execute transition from one to another status
func (r *TempRoleBindingReconciler) executeTransition(ctx context.Context, trb tmprbacv1.TempRoleBinding, status tmprbacv1.TempRoleBindingStatus, next tmprbacv1.TempRoleBindingStatus) (ctrl.Result, error) {

	log := log.FromContext(ctx)
	log.Info(fmt.Sprintf("[TRB] Executing transaction from %s to %s", status.Phase, next.Phase))
	// nothing happen
	if status.Phase == next.Phase {
		log.Info("[TRB] Current and next Phase are the same. No Transition.")
		return ctrl.Result{}, nil
	}

	result, ok, err := r.switchFromPendingToApproved(ctx, trb, status, next)
	if ok {
		return result, err
	}

	result, ok, err = r.switchFromPendingToDeclined(ctx, trb, status, next)
	if ok {
		return result, err
	}

	result, ok, err = r.switchFromAppliedToExpired(ctx, trb, status, next)
	if ok {
		return result, err
	}

	result, ok, err = r.switchFromApprovedToApplied(ctx, trb, status, next)
	if ok {
		return result, err
	}

	return ctrl.Result{}, errors.NewBadRequest(fmt.Sprintf("Invalid Transition for TempRoleBinding from %v to %v", status.Phase, next.Phase))
}

// switchFromPendingToDeclined execute transation from Pending to Declined
func (r *TempRoleBindingReconciler) switchFromPendingToDeclined(ctx context.Context, trb tmprbacv1.TempRoleBinding, status tmprbacv1.TempRoleBindingStatus, next tmprbacv1.TempRoleBindingStatus) (ctrl.Result, bool, error) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusPending && next.Phase == tmprbacv1.TempRoleBindingStatusDeclined {
		return ctrl.Result{}, true, nil
	}
	return ctrl.Result{}, false, nil
}

func (r *TempRoleBindingReconciler) switchFromPendingToApproved(ctx context.Context, trb tmprbacv1.TempRoleBinding, status tmprbacv1.TempRoleBindingStatus, next tmprbacv1.TempRoleBindingStatus) (ctrl.Result, bool, error) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusPending && next.Phase == tmprbacv1.TempRoleBindingStatusApproved {
		return ctrl.Result{}, true, nil
	}
	return ctrl.Result{}, false, nil
}

func (r *TempRoleBindingReconciler) switchFromApprovedToApplied(ctx context.Context, trb tmprbacv1.TempRoleBinding, status tmprbacv1.TempRoleBindingStatus, next tmprbacv1.TempRoleBindingStatus) (ctrl.Result, bool, error) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusApproved && next.Phase == tmprbacv1.TempRoleBindingStatusApplied {
		// create new role bindings
		result, err := r.setTempRoleBindingApproved(ctx, trb)
		return result, true, err
	}
	return ctrl.Result{}, false, nil
}

func (r *TempRoleBindingReconciler) switchFromAppliedToExpired(ctx context.Context, trb tmprbacv1.TempRoleBinding, status tmprbacv1.TempRoleBindingStatus, next tmprbacv1.TempRoleBindingStatus) (ctrl.Result, bool, error) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusApplied && next.Phase == tmprbacv1.TempRoleBindingStatusExpired {
		// delete RoleBindings
		result, err := r.cleanRoleBinding(ctx, types.NamespacedName{Namespace: trb.Namespace, Name: trb.Name})
		return result, true, err
	}
	return ctrl.Result{}, false, nil
}

func (r *TempRoleBindingReconciler) cleanRoleBinding(ctx context.Context, req types.NamespacedName) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var roleBinding rbac.RoleBinding
	err := r.Get(ctx, req, &roleBinding)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "[TmpRoleBinding] Error getting RoleBinding when TempRoleBinding is deleted")
		return ctrl.Result{}, err
	}
	err = r.Delete(ctx, &roleBinding)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "[TempRoleBinding] Error deleting RoleBinding")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// setTempRoleBindingApproved
func (r *TempRoleBindingReconciler) setTempRoleBindingApproved(ctx context.Context, req tmprbacv1.TempRoleBinding) (ctrl.Result, error) {
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
func (r *TempRoleBindingReconciler) reconcileRoleBinding(ctx context.Context, trb tmprbacv1.TempRoleBinding) (*time.Duration, error) {
	log := log.FromContext(ctx)
	duration, err := time.ParseDuration(trb.Spec.Duration)
	if err != nil {
		return nil, err
	}

	roleBinding := rbac.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:        trb.Name,
			Namespace:   trb.Namespace,
			Labels:      trb.Labels,
			Annotations: trb.Annotations,
		},
		Subjects: trb.Spec.Subjects,
		RoleRef:  trb.Spec.RoleRef,
	}

	err = r.Create(ctx, &roleBinding)
	if err != nil {
		log.Error(err, "[TempRoleBinding] unable to create RoleBinding")
		return nil, err
	}

	return &duration, nil
}

func (r *TempRoleBindingReconciler) isTempTempRoleBindingExpired(trb tmprbacv1.TempRoleBinding) bool {

	if trb.Status.Phase != tmprbacv1.TempRoleBindingStatusApplied {
		return false
	}

	appliedCondition, _ := trb.Status.GetConditionApplied()
	lastTimeChecked := appliedCondition.TransitionTime
	duration, _ := time.ParseDuration(trb.Spec.Duration)
	if lastTimeChecked.Add(duration).Before(time.Now()) {
		return true
	}
	return false
}
