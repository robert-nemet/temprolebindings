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
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"

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
//+kubebuilder:rbac:groups=rbac,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac,resources=rolebindings/status,verbs=get

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

	err := r.Get(ctx, req.NamespacedName, &tempRoleBinding)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("[TempRoleBinding] TempRoleBinding is deleted, delete RoleBinding")
			return r.cleanRoleBinding(ctx, req.NamespacedName)
		}
		return ctrl.Result{}, err
	}

	switch tempRoleBinding.ObjectMeta.Annotations[tmprbacv1.StatusAnnotation] {
	case tmprbacv1.TempRoleBindingStatusPending:
		r.reconcileStatus(ctx, tempRoleBinding, tmprbacv1.TempRoleBindingStatusPending)
		break
	case tmprbacv1.TempRoleBindingStatusApproved:
		return r.setTempRoleBindingApproved(ctx, tempRoleBinding)
	case tmprbacv1.TempRoleBindingStatusApplied:
		return r.checkTempTempRoleBindingApplied(ctx, tempRoleBinding)
	case tmprbacv1.TempRoleBindingStatusExpired:
		r.reconcileStatus(ctx, tempRoleBinding, tmprbacv1.TempRoleBindingStatusExpired)
		break
	case tmprbacv1.TempRoleBindingStatusDeclined:
		r.reconcileStatus(ctx, tempRoleBinding, tmprbacv1.TempRoleBindingStatusDeclined)
		break
	}

	return ctrl.Result{}, nil
}

func (r *TempRoleBindingReconciler) reconcileStatus(ctx context.Context, trb tmprbacv1.TempRoleBinding, status string) {
	if trb.Status.Status != status {
		log := log.FromContext(ctx)
		trb.Status.Status = status
		if err := r.Status().Update(ctx, &trb); err != nil {
			log.Error(err, fmt.Sprintf("[TempRoleBinding] error updating TempRoleBinding status from %s to %s", trb.Status.Status, status))
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *TempRoleBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmprbacv1.TempRoleBinding{}).
		Complete(r)
}

func (r *TempRoleBindingReconciler) cleanRoleBinding(ctx context.Context, req types.NamespacedName) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var roleBinding rbac.RoleBinding
	err := r.Get(ctx, req, &roleBinding)
	if err != nil {
		log.Error(err, "[TmpRoleBinding] Error getting RoleBinding when TempRoleBinding is deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	err = r.Delete(ctx, &roleBinding)
	if err != nil {
		log.Error(err, "[TempRoleBinding] Error deleting RoleBinding")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return ctrl.Result{}, nil
}

// setTempRoleBindingApproved
func (r *TempRoleBindingReconciler) setTempRoleBindingApproved(ctx context.Context, req tmprbacv1.TempRoleBinding) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info(fmt.Sprintf("[TempRoleBinding] approved Name %s Namespace %s", req.Name, req.Namespace))

	// Gate when status is changed
	if req.Status.Status != tmprbacv1.TempRoleBindingStatusPending {
		req.ObjectMeta.Annotations[tmprbacv1.StatusAnnotation] = req.Status.Status
		if err := r.Update(ctx, &req); err != nil {
			log.Error(err, "[TempRoleBinding] failed to update status annotation")
			return ctrl.Result{}, err
		}
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

		req.Status.Status = tmprbacv1.TempRoleBindingStatusApplied
		req.Status.LastCheckTime = &metav1.Time{Time: req.ObjectMeta.CreationTimestamp.Time}

		err = r.Status().Update(ctx, &req)
		if err != nil {
			log.Error(err, "[TempRoleBonding] can not update status")
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

func (r *TempRoleBindingReconciler) checkTempTempRoleBindingApplied(ctx context.Context, trb tmprbacv1.TempRoleBinding) (ctrl.Result, error) {

	if trb.Status.Status != tmprbacv1.TempRoleBindingStatusApplied {
		trb.Annotations[tmprbacv1.StatusAnnotation] = trb.Status.Status
		if err := r.Update(ctx, &trb); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	lastTimeChecked := trb.Status.LastCheckTime.Time
	duration, _ := time.ParseDuration(trb.Spec.Duration)
	if lastTimeChecked.Add(duration).Before(time.Now()) {
		_, err := r.cleanRoleBinding(ctx, types.NamespacedName{Namespace: trb.Namespace, Name: trb.Name})
		if err != nil {
			return ctrl.Result{}, err
		}

		trb.Status.Status = tmprbacv1.TempRoleBindingStatusExpired
		err = r.Status().Update(ctx, &trb)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}
