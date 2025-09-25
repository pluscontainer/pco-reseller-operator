package utils

import (
	"context"
	"errors"
	"math/rand"
	"os"

	v1 "k8s.io/api/core/v1"
	k8errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	controllerIdSecretName = "pco-reseller-operator-id"
	controllerIdSecretKey  = "id"
)

var errControllerIdNotSet = errors.New("controller id not set")

// ControllerIdentifier fetches the pco-reseller-operator-id secret and retrieves the unique operator id from it
func ControllerIdentifier(ctx context.Context, r client.Client) (*string, error) {
	controllerNamespace := os.Getenv("CONTROLLER_NAMESPACE")

	controllerIdentifierSecret := &v1.Secret{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      controllerIdSecretName,
		Namespace: controllerNamespace,
	}, controllerIdentifierSecret)

	if err != nil {
		if !k8errors.IsNotFound(err) {
			return nil, err
		}

		//Generate random controller id
		letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

		randomCharacters := make([]rune, 6)
		for i := range randomCharacters {
			randomCharacters[i] = letterRunes[rand.Intn(len(letterRunes))]
		}

		isTrue := true

		controllerIdentifierSecret = &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      controllerIdSecretName,
				Namespace: controllerNamespace,
			},
			StringData: map[string]string{
				controllerIdSecretKey: string(randomCharacters),
			},
			Immutable: &isTrue,
		}

		if err := r.Create(ctx, controllerIdentifierSecret); err != nil {
			return nil, err
		}
	}

	controllerIdentifier := string(controllerIdentifierSecret.Data[controllerIdSecretKey])

	if IsEmpty(controllerIdentifier) {
		return nil, errControllerIdNotSet
	}

	return &controllerIdentifier, nil
}
