package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateVersion CHECK IF AN INSTALLED VERSION MATCHES THE REQUIREMENTS.
func ValidateVersion(installedVersion, requiredVersion string) (bool, string) {
	installedVersion = strings.TrimPrefix(installedVersion, "v")

	if !MatchVersionConstraint(installedVersion, requiredVersion) {
		return false, fmt.Sprintf("Version %s is required, but version %s is installed.", requiredVersion, installedVersion)
	}

	return true, fmt.Sprintf("Required version %s is installed.", requiredVersion)
}

// MatchVersionConstraint MATCH NODE VERSION CONSTRAINTS LIKE >=, >, <=, < AND ^.
func MatchVersionConstraint(installed, required string) bool {
	switch {
	case strings.HasPrefix(required, ">="):
		return compareVersions(installed, required[2:]) >= 0
	case strings.HasPrefix(required, ">"):
		return compareVersions(installed, required[1:]) > 0
	case strings.HasPrefix(required, "<="):
		return compareVersions(installed, required[2:]) <= 0
	case strings.HasPrefix(required, "<"):
		return compareVersions(installed, required[1:]) < 0
	case strings.HasPrefix(required, "^"):
		return compareVersionsWithinMajor(installed, required[1:])
	default:
		return installed == required
	}
}

// COMPARE TWO VERSIONS RETURNING -1, 0, OR 1 FOR LESS, EQUAL, OR GREATER.
func compareVersions(v1, v2 string) int {
	v1Parts, v2Parts := parseSemver(v1), parseSemver(v2)

	for i := 0; len(v1Parts) > i && len(v2Parts) > i; i++ {
		if v1Parts[i] < v2Parts[i] {
			return -1
		} else if v1Parts[i] > v2Parts[i] {
			return 1
		}
	}

	return 0
}

// COMPARE INSTALLED VERSION WITHIN SAME MAJOR VERSION.
func compareVersionsWithinMajor(installed, required string) bool {
	installedParts, requiredParts := parseSemver(installed), parseSemver(required)

	if len(installedParts) == 0 || len(requiredParts) == 0 || installedParts[0] != requiredParts[0] {
		return false
	}

	return compareVersions(installed, required) >= 0
}

// PARSE SEMANTIC VERSION INTO INTEGERS FOR COMPARISON.
func parseSemver(version string) []int {
	parts := regexp.MustCompile(`[0-9]+`).FindAllString(version, -1)
	parsed := make([]int, len(parts))

	for i, part := range parts {
		_, err := fmt.Sscanf(part, "%d", &parsed[i])
		if err != nil {
			return nil
		}
	}

	return parsed
}
