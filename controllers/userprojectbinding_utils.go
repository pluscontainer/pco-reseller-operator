package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Create or patch checks if the given object exists and if so, patches it.
// If the object doesn't exist, it gets created
// ToDo: More generic approach for all reconcilers
func (r *UserProjectBindingReconciler) createOrPatch(ctx context.Context, obj client.Object) error {
	objName := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}

	originalObject := obj.DeepCopyObject().(client.Object)

	if err := r.Get(ctx, objName, obj); err != nil {
		//Return all errors except not found
		if !errors.IsNotFound(err) {
			return err
		}

		//Create object
		return r.Create(ctx, obj)
	}

	//Patch object
	return r.Patch(ctx, obj, client.MergeFrom(originalObject))
}
