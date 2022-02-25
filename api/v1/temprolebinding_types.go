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

package v1

import (
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	TempRoleBindigStatusExpired  = "Expired"
	TempRoleBindigStatusApproved = "Approved"
	TempRoleBindigStatusApplied  = "Applied"
	TempRoleBindigStatusPending  = "Pending"
	TempRoleBindigStatusDeclined = "Declined"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TempRoleBindingSpec defines the desired state of TempRoleBinding
type TempRoleBindingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Duration, duration of RoleBinding from the moment of creation. It should with string indicating timeunit: s, m, h, d
	Duration string `json:"duration,omitempty"`
	// Spec, RoleBinding specification
	Subjects []rbac.Subject `json:"subjects,omitempty"`
	RoleRef  rbac.RoleRef   `json:"roleRef"`
}

// TempRoleBindingStatus defines the observed state of TempRoleBinding
type TempRoleBindingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	Status string `json:"status,omitempty"`

	// Information when was the last time the job was successfully scheduled.
	// +optional
	LastCheckTime *metav1.Time `json:"lastCheckTime,omitempty"`
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
