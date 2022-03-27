package v1

import "errors"

func (b *TempRoleBindingStatus) GetConditionPending() (Condition, error) {
	for _, v := range b.Conditions {
		if v.Type == TempRoleBindingStatusPending {
			return v, nil
		}
	}

	return Condition{}, errors.New("Condition Pending Missing")
}

func (b *TempRoleBindingStatus) GetConditionApplied() (Condition, error) {
	for _, v := range b.Conditions {
		if v.Type == TempRoleBindingStatusApplied {
			return v, nil
		}
	}

	return Condition{}, errors.New("Condition Applied Missing")
}
