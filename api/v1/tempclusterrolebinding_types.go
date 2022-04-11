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
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TempClusterRoleBindingSpec defines the desired state of TempClusterRoleBinding
type TempClusterRoleBindingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

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

// TempClusterRoleBindingStatus defines the observed state of TempClusterRoleBinding
type TempClusterRoleBindingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions, transition list of an object
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
	// Phase, the latest phase, final state of object
	Phase RoleBindingStatus `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TempClusterRoleBinding is the Schema for the tempclusterrolebindings API
type TempClusterRoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TempClusterRoleBindingSpec   `json:"spec,omitempty"`
	Status TempClusterRoleBindingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TempClusterRoleBindingList contains a list of TempClusterRoleBinding
type TempClusterRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TempClusterRoleBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TempClusterRoleBinding{}, &TempClusterRoleBindingList{})
}
