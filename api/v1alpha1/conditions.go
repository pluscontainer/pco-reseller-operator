package v1alpha1

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// ConditionTypes stores the diffrent kinds of conditions
type ConditionTypes string

const (
	// UserReady represents whether the OpenStack user is ready or not
	UserReady ConditionTypes = "UserReady"
	// ProjectReady represents whether the OpenStack project is ready or not
	ProjectReady ConditionTypes = "ProjectReady"
	// RegionReady represents whether the OpenStack region is ready or not
	RegionReady ConditionTypes = "RegionReady"
	// UserProjectBindingReady represents whether the userprojectbinding object is ready or not
	UserProjectBindingReady ConditionTypes = "UserProjectBindingReady"
)

// RegionReadyReasons are the different states of readiness, which a region can have
type RegionReadyReasons string

const (
	// RegionIsReady is set when the region is ready
	RegionIsReady RegionReadyReasons = "RegionIsReady"
	// RegionIsUnready is set when the region is not ready
	RegionIsUnready RegionReadyReasons = "RegionIsUnready"
	// RegionNotFound is set when the region could not be found
	RegionNotFound RegionReadyReasons = "RegionNotFound"
	// RegionUnknown is set if the readiness could not be determined
	RegionUnknown RegionReadyReasons = "UnknownError"
)

func (r RegionReadyReasons) regionStatus() v1.ConditionStatus {
	switch r {
	case RegionIsReady:
		{
			return v1.ConditionTrue
		}
	case RegionUnknown:
		{
			return v1.ConditionUnknown
		}
	default:
		{
			return v1.ConditionFalse
		}
	}
}

// ProjectReadyReasons are the different states of readiness, which a project can have
type ProjectReadyReasons string

const (
	// ProjectIsReady is set when the project is ready
	ProjectIsReady ProjectReadyReasons = "ProjectIsReady"
	// ProjectIsUnready is set when the project is not ready
	ProjectIsUnready ProjectReadyReasons = "ProjectIsUnready"
	// ProjectNotFound is set when the project could not be found
	ProjectNotFound ProjectReadyReasons = "ProjectNotFound"
	// ProjectUnknown is set if the readiness could not be determined
	ProjectUnknown ProjectReadyReasons = "UnknownError"
)

func (r ProjectReadyReasons) projectStatus() v1.ConditionStatus {
	switch r {
	case ProjectIsReady:
		{
			return v1.ConditionTrue
		}
	case ProjectUnknown:
		{
			return v1.ConditionUnknown
		}
	default:
		{
			return v1.ConditionFalse
		}
	}
}

// UserReadyReasons are the different states of readiness, which a user can have
type UserReadyReasons string

const (
	// UserIsReady is set when the user is ready
	UserIsReady UserReadyReasons = "UserIsReady"
	// UserHasNoSecret is set when the users secret could not be found
	UserHasNoSecret UserReadyReasons = "SecretNotFound"
	// UserNotFound is set when the user could not be found
	UserNotFound UserReadyReasons = "UserNotFound"
	// UserIsUnready is set when the user is not ready
	UserIsUnready UserReadyReasons = "UserIsUnready"
	// UserUnknown is set if the readiness could not be determined
	UserUnknown UserReadyReasons = "UnknownError"
)

func (r UserReadyReasons) userStatus() v1.ConditionStatus {
	switch r {
	case UserIsReady:
		{
			return v1.ConditionTrue
		}
	case UserUnknown:
		{
			return v1.ConditionUnknown
		}
	default:
		{
			return v1.ConditionFalse
		}
	}
}

// UserProjectBindingReadyReasons are the different states of readiness, which a userprojectbinding can have
type UserProjectBindingReadyReasons string

const (
	// UserProjectBindingIsReady is set when the userprojectbinding is ready
	UserProjectBindingIsReady UserProjectBindingReadyReasons = "UserProjectBindingIsReady"
	// UserProjectBindingUnknown is set if the readiness could not be determined
	UserProjectBindingUnknown UserProjectBindingReadyReasons = "UnknownError"
)

func (r UserProjectBindingReadyReasons) userProjectBindingStatus() v1.ConditionStatus {
	switch r {
	case UserProjectBindingIsReady:
		{
			return v1.ConditionTrue
		}
	case UserProjectBindingUnknown:
		{
			return v1.ConditionUnknown
		}
	default:
		{
			return v1.ConditionFalse
		}
	}
}
