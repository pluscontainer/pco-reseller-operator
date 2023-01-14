/*
Copyright 2023.

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

	"git.ps-intern.de/mk/gardener/pco-reseller-operator/internal/utils"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var regionlog = logf.Log.WithName("region-resource")

func (r *Region) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-pco-plusserver-com-v1alpha1-region,mutating=true,failurePolicy=fail,sideEffects=None,groups=pco.plusserver.com,resources=regions,verbs=create;update,versions=v1alpha1,name=mregion.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Region{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Region) Default() {
	projectlog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-pco-plusserver-com-v1alpha1-region,mutating=false,failurePolicy=fail,sideEffects=None,groups=pco.plusserver.com,resources=regions,verbs=create;update,versions=v1alpha1,name=vregion.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Region{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Region) ValidateCreate() error {
	regionlog.Info("validate create", "name", r.Name)

	if utils.IsEmpty(r.Spec.Endpoint) {
		return errors.New("endpoint must be specified")
	}
	if utils.IsEmpty(r.Spec.Username) {
		return errors.New("username must be specified")
	}
	if utils.IsEmpty(r.Spec.Password) {
		return errors.New("password must be specified")
	}

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Region) ValidateUpdate(old runtime.Object) error {
	regionlog.Info("validate update", "name", r.Name)

	oldRegion := old.(*Region)
	if oldRegion.Spec.Endpoint != r.Spec.Endpoint {
		return errors.New("endpoint is immutable")
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Region) ValidateDelete() error {
	regionlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
