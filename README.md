# pluscloud open Reseller Operator

The pco-reseller-operator allows Kubernetes Operators / Users to manage projects within our pluscloud open via easy to handle Kubernetes manifests.

## CRDs
This operator implements / acts on 4 CRDs:
- Region (<span>regions.pco.plusserver.com</span>)
- Project (<span>projects.pco.plusserver.com</span>)
- User (<span>users.pco.plusserver.com</span>)
- UserProjectBinding (<span>userprojectbindings.pco.plusserver.com</span>)

### Region
A Region is the equivalent of an endpoint configuration.
It specifies the URL of the Reseller API and the username and password used to authenticate.

An example can be found [here](./config/samples/pco_v1alpha1_region.yaml)

### Project
A Project is exactly what it says: A project within our pluscloud open.

An example can be found [here](./config/samples/pco_v1alpha1_project.yaml)

### User
A User represents a set of credentials stored within a secret.
Since we don't know which projects (and thus regions) the user will be assigned to when we create it, we can't create the OpenStack user via the Reseller API yet.

An example can be found [here](./config/samples/pco_v1alpha1_user.yaml)
(Currently the only field needed for a user is its name)

The generated secret will be named: user-sample-openstack

### UserProjectBinding
A UserProjectBinding gives the specified user access to the specified project.
This resource also causes the OpenStack user to be created / manifested via the Reseller API.

A user can be bound to multiple projects across regions.

An example can be found [here](./config/samples/pco_v1alpha1_userprojectbinding.yaml)

## The problem of uniqueness
We wanted to support running multiple deployments of this operator across multiple clusters but this comes with a challenge:
How do we make projects and users within OpenStack unique?
Nothing would prevent the creation of the same user in the same namespace on multiple clusters.

This is why every controller can be configured with a **controller identifier**.
The controller identifier is part of the OpenStack username / project name and prevents the collision described above.

The current controller identifier is stored within the secret "pco-reseller-operator-id" in the namespace of the operator.
