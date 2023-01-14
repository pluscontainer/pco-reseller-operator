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
	"context"
	"fmt"

	"git.ps-intern.de/mk/gardener/pco-reseller-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// UserSpec defines the desired state of User
type UserSpec struct {
	Description string `json:"description,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

// UserStatus defines the observed state of User
type UserStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

func (r *User) IsReady() bool {
	return meta.IsStatusConditionTrue(r.Status.Conditions, string(UserReady))
}

func (r *User) UpdateUserCondition(ctx context.Context, reconcileClient client.Client, reason UserReadyReasons, message string) error {
	return r.updateCondition(ctx, reconcileClient, string(UserReady), reason.userStatus(), string(reason), message)
}

func (r *User) updateCondition(ctx context.Context, reconcileClient client.Client, typeString string, status v1.ConditionStatus, reason string, message string) error {
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

// User is the Schema for the users API
type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserSpec   `json:"spec,omitempty"`
	Status UserStatus `json:"status,omitempty"`
}

func (u User) Mail(ctx context.Context, r client.Client) (*string, error) {
	controllerId, err := utils.ControllerIdentifier(ctx, r)
	if err != nil {
		return nil, err
	}

	mail := fmt.Sprintf("%s-%s@%s.k8s", u.Name, u.Namespace, *controllerId)
	return &mail, nil
}

func (u User) UserAccessSecretName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: u.Namespace,
		Name:      fmt.Sprintf("%s-openstack", u.Name),
	}
}

func (u User) DefaultUserProjectBindingName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: u.Namespace,
		Name:      fmt.Sprintf("%s-default-project", u.Name),
	}
}

//+kubebuilder:object:root=true

// UserList contains a list of User
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

func init() {
	SchemeBuilder.Register(&User{}, &UserList{})
}
