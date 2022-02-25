/*
Copyright 2022.

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
	"fmt"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TempRoleBinding object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *TempRoleBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var trb tmprbacv1.TempRoleBinding

	err := r.Get(ctx, req.NamespacedName, &trb)
	if err != nil {
		if errors.IsNotFound(err) {
			// delete RoleBinding
			log.Info("[TempRoleBindig] TempRoleBinding is deleted, delete RoleBinding")
			return r.cleanRoleBinding(ctx, req)
		}
		return ctrl.Result{}, err
	}

	if err != nil {
		log.Error(err, "[TempRoleBindig] error geting TempRoleBinding")
		return ctrl.Result{}, err
	}

	switch trb.ObjectMeta.Annotations[tmprbacv1.StatusAnnotation] {
	case tmprbacv1.TempRoleBindigStatusPending:
		trb.ObjectMeta.Annotations[tmprbacv1.StatusAnnotation] = "---"
		break
	case tmprbacv1.TempRoleBindigStatusApproved:
		break
	case tmprbacv1.TempRoleBindigStatusApplied:
		break
	case tmprbacv1.TempRoleBindigStatusExpired:
		trb.Status.Status = tmprbacv1.TempRoleBindigStatusExpired
		err = r.Status().Update(ctx, &trb)
		if err != nil {
			return ctrl.Result{}, err
		}
		break
	case tmprbacv1.TempRoleBindigStatusDeclined:
		break
	}

	err = r.Update(ctx, &trb)
	if err != nil {
		return ctrl.Result{}, err
	}

	// TRB is alredy approved
	if trb.Status.Status == tmprbacv1.TempRoleBindigStatusApplied {
		log.Info("Applied->")
		lastTimeChecked := trb.Status.LastCheckTime.Time
		duration, _ := time.ParseDuration(trb.Spec.Duration)
		// TRB check if expired
		if lastTimeChecked.Add(duration).Before(time.Now()) {
			_, err = r.cleanRoleBinding(ctx, req)
			if err != nil {
				return ctrl.Result{}, err
			}

			trb.Status.Status = tmprbacv1.TempRoleBindigStatusExpired
			err = r.Status().Update(ctx, &trb)
			if err != nil {
				return ctrl.Result{}, err
			}
			log.Info("->Expired")
			return ctrl.Result{}, nil
		}
	}

	if trb.Status.Status == tmprbacv1.TempRoleBindigStatusApproved {
		log.Info("Approved->")
		// check status of trb
		var rb rbac.RoleBinding
		err = r.Get(ctx, req.NamespacedName, &rb)

		if err != nil && errors.IsNotFound(err) {
			// Making new RoleBinding
			log.Info(fmt.Sprintf("[TempRoleBindig] RoleBindig do not exist, create one: %v", req.NamespacedName))

			reschedule, err := r.createRoleBinding(ctx, trb)
			if err != nil {
				return ctrl.Result{}, err
			}

			trb.Status.Status = tmprbacv1.TempRoleBindigStatusApproved
			trb.Status.LastCheckTime = &metav1.Time{Time: trb.ObjectMeta.CreationTimestamp.Time}

			err = r.Status().Update(ctx, &trb)
			if err != nil {
				log.Error(err, "[TempRoleBonding] can not update status")
				return ctrl.Result{}, err
			}
			log.Info("->Applied")
			return ctrl.Result{RequeueAfter: *reschedule}, nil
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TempRoleBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmprbacv1.TempRoleBinding{}).
		Complete(r)
}

// createRoleBinding , create RoleBinding and duration
func (r *TempRoleBindingReconciler) createRoleBinding(ctx context.Context, trb tmprbacv1.TempRoleBinding) (*time.Duration, error) {
	log := log.FromContext(ctx)
	duration, err := time.ParseDuration(trb.Spec.Duration)
	if err != nil {
		return nil, err
	}

	roleBinding := rbac.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:        trb.Name,
			Namespace:   trb.Namespace,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
		Subjects: trb.Spec.Subjects,
		RoleRef:  trb.Spec.RoleRef,
	}

	err = r.Create(ctx, &roleBinding)
	if err != nil {
		log.Error(err, "[TempRoleBindig] unable to create RoleBinding")
		return nil, err
	}

	return &duration, nil
}

func (r *TempRoleBindingReconciler) cleanRoleBinding(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var rb rbac.RoleBinding
	err := r.Get(ctx, req.NamespacedName, &rb)
	if err != nil {
		// log.Error(err, "[TmpRoleBinding] Error getting RoleBinding when TempRoleBinding is deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	err = r.Delete(ctx, &rb)
	if err != nil {
		log.Error(err, "[TempRoleBinding] Error deleting RoleBinding")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
