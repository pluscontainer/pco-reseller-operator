/*
Copyright © 2023 PlusServer GmbH

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

package v1alpha1

import (
	"errors"

	"github.com/pluscontainer/reseller-operator/internal/utils"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var userprojectbindinglog = logf.Log.WithName("userprojectbinding-resource")

func (r *UserProjectBinding) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-pco-plusserver-com-v1alpha1-userprojectbinding,mutating=false,failurePolicy=fail,sideEffects=None,groups=pco.plusserver.com,resources=userprojectbindings,verbs=create;update,versions=v1alpha1,name=vuserprojectbinding.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &UserProjectBinding{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *UserProjectBinding) ValidateCreate() error {
	userprojectbindinglog.Info("validate create", "name", r.Name)

	if utils.IsEmpty(r.Spec.Project) {
		return errors.New(".spec.project must be specified")
	}

	if utils.IsEmpty(r.Spec.User) {
		return errors.New(".spec.user must be specified")
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *UserProjectBinding) ValidateUpdate(old runtime.Object) error {
	userprojectbindinglog.Info("validate update", "name", r.Name)
	oldUser := old.(*UserProjectBinding)

	if oldUser.Spec.Project != r.Spec.Project {
		return errors.New(".spec.project is immutable")
	}
	if oldUser.Spec.User != r.Spec.User {
		return errors.New(".spec.user is immutable")
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *UserProjectBinding) ValidateDelete() error {
	userprojectbindinglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
