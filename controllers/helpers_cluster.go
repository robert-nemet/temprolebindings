package controllers

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	tmprbacv1 "rnemet.dev/temprolebindings/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// calculateCurrentStatus for TempRoleBindingStatus
func calculateCurrentStatusCluster(trb tmprbacv1.TempClusterRoleBinding) tmprbacv1.TempClusterRoleBindingStatus {
	if len(trb.Status.Phase) == 0 {
		return tmprbacv1.TempClusterRoleBindingStatus{
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
func newStatusCluster(currentStatus tmprbacv1.TempClusterRoleBindingStatus, next tmprbacv1.RoleBindingStatus) tmprbacv1.TempClusterRoleBindingStatus {
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
func holdOrApplyCluster(trb tmprbacv1.TempClusterRoleBinding) tmprbacv1.RoleBindingStatus {
	// d, _ := time.ParseDuration(trb.Spec.Duration)
	if (trb.Spec.Duration == metav1.Duration{}) && metav1.Now().Local().Before(trb.Spec.StartStop.From.Time) {
		return tmprbacv1.TempRoleBindingStatusHold
	}
	return tmprbacv1.TempRoleBindingStatusApplied
}

// isTempTempRoleBindingExpired validates expiration
func isTempTempRoleBindingExpiredCluster(trb tmprbacv1.TempClusterRoleBinding) bool {

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
func calculateNextStatusCluster(trb tmprbacv1.TempClusterRoleBinding) (currentStatus tmprbacv1.TempClusterRoleBindingStatus, nextStatus tmprbacv1.TempClusterRoleBindingStatus) {
	currentStatus = calculateCurrentStatusCluster(trb)

	switch currentStatus.Phase {
	case tmprbacv1.TempRoleBindingStatusPending:
		if trb.Spec.ApprovalRequired {
			if status, ok := trb.Annotations[tmprbacv1.StatusAnnotation]; ok {
				// set one from annotation
				next := tmprbacv1.RoleBindingStatus(status)
				return currentStatus, newStatusCluster(currentStatus, next)
			}
			// status not updated by annotation
			return currentStatus, currentStatus
		}
		return currentStatus, newStatusCluster(currentStatus, tmprbacv1.TempRoleBindingStatusApproved)
	case tmprbacv1.TempRoleBindingStatusHold:
		next := holdOrApplyCluster(trb)
		return currentStatus, newStatusCluster(currentStatus, next)
	case tmprbacv1.TempRoleBindingStatusApproved:
		next := holdOrApplyCluster(trb)
		return currentStatus, newStatusCluster(currentStatus, next)
	case tmprbacv1.TempRoleBindingStatusApplied:
		if isTempTempRoleBindingExpiredCluster(trb) {
			return currentStatus, newStatusCluster(currentStatus, tmprbacv1.TempRoleBindingStatusExpired)
		}
		return currentStatus, currentStatus
	default: // tmprbacv1.TempRoleBindingStatusExpired | tmprbacv1.TempRoleBindingStatusDeclined
		return currentStatus, currentStatus
	}
}

// switchFromPendingToDeclined execute transation from Pending to Declined
func switchFromPendingToDeclinedCluster(trb tmprbacv1.TempClusterRoleBinding, status tmprbacv1.TempClusterRoleBindingStatus, next tmprbacv1.TempClusterRoleBindingStatus) (ctrl.Result, bool) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusPending && next.Phase == tmprbacv1.TempRoleBindingStatusDeclined {
		return ctrl.Result{}, true
	}
	return ctrl.Result{}, false
}

func switchFromPendingToApprovedCluster(trb tmprbacv1.TempClusterRoleBinding, status tmprbacv1.TempClusterRoleBindingStatus, next tmprbacv1.TempClusterRoleBindingStatus) (ctrl.Result, bool) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusPending && next.Phase == tmprbacv1.TempRoleBindingStatusApproved {
		return ctrl.Result{}, true
	}
	return ctrl.Result{}, false
}

func switchFromApprovedToHoldCluster(trb tmprbacv1.TempClusterRoleBinding, status tmprbacv1.TempClusterRoleBindingStatus, next tmprbacv1.TempClusterRoleBindingStatus) (ctrl.Result, bool) {
	if status.Phase == tmprbacv1.TempRoleBindingStatusApproved && next.Phase == tmprbacv1.TempRoleBindingStatusHold {
		return ctrl.Result{}, true
	}
	return ctrl.Result{}, false
}
