package controllers

import (
	"context"

	"git.ps-intern.de/mk/gardener/pco-reseller-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
)

const operationNotInitialized = "OperationNotInitialized"

func (r *ProjectReconciler) initializeConditions(ctx context.Context, project *v1alpha1.Project) error {
	if meta.FindStatusCondition(project.Status.Conditions, string(v1alpha1.RegionReady)) == nil {
		if err := project.UpdateRegionCondition(ctx, r.Client, v1alpha1.RegionUnknown, "Operation not initialized"); err != nil {
			return err
		}
	}

	if meta.FindStatusCondition(project.Status.Conditions, string(v1alpha1.ProjectReady)) == nil {
		if err := project.UpdateProjectCondition(ctx, r.Client, v1alpha1.ProjectUnknown, "Operation not initialized"); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProjectReconciler) setProjectReadyError(ctx context.Context, project *v1alpha1.Project, err error) error {
	if err := project.UpdateProjectCondition(ctx, r.Client, v1alpha1.ProjectUnknown, err.Error()); err != nil {
		return err
	}
	return nil
}
