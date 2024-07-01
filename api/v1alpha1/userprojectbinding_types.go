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
	"github.com/pluscontainer/reseller-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// UserProjectBindingSpec defines the desired state of UserProjectBinding
type UserProjectBindingSpec struct {
	Project               string `json:"project"`
	User                  string `json:"user"`
	ApplicationCredential bool   `json:"applicationCredential,omitempty"`
}

// UserProjectBindingStatus defines the observed state of UserProjectBinding
type UserProjectBindingStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

func (v *UserProjectBinding) AwaitReady(ctx context.Context, timeout time.Duration, client client.Client, logger logr.Logger) error {
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

func (v *UserProjectBinding) awaitAllConditionsTrue(ctx context.Context, client client.Client, logger logr.Logger, out chan bool) {
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

func (upb UserProjectBinding) ApplicationCredentialName() string {
	return fmt.Sprintf("%s-appcred", upb.Name)
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
