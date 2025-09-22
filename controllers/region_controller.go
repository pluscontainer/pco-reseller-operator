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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/pluscontainer/pco-reseller-cli/pkg/psos"
	"github.com/pluscontainer/pco-reseller-operator/api/v1alpha1"
	pcov1alpha1 "github.com/pluscontainer/pco-reseller-operator/api/v1alpha1"
)

// RegionReconciler reconciles a Region object
type RegionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const controllerFinalizer = "pco.plusserver.com/finalizer"

//+kubebuilder:rbac:groups=pco.plusserver.com,resources=regions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pco.plusserver.com,resources=regions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pco.plusserver.com,resources=regions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Region object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *RegionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Region")

	// Fetch the Region instance
	region := &v1alpha1.Region{}
	err := r.Get(ctx, req.NamespacedName, region)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("Region resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get region.")
		return ctrl.Result{}, err
	}

	// Check if the region is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isRegionMarkedToBeDelted := region.GetDeletionTimestamp() != nil
	if isRegionMarkedToBeDelted {
		if controllerutil.ContainsFinalizer(region, controllerFinalizer) {
			// Run finalization logic for controllerFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeRegion(ctx, logger, *region); err != nil {
				if err != errProjectStillReferencingRegion {
					return ctrl.Result{}, err
				}

				//Projects are still referencing region
				return ctrl.Result{RequeueAfter: time.Duration(3) * time.Second}, nil
			}

			// Remove controllerFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(region, controllerFinalizer)
			err := r.Update(ctx, region)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(region, controllerFinalizer) {
		controllerutil.AddFinalizer(region, controllerFinalizer)
		err = r.Update(ctx, region)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	_, err = psos.Login(region.Spec.Endpoint, region.Spec.Username, region.Spec.Password)
	if err != nil {
		if err := region.UpdateRegionCondition(ctx, r.Client, pcov1alpha1.RegionIsUnready, err.Error()); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, err
	}

	if err := region.UpdateRegionCondition(ctx, r.Client, pcov1alpha1.RegionIsReady, "API of region is reachable"); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling finished")
	return ctrl.Result{RequeueAfter: time.Minute * 15}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RegionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 5,
		}).
		WithEventFilter(pred).
		For(&pcov1alpha1.Region{}).
		Complete(r)
}
