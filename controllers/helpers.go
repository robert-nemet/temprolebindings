package controllers

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	tmprbacv1 "rnemet.dev/temprolebindings/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// calculateCurrentStatus for TempRoleBindingStatus
func calculateCurrentStatus(trb tmprbacv1.TempRoleBinding) tmprbacv1.TempRoleBindingStatus {
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
				{
					Status: false,
					Type:   tmprbacv1.TempRoleBindingStatusHold,
				},
			},
			Phase: tmprbacv1.TempRoleBindingStatusPending,
		}
	}
	return trb.Status
}

// newStatus, make new status
func newStatus(currentStatus tmprbacv1.TempRoleBindingStatus, next tmprbacv1.RoleBindingStatus) tmprbacv1.TempRoleBindingStatus {
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

// holdOrApply, decide to hold or apply
func holdOrApply(trb tmprbacv1.TempRoleBinding) tmprbacv1.RoleBindingStatus {
	// d, _ := time.ParseDuration(trb.Spec.Duration)
	if (trb.Spec.Duration == metav1.Duration{}) && metav1.Now().Local().Before(trb.Spec.StartStop.From.Time) {
		return tmprbacv1.TempRoleBindingStatusHold
	}
	return tmprbacv1.TempRoleBindingStatusApplied
}

// isTempTempRoleBindingExpired validates expiration
func isTempTempRoleBindingExpired(trb tmprbacv1.TempRoleBinding) bool {

	if trb.Status.Phase != tmprbacv1.TempRoleBindingStatusApplied {
		return false
	}

	appliedCondition, _ := trb.Status.GetConditionApplied()
	lastTimeChecked := appliedCondition.TransitionTime
	duration := trb.Spec.Duration.Duration
	if duration > 0 {
		if lastTimeChecked.Add(duration).Before(time.Now()) {
			return true
		}
		return false
	}
	return metav1.Now().After(trb.Spec.StartStop.To.Time)
}

// calculateNextStatus ...
func calculateNextStatus(trb tmprbacv1.TempRoleBinding) (currentStatus tmprbacv1.TempRoleBindingStatus, nextStatus tmprbacv1.TempRoleBindingStatus) {
	currentStatus = calculateCurrentStatus(trb)

	switch currentStatus.Phase {
	case tmprbacv1.TempRoleBindingStatusPending:
		if trb.Spec.ApprovalRequired {
			if status, ok := trb.Annotations[tmprbacv1.StatusAnnotation]; ok {
				// set one from annotation
				next := tmprbacv1.RoleBindingStatus(status)
				return currentStatus, newStatus(currentStatus, next)
			}
			// status not updated by annotation
			return currentStatus, currentStatus
		}
		return currentStatus, newStatus(currentStatus, tmprbacv1.TempRoleBindingStatusApproved)
	case tmprbacv1.TempRoleBindingStatusHold:
		next := holdOrApply(trb)
		return currentStatus, newStatus(currentStatus, next)
	case tmprbacv1.TempRoleBindingStatusApproved:
		next := holdOrApply(trb)
		return currentStatus, newStatus(currentStatus, next)
	case tmprbacv1.TempRoleBindingStatusApplied:
		if isTempTempRoleBindingExpired(trb) {
			return currentStatus, newStatus(currentStatus, tmprbacv1.TempRoleBindingStatusExpired)
		}
		return currentStatus, currentStatus
	default: // tmprbacv1.TempRoleBindingStatusExpired | tmprbacv1.TempRoleBindingStatusDeclined
		return currentStatus, currentStatus
	}
}

// switchFromPendingToDeclined execute transation from Pending to Declined
func switchFromPendingToDeclined(trb tmprbacv1.TempRoleBinding, status tmprbacv1.TempRoleBindingStatus, next tmprbacv1.TempRoleBindingStatus) (ctrl.Result, bool) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusPending && next.Phase == tmprbacv1.TempRoleBindingStatusDeclined {
		return ctrl.Result{}, true
	}
	return ctrl.Result{}, false
}

func switchFromPendingToApproved(trb tmprbacv1.TempRoleBinding, status tmprbacv1.TempRoleBindingStatus, next tmprbacv1.TempRoleBindingStatus) (ctrl.Result, bool) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusPending && next.Phase == tmprbacv1.TempRoleBindingStatusApproved {
		return ctrl.Result{}, true
	}
	return ctrl.Result{}, false
}

func switchFromApprovedToHold(trb tmprbacv1.TempRoleBinding, status tmprbacv1.TempRoleBindingStatus, next tmprbacv1.TempRoleBindingStatus) (ctrl.Result, bool) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusApproved && next.Phase == tmprbacv1.TempRoleBindingStatusHold {
		return ctrl.Result{}, true
	}
	return ctrl.Result{}, false
}
