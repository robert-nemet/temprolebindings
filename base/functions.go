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
package base

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	tmprbacv1 "rnemet.dev/temprolebindings/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func IsPhaseSet(t tmprbacv1.BaseStatus) bool {
	return len(t.Phase) > 0
}

func MakeDefaultStatus() tmprbacv1.BaseStatus {
	return tmprbacv1.BaseStatus{
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
			{
				Status: false,
				Type:   tmprbacv1.TempRoleBindingStatusHold,
			},
		},
		Phase: tmprbacv1.TempRoleBindingStatusPending,
	}
}

// newStatus, make new status
func NewStatus(currentStatus tmprbacv1.BaseStatus, next tmprbacv1.RoleBindingStatus) tmprbacv1.BaseStatus {
	for i, v := range currentStatus.Conditions {
		if v.Type == next {
			currentStatus.Conditions[i].Status = true
			currentStatus.Conditions[i].TransitionTime = metav1.Now()
			currentStatus.Phase = next
			return currentStatus
		}
	}
	return currentStatus
}

// calculateCurrentStatus for TempRoleBindingStatus
func CalculateCurrentStatus(status tmprbacv1.BaseStatus) tmprbacv1.BaseStatus {
	if IsPhaseSet(status) {
		return status
	}
	return MakeDefaultStatus()
}

// switchFromPendingToDeclined execute transation from Pending to Declined
func SwitchFromPendingToDeclined(status tmprbacv1.BaseStatus, next tmprbacv1.BaseStatus) (ctrl.Result, bool) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusPending && next.Phase == tmprbacv1.TempRoleBindingStatusDeclined {
		return ctrl.Result{}, true
	}
	return ctrl.Result{}, false
}

func SwitchFromPendingToApproved(status tmprbacv1.BaseStatus, next tmprbacv1.BaseStatus) (ctrl.Result, bool) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusPending && next.Phase == tmprbacv1.TempRoleBindingStatusApproved {
		return ctrl.Result{}, true
	}
	return ctrl.Result{}, false
}

func SwitchFromApprovedToHold(status tmprbacv1.BaseStatus, next tmprbacv1.BaseStatus) (ctrl.Result, bool) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusApproved && next.Phase == tmprbacv1.TempRoleBindingStatusHold {
		return ctrl.Result{}, true
	}
	return ctrl.Result{}, false
}

func SwitchFromAppliedToExpired(status tmprbacv1.BaseStatus, next tmprbacv1.BaseStatus) (ctrl.Result, bool) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusApplied && next.Phase == tmprbacv1.TempRoleBindingStatusExpired {
		return ctrl.Result{}, true
	}
	return ctrl.Result{}, false
}

// holdOrApply, decide to hold or apply
func HoldOrApply(spec tmprbacv1.BaseSpec) tmprbacv1.RoleBindingStatus {
	if (spec.Duration == metav1.Duration{}) && metav1.Now().Local().Before(spec.StartStop.From.Time) {
		return tmprbacv1.TempRoleBindingStatusHold
	}
	return tmprbacv1.TempRoleBindingStatusApplied
}

// isTempTempRoleBindingExpired validates expiration
func IsTempTempRoleBindingExpired(spec tmprbacv1.BaseSpec, status tmprbacv1.BaseStatus) bool {

	if status.Phase != tmprbacv1.TempRoleBindingStatusApplied {
		return false
	}

	appliedCondition, _ := status.GetConditionApplied()
	lastTimeChecked := appliedCondition.TransitionTime
	duration := spec.Duration.Duration
	if duration > 0 {
		if lastTimeChecked.Add(duration).Before(time.Now()) {
			return true
		}
		return false
	}
	return metav1.Now().After(spec.StartStop.To.Time)
}

func IsApprovalRequired(annotations map[string]string) bool {
	_, ok := annotations[tmprbacv1.StatusAnnotation]
	return ok
}

func GetStatusFromAnnotations(annotations map[string]string) tmprbacv1.RoleBindingStatus {
	status, ok := annotations[tmprbacv1.StatusAnnotation]
	if ok {
		return tmprbacv1.RoleBindingStatus(status)
	}
	return tmprbacv1.TempRoleBindingStatusNotDefined
}

// GetCurrentAndNextStatus ...
func GetCurrentAndNextStatus(status tmprbacv1.BaseStatus, annotations map[string]string, spec tmprbacv1.BaseSpec) (currentStatus tmprbacv1.BaseStatus, nextStatus tmprbacv1.BaseStatus) {
	currentStatus = CalculateCurrentStatus(status)

	switch currentStatus.Phase {
	case tmprbacv1.TempRoleBindingStatusPending:
		if spec.ApprovalRequired {
			if IsApprovalRequired(annotations) {
				// set one from annotation
				next := GetStatusFromAnnotations(annotations)
				return currentStatus, NewStatus(currentStatus, next)
			}
			// status not updated by annotation
			return currentStatus, currentStatus
		}
		return currentStatus, NewStatus(currentStatus, tmprbacv1.TempRoleBindingStatusApproved)
	case tmprbacv1.TempRoleBindingStatusHold:
		next := HoldOrApply(spec)
		return currentStatus, NewStatus(currentStatus, next)
	case tmprbacv1.TempRoleBindingStatusApproved:
		next := HoldOrApply(spec)
		return currentStatus, NewStatus(currentStatus, next)
	case tmprbacv1.TempRoleBindingStatusApplied:
		if IsTempTempRoleBindingExpired(spec, status) {
			return currentStatus, NewStatus(currentStatus, tmprbacv1.TempRoleBindingStatusExpired)
		}
		return currentStatus, currentStatus
	default: // tmprbacv1.TempRoleBindingStatusExpired | tmprbacv1.TempRoleBindingStatusDeclined
		return currentStatus, currentStatus
	}
}
