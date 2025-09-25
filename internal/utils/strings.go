package utils

import "strings"

// IsEmpty returns if the given string is empty after trimming spaces
func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
