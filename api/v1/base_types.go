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
package v1

import (
	"errors"

	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BaseStatus struct {
	// Conditions, transition list of an object
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
	// Phase, the latest phase, final state of object
	Phase RoleBindingStatus `json:"phase,omitempty"`
}

type BaseSpec struct {
	// ApprovalRequired, flag if approval is required
	ApprovalRequired bool `json:"approvalRequired,omitempty"`
	// Duration, duration of RoleBinding from the moment of creation
	Duration metav1.Duration `json:"duration,omitempty"`
	// StartStop, defines when TempRoleBinding is applied and when expires
	StartStop StartStop `json:"startStop,omitempty"`
	// Spec, RoleBinding specification
	Subjects []rbac.Subject `json:"subjects,omitempty"`
	RoleRef  rbac.RoleRef   `json:"roleRef,omitempty"`
}

func (bs BaseStatus) ToTempRoleBindingStatus() TempRoleBindingStatus {
	return TempRoleBindingStatus(bs)
}

func (b BaseStatus) GetConditionPending() (Condition, error) {
	for _, v := range b.Conditions {
		if v.Type == TempRoleBindingStatusPending {
			return v, nil
		}
	}

	return Condition{}, errors.New("Condition Pending Missing")
}

func (b BaseStatus) GetConditionApplied() (Condition, error) {
	for _, v := range b.Conditions {
		if v.Type == TempRoleBindingStatusApplied {
			return v, nil
		}
	}

	return Condition{}, errors.New("Condition Applied Missing")
}
