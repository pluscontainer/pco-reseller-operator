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
var regionlog = logf.Log.WithName("region-resource")

// SetupWebhookWithManager registers the webhook with the manager
func (r *Region) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		WithValidator(&RegionCustomValidator{}).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-pco-plusserver-com-v1alpha1-region,mutating=true,failurePolicy=fail,sideEffects=None,groups=pco.plusserver.com,resources=regions,verbs=create;update,versions=v1alpha1,name=mregion.kb.io,admissionReviewVersions=v1

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-pco-plusserver-com-v1alpha1-region,mutating=false,failurePolicy=fail,sideEffects=None,groups=pco.plusserver.com,resources=regions,verbs=create;update,versions=v1alpha1,name=vregion.kb.io,admissionReviewVersions=v1

type RegionCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &RegionCustomValidator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (v *RegionCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	region, ok := obj.(*Region)
	if !ok {
		return nil, fmt.Errorf("expected a Region object but got %T", obj)
	}
	regionlog.Info("validate create", "name", region.Name)

	if region.Spec.SecretRef == nil {
		if utils.IsEmpty(region.Spec.Endpoint) {
			return nil, errors.New("endpoint must be specified if no Secret is specified")
		}
		if utils.IsEmpty(region.Spec.Username) {
			return nil, errors.New("username must be specified if no Secret is specified")
		}
		if utils.IsEmpty(region.Spec.Password) {
			return nil, errors.New("password must be specified if no Secret is specified")
		}
	}

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (v *RegionCustomValidator) ValidateUpdate(ctx context.Context, newObj runtime.Object, oldObj runtime.Object) (admission.Warnings, error) {
	oldRegion, ok := oldObj.(*Region)
	if !ok {
		return nil, fmt.Errorf("expected a Region object but got %T", oldObj)
	}
	newRegion, ok := newObj.(*Region)
	if !ok {
		return nil, fmt.Errorf("expected a Region object but got %T", newObj)
	}
	regionlog.Info("validate update", "name", oldRegion.Name)

	if oldRegion.Spec.Endpoint != newRegion.Spec.Endpoint {
		return nil, errors.New("endpoint is immutable")
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (v *RegionCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	region, ok := obj.(*Region)
	if !ok {
		return nil, fmt.Errorf("expected a Region object but got %T", obj)
	}
	regionlog.Info("validate delete", "name", region.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
