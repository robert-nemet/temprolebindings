package controllers

import (
	tmprbacv1 "rnemet.dev/temprolebindings/api/v1"
)

// calculateNextStatus ...
func calculateNextStatusCluster(trb tmprbacv1.TempClusterRoleBinding) (currentStatus tmprbacv1.BaseStatus, nextStatus tmprbacv1.BaseStatus) {
	currentStatus = calculateCurrentStatus(trb.Status.ToBaseStatus())

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
		next := holdOrApply(tmprbacv1.BaseSpec(trb.Spec))
		return currentStatus, newStatus(currentStatus, next)
	case tmprbacv1.TempRoleBindingStatusApproved:
		next := holdOrApply(tmprbacv1.BaseSpec(trb.Spec))
		return currentStatus, newStatus(currentStatus, next)
	case tmprbacv1.TempRoleBindingStatusApplied:
		if isTempTempRoleBindingExpired(tmprbacv1.BaseSpec(trb.Spec), trb.Status.ToBaseStatus()) {
			return currentStatus, newStatus(currentStatus, tmprbacv1.TempRoleBindingStatusExpired)
		}
		return currentStatus, currentStatus
	default: // tmprbacv1.TempRoleBindingStatusExpired | tmprbacv1.TempRoleBindingStatusDeclined
		return currentStatus, currentStatus
	}
}
