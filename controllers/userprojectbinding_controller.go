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

package controllers

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	oapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/pluscloudopen/reseller-cli/v2/pkg/openapi"
	"github.com/pluscloudopen/reseller-cli/v2/pkg/psos"
	"github.com/pluscloudopen/reseller-operator/api/v1alpha1"
	pcov1alpha1 "github.com/pluscloudopen/reseller-operator/api/v1alpha1"
	"github.com/pluscloudopen/reseller-operator/internal/utils"
)

// UserProjectBindingReconciler reconciles a UserProjectBinding object
type UserProjectBindingReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=pco.plusserver.com,resources=userprojectbindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pco.plusserver.com,resources=userprojectbindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pco.plusserver.com,resources=userprojectbindings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the UserProjectBinding object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *UserProjectBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling UserProjectBinding")

	// Fetch the UBP instance
	upb := &v1alpha1.UserProjectBinding{}
	err := r.Get(ctx, req.NamespacedName, upb)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("UBP resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get ubp.")
		return ctrl.Result{}, err
	}

	// Check if the user is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isUPBMarkedToBeDelted := upb.GetDeletionTimestamp() != nil
	if isUPBMarkedToBeDelted {
		if controllerutil.ContainsFinalizer(upb, controllerFinalizer) {
			// Run finalization logic for controllerFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeUPB(ctx, logger, *upb); err != nil {
				return ctrl.Result{}, err
			}

			// Remove controllerFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(upb, controllerFinalizer)
			err := r.Update(ctx, upb)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(upb, controllerFinalizer) {
		controllerutil.AddFinalizer(upb, controllerFinalizer)
		err = r.Update(ctx, upb)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	user := &v1alpha1.User{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: upb.Namespace,
		Name:      upb.Spec.User,
	}, user); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}

		if err := upb.UpdateUserCondition(ctx, r.Client, v1alpha1.UserNotFound, err.Error()); err != nil {
			return ctrl.Result{}, err
		}

		logger.Info(fmt.Sprintf("User %s not found, waiting for user to appear", upb.Spec.User))
		return ctrl.Result{RequeueAfter: time.Duration(3) * time.Second}, nil
	}

	if !user.IsReady() {
		if err := upb.UpdateUserCondition(ctx, r.Client, v1alpha1.UserIsUnready, "User isn't ready"); err != nil {
			return ctrl.Result{}, err
		}

		logger.Info(fmt.Sprintf("User %s isn't ready", upb.Spec.User))
		return ctrl.Result{RequeueAfter: time.Duration(3) * time.Second}, nil
	}

	//Set user condition ready
	if err := upb.UpdateUserCondition(ctx, r.Client, v1alpha1.UserIsReady, "User is ready"); err != nil {
		return ctrl.Result{}, err
	}

	userAccessSecret := &v1.Secret{}
	if err := r.Get(ctx, user.UserAccessSecretName(), userAccessSecret); err != nil {
		return ctrl.Result{}, err
	}

	project := &v1alpha1.Project{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: upb.Namespace,
		Name:      upb.Spec.Project,
	}, project); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}

		if err := upb.UpdateProjectCondition(ctx, r.Client, v1alpha1.ProjectNotFound, err.Error()); err != nil {
			return ctrl.Result{}, err
		}

		logger.Info(fmt.Sprintf("Project %s not found, waiting for project to appear", upb.Spec.Project))
		return ctrl.Result{RequeueAfter: time.Duration(3) * time.Second}, nil
	}

	if !project.IsReady() {
		if err := upb.UpdateProjectCondition(ctx, r.Client, v1alpha1.ProjectIsUnready, "Project isn't ready"); err != nil {
			return ctrl.Result{}, err
		}

		logger.Info(fmt.Sprintf("Project %s isn't ready", upb.Spec.Project))
		return ctrl.Result{RequeueAfter: time.Duration(3) * time.Second}, nil
	}

	//Set project condition ready
	if err := upb.UpdateProjectCondition(ctx, r.Client, v1alpha1.ProjectIsReady, "Project is ready"); err != nil {
		return ctrl.Result{}, err
	}

	region := &v1alpha1.Region{}
	if err := r.Get(ctx, types.NamespacedName{Name: project.Spec.Region}, region); err != nil {
		return ctrl.Result{}, err
	}

	psOsClient, err := psos.Login(region.Spec.Endpoint, region.Spec.Username, region.Spec.Password)
	if err != nil {
		return ctrl.Result{}, err
	}

	controllerIdentifier, err := utils.ControllerIdentifier(ctx, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	//Get referenced project
	openStackProject, err := utils.GetOpenStackProject(ctx, psOsClient, utils.GetOpenStackProjectName(*controllerIdentifier, types.NamespacedName{
		Namespace: project.Namespace,
		Name:      project.Name,
	}))

	if err != nil {
		return ctrl.Result{}, err
	}

	mail, err := user.Mail(ctx, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	//Get referenced user
	openStackUser, err := utils.GetOpenStackUser(ctx, psOsClient, *mail)
	if err != nil {
		if err != utils.ErrOpenStackUserNotFound {
			return ctrl.Result{}, err
		}
	}

	if openStackUser == nil {
		openStackUser, err = psOsClient.CreateUser(ctx, openapi.CreateOpenStackUser{
			Name:           oapi_types.Email(*mail),
			Description:    user.Spec.Description,
			Enabled:        user.Spec.Enabled,
			DefaultProject: &openStackProject.Id,
			Password:       string(userAccessSecret.Data[secretPasswordKey]),
		})

		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		password := string(userAccessSecret.Data[secretPasswordKey])
		openStackUser, err = psOsClient.UpdateUser(ctx, openStackUser.Id, openapi.UpdateOpenStackUser{
			Name:        (*oapi_types.Email)(mail),
			Description: &user.Spec.Description,
			Enabled:     user.Spec.Enabled,
			//DefaultProject: &openStackProject.Id, <- Don't update the default project
			Password: &password,
		})

		if err != nil {
			return ctrl.Result{}, err
		}
	}

	logger.Info(fmt.Sprintf("Ensured user %s", openStackUser.Id))

	usersInOpenStackProject, err := psOsClient.GetUsersInProject(ctx, openStackProject.Id)
	if err != nil {
		return ctrl.Result{}, err
	}

	userAlreadyAdded := false
	if usersInOpenStackProject != nil {
		for _, k := range *usersInOpenStackProject {
			if k.User == openStackUser.Id {
				userAlreadyAdded = true
				break
			}
		}
	}

	if !userAlreadyAdded {
		if err := psOsClient.AddUserToProject(ctx, openStackProject.Id, openStackUser.Id); err != nil {
			return ctrl.Result{}, err
		}

		logger.Info(fmt.Sprintf("Added user %s to project %s", openStackUser.Id, openStackProject.Id))
	}

	if upb.Spec.ApplicationCredential {
		if err := r.ensureApplicationCredential(ctx, logger, upb, *region, *openStackProject, openStackUser.Id, string(userAccessSecret.Data[secretUsernameKey]), string(userAccessSecret.Data[secretPasswordKey])); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if err := r.deprovisionApplicationCredential(ctx, logger, upb, *region, *openStackProject, openStackUser.Id, string(userAccessSecret.Data[secretUsernameKey]), string(userAccessSecret.Data[secretPasswordKey])); err != nil {
			return ctrl.Result{}, err
		}
	}

	upb.UpdateUserProjectBindingCondition(ctx, r.Client, pcov1alpha1.UserProjectBindingIsReady, "UPB ensured")

	logger.Info("Reconciling finished")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserProjectBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pcov1alpha1.UserProjectBinding{}).
		Complete(r)
}
