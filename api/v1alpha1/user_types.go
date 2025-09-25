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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pluscontainer/pco-reseller-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// UserSpec defines the desired state of User
type UserSpec struct {
	// Description is a free-text field for storing information about the user
	Description string `json:"description,omitempty"`
	// Enabled represents if the user is enabled or not
	Enabled *bool `json:"enabled,omitempty"`
}

// UserStatus defines the observed state of User
type UserStatus struct {
	// Conditions store the conditions of the user object
	Conditions []metav1.Condition `json:"conditions"`
}

// AwaitReady blocks until the user is completely ready
func (v *User) AwaitReady(ctx context.Context, timeout time.Duration, client client.Client, logger logr.Logger) error {
	timeoutContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	out := make(chan bool)

	go v.awaitAllConditionsTrue(timeoutContext, client, logger, out)

	select {
	case <-out:
		return nil
	case <-timeoutContext.Done():
		return timeoutContext.Err()
	}
}

func (v *User) awaitAllConditionsTrue(ctx context.Context, client client.Client, logger logr.Logger, out chan bool) {
	for !utils.AllConditionsTrue(v.Status.Conditions) {
		//Check if the context is still good
		if ctx.Err() != nil {
			return
		}

		//Wait before refreshing object
		time.Sleep(3 * time.Second)

		//Refresh object
		if err := client.Get(ctx, types.NamespacedName{Namespace: v.Namespace, Name: v.Name}, v); err != nil {
			logger.Error(err, "Couldn't refresh conditions")
		}
	}

	out <- true
}

// UpdateUserCondition updates the given condition in the user object and patches its status subresource
func (r *User) UpdateUserCondition(ctx context.Context, reconcileClient client.Client, reason UserReadyReasons, message string) error {
	return r.updateCondition(ctx, reconcileClient, string(UserReady), reason.userStatus(), string(reason), message)
}

func (r *User) updateCondition(ctx context.Context, reconcileClient client.Client, typeString string, status metav1.ConditionStatus, reason string, message string) error {
	oldUser := r.DeepCopy()

	meta.SetStatusCondition(&r.Status.Conditions, metav1.Condition{
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

// Mail returns the e-mail address of the user
func (u User) Mail(ctx context.Context, r client.Client) (*string, error) {
	controllerId, err := utils.ControllerIdentifier(ctx, r)
	if err != nil {
		return nil, err
	}

	mail := fmt.Sprintf("%s-%s@%s.k8s", u.Name, u.Namespace, *controllerId)
	return &mail, nil
}

// UserAccessSecretName returns the secret name for the user object
func (u User) UserAccessSecretName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: u.Namespace,
		Name:      fmt.Sprintf("%s-openstack", u.Name),
	}
}

// DefaultUserProjectBindingName returns the default project for the user
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

	Items []User `json:"items"`
}

func init() {
	SchemeBuilder.Register(&User{}, &UserList{})
}
