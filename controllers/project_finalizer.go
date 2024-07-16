package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pluscontainer/pco-reseller-cli/pkg/psos"
	"github.com/pluscontainer/pco-reseller-operator/api/v1alpha1"
	"github.com/pluscontainer/pco-reseller-operator/internal/utils"
	k8errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var errUserStillReferencingProject = errors.New("users are still referencing project")

func (r *ProjectReconciler) finalizeProject(ctx context.Context, logger logr.Logger, project v1alpha1.Project) error {
	//We wanna keep the finalizer as simple as possible
	//Don't check region availablility etc...

	//Check if users (UPB) still reference the project
	userProjectBindings := &v1alpha1.UserProjectBindingList{}
	if err := r.List(ctx, userProjectBindings, &client.ListOptions{Namespace: project.Namespace}); err != nil {
		return err
	}

	for _, k := range userProjectBindings.Items {
		if k.Spec.Project == project.Name {
			logger.Info("Users are still referencing project. Blocking deletion")
			return errUserStillReferencingProject
		}
	}

	region := &v1alpha1.Region{}
	if err := r.Get(ctx, types.NamespacedName{Name: project.Spec.Region}, region); err != nil {
		reason := v1alpha1.RegionUnknown
		if k8errors.IsNotFound(err) {
			reason = v1alpha1.RegionNotFound
		}

		if err := project.UpdateRegionCondition(ctx, r.Client, reason, err.Error()); err != nil {
			return err
		}

		return err
	}

	psOsClient, err := psos.Login(region.Spec.Endpoint, region.Spec.Username, region.Spec.Password)
	if err != nil {
		//Set region condition unready if login fails
		if err := project.UpdateRegionCondition(ctx, r.Client, v1alpha1.RegionUnknown, err.Error()); err != nil {
			return err
		}

		return err
	}

	controllerIdentifier, err := utils.ControllerIdentifier(ctx, r.Client)
	if err != nil {
		return err
	}

	openStackProjectName := utils.GetOpenStackProjectName(*controllerIdentifier, types.NamespacedName{
		Namespace: project.Namespace,
		Name:      project.Name,
	})

	openStackProject, err := utils.GetOpenStackProject(ctx, psOsClient, openStackProjectName)
	if err != nil {
		if err != utils.ErrOpenStackProjectNotFound {
			return err
		}

		//If OpenStack Project is already gone -> Done
		logger.Info(fmt.Sprintf("OpenStack Project %s already gone", openStackProjectName))
		return nil
	}

	if err := psOsClient.DeleteProject(ctx, openStackProject.Id); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("OpenStack Project %s deleted", openStackProject.Id))
	return nil
}
