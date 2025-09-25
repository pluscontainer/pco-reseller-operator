package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pluscontainer/pco-reseller-cli/pkg/openapi"
	"github.com/pluscontainer/pco-reseller-cli/pkg/psos"
	"k8s.io/apimachinery/pkg/types"
)

// ErrOpenStackProjectNotFound is returned if the project is not found
var ErrOpenStackProjectNotFound = errors.New("openstack project not found")

// ErrOpenStackUserNotFound is returned if the user is not found
var ErrOpenStackUserNotFound = errors.New("openstack user not found")

// GetOpenStackProject returns the openstack project from the API
func GetOpenStackProject(ctx context.Context, client *psos.PsOpenstackClient, openStackProjectName string) (*openapi.ProjectCreatedResponse, error) {
	existingProjects, err := client.GetProjects(ctx)
	if err != nil {
		return nil, err
	}

	for _, k := range *existingProjects {
		//Use HasSuffix as the domain gets prepended to our project name
		if strings.HasSuffix(k.Name, openStackProjectName) {
			return &k, nil
		}
	}

	return nil, ErrOpenStackProjectNotFound
}

// GetOpenStackUser returns the openstack user from the API
func GetOpenStackUser(ctx context.Context, client *psos.PsOpenstackClient, openStackUsername string) (*openapi.CreatedOpenStackUser, error) {
	existingUsers, err := client.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	for _, k := range *existingUsers {
		if k.Name == openStackUsername {
			return &k, nil
		}
	}

	return nil, ErrOpenStackUserNotFound
}

// GetOpenStackProjectName templates the project name
func GetOpenStackProjectName(controllerId string, project types.NamespacedName) string {
	return fmt.Sprintf("%s-%s-%s", controllerId, project.Namespace, project.Name)
}
