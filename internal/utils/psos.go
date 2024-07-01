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

var ErrOpenStackProjectNotFound = errors.New("openstack project not found")
var ErrOpenStackUserNotFound = errors.New("openstack user not found")

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

func GetOpenStackProjectName(controllerId string, project types.NamespacedName) string {
	return fmt.Sprintf("%s-%s-%s", controllerId, project.Namespace, project.Name)
}
