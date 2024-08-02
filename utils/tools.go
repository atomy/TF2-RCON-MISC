package utils

import (
	"strings"
)

// IsCrit, detects a crit in the frag-line
func IsCrit(str string) (bool, error) {
	str = strings.TrimSpace(str)

	if str == "(crit)" {
		return true, nil
	}

	return false, nil
}
