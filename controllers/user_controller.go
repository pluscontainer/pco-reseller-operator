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

package controllers

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/plusserver/pluscloudopen-reseller-operator/api/v1alpha1"
	pcov1alpha1 "github.com/plusserver/pluscloudopen-reseller-operator/api/v1alpha1"
	"github.com/sethvargo/go-password/password"
)

// UserReconciler reconciles a User object
type UserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const secretUsernameKey = "username"
const secretPasswordKey = "password"

//+kubebuilder:rbac:groups=pco.plusserver.com,resources=users,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pco.plusserver.com,resources=users/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pco.plusserver.com,resources=users/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the User object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling User")

	// Fetch the User instance
	user := &v1alpha1.User{}
	err := r.Get(ctx, req.NamespacedName, user)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("User resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get user.")
		return ctrl.Result{}, err
	}

	// Check if the user is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isUserMarkedToBeDelted := user.GetDeletionTimestamp() != nil
	if isUserMarkedToBeDelted {
		if controllerutil.ContainsFinalizer(user, controllerFinalizer) {
			// Run finalization logic for controllerFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeUser(ctx, logger, *user); err != nil {
				if err != errUserProjectBindingPresent {
					return ctrl.Result{}, err
				}

				//UserProjectBindings need to be deleted by controller
				logger.Info("Waiting for all UserProjectBindings to be deleted")
				return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, nil
			}

			// Remove controllerFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(user, controllerFinalizer)
			err := r.Update(ctx, user)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(user, controllerFinalizer) {
		controllerutil.AddFinalizer(user, controllerFinalizer)
		err = r.Update(ctx, user)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	accessSecretName := user.UserAccessSecretName()

	//Ensure secret with access information
	accessSecret := &v1.Secret{}
	err = r.Get(ctx, accessSecretName, accessSecret)

	if err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}

		//As we need to generate the secret -> Update status to not ready
		if err := user.UpdateUserCondition(ctx, r.Client, pcov1alpha1.UserHasNoSecret, fmt.Sprintf("Secret %s missing", accessSecretName.Name)); err != nil {
			return ctrl.Result{}, err
		}

		isTrue := true

		generatedPassword, err := password.Generate(32, 5, 0, false, false)
		if err != nil {
			return ctrl.Result{}, err
		}

		mail, err := user.Mail(ctx, r.Client)
		if err != nil {
			return ctrl.Result{}, err
		}

		accessSecret = &v1.Secret{
			ObjectMeta: meta_v1.ObjectMeta{
				Namespace: accessSecretName.Namespace,
				Name:      accessSecretName.Name,
			},
			Immutable: &isTrue,
			StringData: map[string]string{
				secretUsernameKey: *mail,
				secretPasswordKey: generatedPassword,
			},
		}

		if err := r.Create(ctx, accessSecret); err != nil {
			return ctrl.Result{}, err
		}

		logger.Info("User secret created")
	}

	if err := user.UpdateUserCondition(ctx, r.Client, pcov1alpha1.UserIsReady, fmt.Sprintf("Secret %s ensured", accessSecretName.Name)); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling finished")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pcov1alpha1.User{}).
		Complete(r)
}
