package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pluscloudopen/reseller-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	k8errors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var errUserProjectBindingPresent = errors.New("UserProjectBinding still referencing user")

func (r *UserReconciler) finalizeUser(ctx context.Context, logger logr.Logger, user v1alpha1.User) error {
	//CLeanup all UserProjectBindings
	userProjectBindings := &v1alpha1.UserProjectBindingList{}
	if err := r.List(ctx, userProjectBindings, &client.ListOptions{
		Namespace: user.Namespace,
	}); err != nil {
		return err
	}

	if len(userProjectBindings.Items) > 0 {
		for _, k := range userProjectBindings.Items {
			if err := r.Delete(ctx, &k); err != nil {
				return err
			}
		}

		return errUserProjectBindingPresent
	}

	accessSecret := &v1.Secret{}
	if err := r.Get(ctx, user.UserAccessSecretName(), accessSecret); err != nil {
		if !k8errors.IsNotFound(err) {
			return err
		}

		logger.Info(fmt.Sprintf("Secret %s already gone", user.UserAccessSecretName()))
	} else {
		//Secret is still present

		if err := r.Delete(ctx, accessSecret); err != nil {
			return err
		}
	}

	logger.Info("User finalized")
	return nil
}
