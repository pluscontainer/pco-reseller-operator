package controllers

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/pluscloudopen/reseller-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var errProjectStillReferencingRegion = errors.New("projects are still referencing region")

func (r *RegionReconciler) finalizeRegion(ctx context.Context, logger logr.Logger, region v1alpha1.Region) error {
	//Check if any projects reference region
	//Namespace: "" = All namespaces
	projectList := &v1alpha1.ProjectList{}
	if err := r.List(ctx, projectList, &client.ListOptions{Namespace: ""}); err != nil {
		return err
	}

	referenceExists := false
	for _, k := range projectList.Items {
		if k.Spec.Region == region.Name {
			referenceExists = true
		}
	}

	if referenceExists {
		logger.Info("Projects are still referencing region. Blocking deletion")
		return errProjectStillReferencingRegion
	}

	return nil
}
