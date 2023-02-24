package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pluscloudopen/reseller-cli/v2/pkg/psos"
	"github.com/pluscloudopen/reseller-operator/api/v1alpha1"
	"github.com/pluscloudopen/reseller-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *UserProjectBindingReconciler) finalizeUPB(ctx context.Context, logger logr.Logger, upb v1alpha1.UserProjectBinding) error {
	user := &v1alpha1.User{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: upb.Namespace,
		Name:      upb.Spec.User,
	}, user); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("User %s not present, UPB can be deleted", upb.Spec.User))
			return nil
		}

		return err
	}

	project := &v1alpha1.Project{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: upb.Namespace,
		Name:      upb.Spec.Project,
	}, project); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("Project %s not present, UPB can be deleted", upb.Spec.Project))
			return nil
		}
		return err
	}

	region := &v1alpha1.Region{}
	if err := r.Get(ctx, types.NamespacedName{Name: project.Spec.Region}, region); err != nil {
		return err
	}

	psOsClient, err := psos.Login(region.Spec.Endpoint, region.Spec.Username, region.Spec.Password)
	if err != nil {
		return err
	}

	controllerIdentifier, err := utils.ControllerIdentifier(ctx, r.Client)
	if err != nil {
		return err
	}

	//Get referenced project
	openStackProject, err := utils.GetOpenStackProject(ctx, psOsClient, utils.GetOpenStackProjectName(*controllerIdentifier, types.NamespacedName{
		Namespace: project.Namespace,
		Name:      project.Name,
	}))
	if err != nil {
		if err == utils.ErrOpenStackProjectNotFound {
			logger.Info(fmt.Sprintf("Project %s not present within OpenStack, UPB can be deleted", upb.Spec.Project))
			return nil
		}

		return err
	}

	mail, err := user.Mail(ctx, r.Client)
	if err != nil {
		return err
	}

	//Get referenced user
	openStackUser, err := utils.GetOpenStackUser(ctx, psOsClient, *mail)
	if err != nil {
		if err == utils.ErrOpenStackUserNotFound {
			logger.Info(fmt.Sprintf("User %s not present within OpenStack, UPB can be deleted", upb.Spec.Project))
			return nil
		}

		return err
	}

	if err := psOsClient.RemoveUserFromProject(ctx, openStackProject.Id, openStackUser.Id); err != nil {
		return err
	}

	//Check if user is still needed in region
	userProjectBindings := &v1alpha1.UserProjectBindingList{}
	if err := r.List(ctx, userProjectBindings, &client.ListOptions{Namespace: upb.Namespace}); err != nil {
		return err
	}

	userProjects := make([]v1alpha1.Project, 0)
	for _, k := range userProjectBindings.Items {
		if k.Spec.User == upb.Spec.User {
			userProject := &v1alpha1.Project{}

			if err := r.Get(ctx, types.NamespacedName{
				Namespace: k.Namespace,
				Name:      k.Spec.Project,
			}, userProject); err != nil {
				return err
			}

			userProjects = append(userProjects, *userProject)
		}
	}

	userHasProjectsInRegion := false
	for _, k := range userProjects {
		if k.Spec.Region == project.Spec.Region && k.Name != project.Name {
			userHasProjectsInRegion = true
		}
	}

	if !userHasProjectsInRegion {
		logger.Info("User has no projects in region left. Gonna delete user in region.")

		if err := psOsClient.DeleteUser(ctx, openStackUser.Id); err != nil {
			return err
		}
	} else {
		logger.Info("User has projects in region left. Gonna skip deletion.")
	}

	logger.Info("UPB finalized")
	return nil
}
