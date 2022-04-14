package v1

import "errors"

func (b *TempClusterRoleBindingStatus) GetConditionPending() (Condition, error) {
	for _, v := range b.Conditions {
		if v.Type == TempRoleBindingStatusPending {
			return v, nil
		}
	}

	return Condition{}, errors.New("Condition Pending Missing")
}

func (b *TempClusterRoleBindingStatus) GetConditionApplied() (Condition, error) {
	for _, v := range b.Conditions {
		if v.Type == TempRoleBindingStatusApplied {
			return v, nil
		}
	}

	return Condition{}, errors.New("Condition Applied Missing")
}

func (b TempClusterRoleBindingStatus) ToBaseStatus() BaseStatus {
	return BaseStatus(b)
}
