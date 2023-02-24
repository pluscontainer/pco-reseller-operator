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

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RegionSpec defines the desired state of Region
type RegionSpec struct {
	Endpoint string `json:"endpoint,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// RegionStatus defines the observed state of Region
type RegionStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

func (r *Region) IsReady() bool {
	return meta.IsStatusConditionTrue(r.Status.Conditions, string(RegionReady))
}

func (r *Region) UpdateRegionCondition(ctx context.Context, reconcileClient client.Client, reason RegionReadyReasons, message string) error {
	return r.updateCondition(ctx, reconcileClient, string(RegionReady), reason.regionStatus(), string(reason), message)
}

func (r *Region) updateCondition(ctx context.Context, reconcileClient client.Client, typeString string, status v1.ConditionStatus, reason string, message string) error {
	oldRegion := r.DeepCopy()

	meta.SetStatusCondition(&r.Status.Conditions, v1.Condition{
		Type:    typeString,
		Status:  status,
		Reason:  reason,
		Message: message,
	})

	return reconcileClient.Status().Patch(ctx, r, client.MergeFrom(oldRegion))
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// Region is the Schema for the regions API
type Region struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RegionSpec   `json:"spec,omitempty"`
	Status RegionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RegionList contains a list of Region
type RegionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Region `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Region{}, &RegionList{})
}
