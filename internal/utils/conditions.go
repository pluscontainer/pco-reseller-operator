package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AllConditionsTrue checks if all specified conditions have the status true
// If the slice is empty, it returns false
func AllConditionsTrue(conditions []metav1.Condition) bool {
	if len(conditions) == 0 {
		return false
	}

	for _, k := range conditions {
		if k.Status != metav1.ConditionTrue {
			return false
		}
	}

	return true
}
