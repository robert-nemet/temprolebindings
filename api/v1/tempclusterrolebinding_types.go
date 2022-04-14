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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TempClusterRoleBindingSpec defines the desired state of TempClusterRoleBinding
type TempClusterRoleBindingSpec BaseSpec

// TempClusterRoleBindingStatus defines the observed state of TempClusterRoleBinding
type TempClusterRoleBindingStatus BaseStatus

//+kubebuilder:resource:scope=Cluster
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
