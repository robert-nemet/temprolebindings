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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RoleBindingStatus string

const (
	TempRoleBindingStatusPending  RoleBindingStatus = "Pending"
	TempRoleBindingStatusApproved RoleBindingStatus = "Approved"
	TempRoleBindingStatusHold     RoleBindingStatus = "Hold"
	TempRoleBindingStatusApplied  RoleBindingStatus = "Applied"
	TempRoleBindingStatusExpired  RoleBindingStatus = "Expired"
	TempRoleBindingStatusDeclined RoleBindingStatus = "Declined"
	TempRoleBindingStatusError    RoleBindingStatus = "Error"

	VersionAnnotation = "tmprbac/version"
	StatusAnnotation  = "tmprbac/status"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TempRoleBindingSpec defines the desired state of TempRoleBinding
type TempRoleBindingSpec BaseSpec

// TempRoleBindingStatus defines the observed state of TempRoleBinding
type TempRoleBindingStatus BaseStatus

// Condition is definition of object status
type Condition struct {
	// TransitionTime when transtion is executed
	TransitionTime metav1.Time `json:"transitionTime,omitempty"`
	// Status if condition is met True or False
	Status bool `json:"status"`
	// Message explanatin for condition
	Message string `json:"message,omitempty"`
	// Type statu stype
	Type RoleBindingStatus `json:"type"`
}

// StartStop specify when TRB is active, time format is RFC3339
type StartStop struct {
	From metav1.Time `json:"from,omitempty"`
	To   metav1.Time `json:"to,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TempRoleBinding is the Schema for the temprolebindings API
type TempRoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TempRoleBindingSpec   `json:"spec,omitempty"`
	Status TempRoleBindingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TempRoleBindingList contains a list of TempRoleBinding
type TempRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TempRoleBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TempRoleBinding{}, &TempRoleBindingList{})
}
