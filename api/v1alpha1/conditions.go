package v1alpha1

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ConditionTypes string

const (
	UserReady               ConditionTypes = "UserReady"
	ProjectReady            ConditionTypes = "ProjectReady"
	RegionReady             ConditionTypes = "RegionReady"
	UserProjectBindingReady ConditionTypes = "UserProjectBindingReady"
)

type RegionReadyReasons string

const (
	RegionIsReady   RegionReadyReasons = "RegionIsReady"
	RegionIsUnready RegionReadyReasons = "RegionIsUnready"
	RegionNotFound  RegionReadyReasons = "RegionNotFound"
	RegionUnknown   RegionReadyReasons = "UnknownError"
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

type ProjectReadyReasons string

const (
	ProjectIsReady   ProjectReadyReasons = "ProjectIsReady"
	ProjectIsUnready ProjectReadyReasons = "ProjectIsUnready"
	ProjectNotFound  ProjectReadyReasons = "ProjectNotFound"
	ProjectUnknown   ProjectReadyReasons = "UnknownError"
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

type UserReadyReasons string

const (
	UserIsReady     UserReadyReasons = "UserIsReady"
	UserHasNoSecret UserReadyReasons = "SecretNotFound"
	UserNotFound    UserReadyReasons = "UserNotFound"
	UserIsUnready   UserReadyReasons = "UserIsUnready"
	UserUnknown     UserReadyReasons = "UnknownError"
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

type UserProjectBindingReadyReasons string

const (
	UserProjectBindingIsReady UserProjectBindingReadyReasons = "UserProjectBindingIsReady"
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
