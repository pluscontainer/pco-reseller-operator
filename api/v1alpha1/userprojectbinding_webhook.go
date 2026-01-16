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
	"context"
	"errors"
	"fmt"

	"github.com/pluscontainer/pco-reseller-operator/internal/utils"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var userprojectbindinglog = logf.Log.WithName("userprojectbinding-resource")

// SetupWebhookWithManager registers the webhook within the manager
func (r *UserProjectBinding) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		WithValidator(&UserProjectBindingCustomValidator{}).
		For(r).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:path=/validate-pco-plusserver-com-v1alpha1-userprojectbinding,mutating=false,failurePolicy=fail,sideEffects=None,groups=pco.plusserver.com,resources=userprojectbindings,verbs=create;update,versions=v1alpha1,name=vuserprojectbinding.kb.io,admissionReviewVersions=v1
type UserProjectBindingCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &UserProjectBindingCustomValidator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (v *UserProjectBindingCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	upb, ok := obj.(*UserProjectBinding)
	if !ok {
		return nil, fmt.Errorf("expected a UserProjectBinding object but got %T", obj)
	}
	userprojectbindinglog.Info("validate create", "name", upb.Name)

	if utils.IsEmpty(upb.Spec.Project) {
		return nil, errors.New(".spec.project must be specified")
	}

	if utils.IsEmpty(upb.Spec.User) {
		return nil, errors.New(".spec.user must be specified")
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (v *UserProjectBindingCustomValidator) ValidateUpdate(ctx context.Context, newObj runtime.Object, oldObj runtime.Object) (admission.Warnings, error) {
	newUpb, ok := newObj.(*UserProjectBinding)
	if !ok {
		return nil, fmt.Errorf("expected a UserProjectBinding object but got %T", newObj)
	}
	oldUpb, ok := oldObj.(*UserProjectBinding)
	if !ok {
		return nil, fmt.Errorf("expected a UserProjectBinding object but got %T", oldObj)
	}
	userprojectbindinglog.Info("validate update", "name", newUpb.Name)

	if oldUpb.Spec.Project != newUpb.Spec.Project {
		return nil, errors.New(".spec.project is immutable")
	}
	if oldUpb.Spec.User != newUpb.Spec.User {
		return nil, errors.New(".spec.user is immutable")
	}
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (v *UserProjectBindingCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	upb, ok := obj.(*UserProjectBinding)
	if !ok {
		return nil, fmt.Errorf("expected a UserProjectBinding object but got %T", obj)
	}
	userprojectbindinglog.Info("validate delete", "name", upb.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
