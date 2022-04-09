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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var temprolebindinglog = logf.Log.WithName("temprolebinding-resource")

func (r *TempRoleBinding) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-tmprbac-rnemet-dev-v1-temprolebinding,mutating=true,failurePolicy=fail,sideEffects=None,groups=tmprbac.rnemet.dev,resources=temprolebindings,verbs=create;update,versions=v1,name=mtemprolebinding.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &TempRoleBinding{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *TempRoleBinding) Default() {
	temprolebindinglog.Info("default", "name", r.Name)

	if len(r.ObjectMeta.Annotations) == 0 {
		r.ObjectMeta.Annotations = make(map[string]string)

		r.ObjectMeta.Annotations[VersionAnnotation] = "v1"
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-tmprbac-rnemet-dev-v1-temprolebinding,mutating=false,failurePolicy=fail,sideEffects=None,groups=tmprbac.rnemet.dev,resources=temprolebindings,verbs=create;update,versions=v1,name=vtemprolebinding.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &TempRoleBinding{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *TempRoleBinding) ValidateCreate() error {
	temprolebindinglog.Info("validate create", "name", r.Name)

	// // validate duration
	// if _, err := time.ParseDuration(r.Spec.Duration); err != nil {
	// 	return errors.Wrap(err, fmt.Sprintf("TempRoleBinding %s invalid duration", r.Name))
	// }

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *TempRoleBinding) ValidateUpdate(old runtime.Object) error {
	temprolebindinglog.Info("validate update", "name", r.Name)

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *TempRoleBinding) ValidateDelete() error {
	temprolebindinglog.Info("validate delete", "name", r.Name)

	return nil
}
