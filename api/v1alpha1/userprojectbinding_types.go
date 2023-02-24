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

// UserProjectBindingSpec defines the desired state of UserProjectBinding
type UserProjectBindingSpec struct {
	Project string `json:"project,omitempty"`
	User    string `json:"user,omitempty"`
}

// UserProjectBindingStatus defines the observed state of UserProjectBinding
type UserProjectBindingStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

func (r *UserProjectBinding) IsReady() bool {
	return meta.IsStatusConditionTrue(r.Status.Conditions, string(ProjectReady)) && meta.IsStatusConditionTrue(r.Status.Conditions, string(UserReady)) && meta.IsStatusConditionTrue(r.Status.Conditions, string(UserProjectBindingReady))
}

func (r *UserProjectBinding) UpdateUserProjectBindingCondition(ctx context.Context, reconcileClient client.Client, reason UserProjectBindingReadyReasons, message string) error {
	return r.updateCondition(ctx, reconcileClient, string(UserProjectBindingReady), reason.userProjectBindingStatus(), string(reason), message)
}

func (r *UserProjectBinding) UpdateUserCondition(ctx context.Context, reconcileClient client.Client, reason UserReadyReasons, message string) error {
	return r.updateCondition(ctx, reconcileClient, string(UserReady), reason.userStatus(), string(reason), message)
}

func (r *UserProjectBinding) UpdateProjectCondition(ctx context.Context, reconcileClient client.Client, reason ProjectReadyReasons, message string) error {
	return r.updateCondition(ctx, reconcileClient, string(ProjectReady), reason.projectStatus(), string(reason), message)
}

func (r *UserProjectBinding) updateCondition(ctx context.Context, reconcileClient client.Client, typeString string, status v1.ConditionStatus, reason string, message string) error {
	oldUser := r.DeepCopy()

	meta.SetStatusCondition(&r.Status.Conditions, v1.Condition{
		Type:    typeString,
		Status:  status,
		Reason:  reason,
		Message: message,
	})

	return reconcileClient.Status().Patch(ctx, r, client.MergeFrom(oldUser))
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// UserProjectBinding is the Schema for the userprojectbindings API
type UserProjectBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserProjectBindingSpec   `json:"spec,omitempty"`
	Status UserProjectBindingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// UserProjectBindingList contains a list of UserProjectBinding
type UserProjectBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserProjectBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UserProjectBinding{}, &UserProjectBindingList{})
}
