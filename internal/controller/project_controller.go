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

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/pluscontainer/pco-reseller-cli/pkg/openapi"
	"github.com/pluscontainer/pco-reseller-cli/pkg/psos"
	pcov1alpha1 "github.com/pluscontainer/pco-reseller-operator/api/v1alpha1"
	"github.com/pluscontainer/pco-reseller-operator/internal/utils"
)

// ProjectReconciler reconciles a Project object
type ProjectReconciler struct {
	client.Client

	Scheme *runtime.Scheme
}

const regionTimeout = 1 * time.Minute

//+kubebuilder:rbac:groups=pco.plusserver.com,resources=projects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pco.plusserver.com,resources=projects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pco.plusserver.com,resources=projects/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Project object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Project")

	// Fetch the Memcached instance
	project := &pcov1alpha1.Project{}
	err := r.Get(ctx, req.NamespacedName, project)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("Project resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get project.")
		return ctrl.Result{}, err
	}

	// Check if the region is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isRegionMarkedToBeDelted := project.GetDeletionTimestamp() != nil
	if isRegionMarkedToBeDelted {
		if controllerutil.ContainsFinalizer(project, controllerFinalizer) {
			// Run finalization logic for controllerFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeProject(ctx, logger, *project); err != nil {
				if err != errUserStillReferencingProject {
					return ctrl.Result{}, err
				}

				//Users are still referencing project
				return ctrl.Result{RequeueAfter: time.Duration(3) * time.Second}, nil
			}

			// Remove controllerFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(project, controllerFinalizer)
			err := r.Update(ctx, project)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(project, controllerFinalizer) {
		controllerutil.AddFinalizer(project, controllerFinalizer)
		err = r.Update(ctx, project)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	//Initialize condition fields
	if err := r.initializeConditions(ctx, project); err != nil {
		return ctrl.Result{}, err
	}

	region := &pcov1alpha1.Region{}
	if err := r.Get(ctx, types.NamespacedName{Name: project.Spec.Region}, region); err != nil {
		reason := pcov1alpha1.RegionUnknown
		if errors.IsNotFound(err) {
			reason = pcov1alpha1.RegionNotFound
		}

		if err := project.UpdateRegionCondition(ctx, r.Client, reason, err.Error()); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, err
	}

	//Check if region is ready

	if awaitErr := region.AwaitReady(ctx, regionTimeout, r.Client, logger); awaitErr != nil {
		if err := project.UpdateRegionCondition(ctx, r.Client, pcov1alpha1.RegionIsUnready, "Referenced region isn't ready"); err != nil {
			return ctrl.Result{}, err
		}

		//Wait for region to become ready -> Requeue request
		return ctrl.Result{}, awaitErr
	}

	if err := project.UpdateRegionCondition(ctx, r.Client, pcov1alpha1.RegionIsReady, fmt.Sprintf("Region %s is ready", region.Name)); err != nil {
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

	openStackProjectName := utils.GetOpenStackProjectName(*controllerIdentifier, req.NamespacedName)

	//Check if project exists
	openStackProject, err := utils.GetOpenStackProject(ctx, psOsClient, openStackProjectName)
	if err != nil {
		if err != utils.ErrOpenStackProjectNotFound {
			return ctrl.Result{}, err
		}
	}

	//Create project if it isn't present
	if openStackProject == nil {
		isFalse := false
		isTrue := true

		openStackProject, err = psOsClient.CreateProject(ctx, openapi.ProjectCreate{
			Name:                openStackProjectName,
			Description:         project.Spec.Description,
			Enabled:             &isTrue,
			NetworkPreconfigure: &isFalse,
		})

		if err != nil {
			if err := r.setProjectReadyError(ctx, project, err); err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}
	} else {
		isTrue := true

		//Update the project within openstack, if it already exists, to reflect the current state of the kubernetes object
		//Check if the project within openstack matches the kubernetes object
		projectMatches := true
		if openStackProject.Name != openStackProjectName {
			projectMatches = false
		}
		if openStackProject.Description != project.Spec.Description {
			projectMatches = false
		}
		if openStackProject.Enabled != &isTrue {
			projectMatches = false
		}
		if !projectMatches {
			openStackProject, err = psOsClient.UpdateProject(ctx, openStackProject.Id, openapi.ProjectUpdate{
				Name:        &openStackProjectName,
				Description: &project.Spec.Description,
				Enabled:     &isTrue,
			})
			if err != nil {
				if err := r.setProjectReadyError(ctx, project, err); err != nil {
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, err
			}
		}
	}

	//At this stage, the current representation of the openstack project is stored in openStackProject
	logger.Info(fmt.Sprintf("OpenStack Project %s ensured", openStackProject.Id))

	//The local v1alpha1.QuotaCollection matches the openapi definition in openapi.UpdateQuota -> We use json marshalling to convert the objects
	jsonBytes, err := json.Marshal(project.Spec.Quotas)
	if err != nil {
		if err := r.setProjectReadyError(ctx, project, err); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, err
	}

	projectQuota := openapi.UpdateQuota{}
	if err := json.Unmarshal(jsonBytes, &projectQuota); err != nil {
		if err := r.setProjectReadyError(ctx, project, err); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, err
	}

	currentProjectQuota, err := psOsClient.GetProjectQuota(ctx, openStackProject.Id)
	if err != nil {
		return ctrl.Result{}, err
	}

	//Ensure quotas are set correctly
	if projectQuota != *currentProjectQuota {
		if _, err := psOsClient.UpdateProjectQuota(ctx, openStackProject.Id, projectQuota); err != nil {
			if err := r.setProjectReadyError(ctx, project, err); err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}
	}

	logger.Info(fmt.Sprintf("Quota for project %s ensured", openStackProject.Id))

	if err := project.UpdateProjectCondition(ctx, r.Client, pcov1alpha1.ProjectIsReady, fmt.Sprintf("Project %s ensured", openStackProject.Id)); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling finished, will reconcile again in 1h")
	return ctrl.Result{RequeueAfter: 1 * time.Hour}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 5,
		}).
		WithEventFilter(pred).
		For(&pcov1alpha1.Project{}).
		Complete(r)
}
