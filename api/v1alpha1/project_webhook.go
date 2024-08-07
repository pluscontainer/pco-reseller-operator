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

	"github.com/pluscontainer/pco-reseller-operator/internal/utils"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var projectlog = logf.Log.WithName("project-resource")

func (r *Project) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-pco-plusserver-com-v1alpha1-project,mutating=true,failurePolicy=fail,sideEffects=None,groups=pco.plusserver.com,resources=projects,verbs=create;update,versions=v1alpha1,name=mproject.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Project{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Project) Default() {
	projectlog.Info("default", "name", r.Name)

	if utils.IsEmpty(r.Spec.Description) {
		r.Spec.Description = "Generated by PCO Reseller Operator"
	}

	if r.Spec.Quotas == nil {
		r.Spec.Quotas = &QuotaCollection{
			Compute: defaultComputeQuota(),
			Volume:  defaultStorageQuota(),
			Network: defaultNetworkQuota(),
		}
	}

	if r.Spec.Quotas.Compute == nil {
		r.Spec.Quotas.Compute = defaultComputeQuota()
	}
	if r.Spec.Quotas.Volume == nil {
		r.Spec.Quotas.Volume = defaultStorageQuota()
	}
	if r.Spec.Quotas.Network == nil {
		r.Spec.Quotas.Network = defaultNetworkQuota()
	}

}

func defaultComputeQuota() *ComputeQuotas {
	return &ComputeQuotas{
		Instances:          64,
		Cores:              intPointer(64),
		FloatingIps:        intPointer(10),
		Ram:                intPointer(262144),
		KeyPairs:           500,
		SecurityGroups:     intPointer(500),
		SecurityGroupRules: intPointer(500),
		ServerGroups:       10,
		ServerGroupMembers: intPointer(15),
		MetadataItems:      100,
	}
}

func defaultStorageQuota() *VolumeQuotas {
	return &VolumeQuotas{
		Gigabytes:       intPointer(2000),
		Volumes:         500,
		Snapshots:       intPointer(99),
		Backups:         99,
		BackupGigabytes: 2000,
	}
}

func defaultNetworkQuota() *NetworkQuotas {
	return &NetworkQuotas{
		Floatingip:        intPointer(10),
		Network:           10,
		Router:            10,
		SecurityGroup:     500,
		SecurityGroupRule: 500,
		Subnet:            100,
	}
}

func intPointer(v int) *int {
	return &v
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-pco-plusserver-com-v1alpha1-project,mutating=false,failurePolicy=fail,sideEffects=None,groups=pco.plusserver.com,resources=projects,verbs=create;update,versions=v1alpha1,name=vproject.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Project{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Project) ValidateCreate() (admission.Warnings, error) {
	projectlog.Info("validate create", "name", r.Name)

	if utils.IsEmpty(r.Spec.Region) {
		return nil, errors.New(".spec.region must be specified")
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Project) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	projectlog.Info("validate update", "name", r.Name)

	oldProject := old.(*Project)

	if oldProject.Spec.Region != r.Spec.Region {
		return nil, errors.New("region field is immutable")
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Project) ValidateDelete() (admission.Warnings, error) {
	projectlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
