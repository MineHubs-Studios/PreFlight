package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// PRE-COMPILED REGULAR EXPRESSIONS FOR BETTER PERFORMANCE.
var (
	phpVersionRegex = regexp.MustCompile(`PHP (\d+\.\d+\.\d+)`)
	semverRegex     = regexp.MustCompile(`(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([\w.-]+))?(?:\+([\w.-]+))?`)
	numberRegex     = regexp.MustCompile(`[0-9]+`)
)

// VersionParts REPRESENTS A SEMANTIC VERSION SPLIT INTO COMPONENTS.
type VersionParts struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
}

// ValidateVersion CHECK IF AN INSTALLED VERSION MATCHES THE REQUIREMENTS.
func ValidateVersion(installedVersion, requiredVersion string) (bool, string) {
	// REMOVE ANY "v" PREFIX AND PHP SPECIFIC FORMATTING.
	if strings.Contains(installedVersion, "PHP") {
		installedVersion = extractPHPVersion(installedVersion)
	} else {
		installedVersion = strings.TrimPrefix(installedVersion, "v")
	}

	if !MatchVersionConstraint(installedVersion, requiredVersion) {
		return false, fmt.Sprintf("Version %s is required, but version %s is installed.", requiredVersion, installedVersion)
	}

	return true, fmt.Sprintf("Required version %s is installed.", requiredVersion)
}

// extractPHPVersion EXTRACTS THE VERSION NUMBER FROM A PHP VERSION STRING.
func extractPHPVersion(phpVersionString string) string {
	matches := phpVersionRegex.FindStringSubmatch(phpVersionString)

	if len(matches) >= 2 {
		return matches[1]
	}

	return phpVersionString
}

// MatchVersionConstraint CHECKS IF AN INSTALLED VERSION MEETS THE VERSION REQUIREMENT.
// SUPPORTED OPERATORS, >=, >, <=, <, ^, and ~).
func MatchVersionConstraint(installed, required string) bool {
	// IF REQUIRED IS EMPTY, ACCEPT ANY INSTALLED VERSION.
	if required == "" {
		return true
	}

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
		return matchCaretRange(installed, required[1:])
	case strings.HasPrefix(required, "~"):
		return matchTildeRange(installed, required[1:])
	default:
		// CHECK FOR VERSION RANGES WITH SPACES.
		if strings.Contains(required, " - ") {
			return matchVersionRange(installed, required)
		}

		// DIRECT COMPARISON.
		return compareVersions(installed, required) == 0
	}
}

// matchCaretRange IMPLEMENTS THE ^ OPERATOR FROM NPM'S SEMVER.
// ^1.2.3 -> >=1.2.3 <2.0.0
func matchCaretRange(installed, required string) bool {
	installedParts := parseDetailedSemver(installed)
	requiredParts := parseDetailedSemver(required)

	// If PARSING FAILS, FALL BACK TO DIRECT COMPARISON.
	if installedParts == nil || requiredParts == nil {
		return installed == required
	}

	// VERSIONS MUST HAVE AT LEAST THE SAME MAJOR VERSION.
	if installedParts.Major != requiredParts.Major {
		// FOR VERSION 0.x.x, ^ MEANS CHANGES ONLY ALLOWED IN PATCH-LEVEL.
		if requiredParts.Major == 0 && requiredParts.Minor > 0 {
			if installedParts.Major > 0 || installedParts.Minor > requiredParts.Minor {
				return false
			}

			return installedParts.Minor == requiredParts.Minor && installedParts.Patch >= requiredParts.Patch
		}
		// FOR VERSIONS 0.0.x, ^ MEANS NO CHANGES ALLOWED.
		if requiredParts.Major == 0 && requiredParts.Minor == 0 {
			return installedParts.Major == 0 && installedParts.Minor == 0 &&
				installedParts.Patch == requiredParts.Patch
		}

		return false
	}

	// FOR NON-ZERO MAJOR VERSION, ^ ALLOWS MINOR AND PATCH CHANGES.
	if requiredParts.Major > 0 {
		return compareVersions(installed, required) >= 0 && installedParts.Major == requiredParts.Major
	}

	// FOR 0.y.z VERSIONS, ^ MEANS THE SAME AS ~
	return matchTildeRange(installed, required)
}

// matchTildeRange IMPLEMENTS THE ~ OPERATOR FROM NPM'S SEMVER.
// ~1.2.3 -> >=1.2.3 <1.3.0
// ~1.2 -> >=1.2.0 <1.3.0
func matchTildeRange(installed, required string) bool {
	requiredParts := parseDetailedSemver(required)
	installedParts := parseDetailedSemver(installed)

	if requiredParts == nil || installedParts == nil {
		return installed == required
	}

	// MAJOR VERSION MUST ALWAYS MATCH.
	if installedParts.Major != requiredParts.Major {
		return false
	}

	// IF A MINOR VERSION IS SPECIFIED, IT MUST ALSO MATCH.
	if requiredParts.Minor != -1 && installedParts.Minor != requiredParts.Minor {
		return false
	}

	// IF BOTH MAJOR AND MINOR MATCH, PATCH VERSION MUST BE >= REQUIRED.
	if requiredParts.Patch != -1 {
		if installedParts.Minor == requiredParts.Minor {
			return installedParts.Patch >= requiredParts.Patch
		}

		// IF MINOR IS HIGHER, ANY PATCH VERSION IS OK.
		return installedParts.Minor > requiredParts.Minor
	}

	return true
}

// matchVersionRange IMPLEMENTS VERSION RANGE CHECKING IN THE FORMAT (1.0.0 - 2.0.0).
func matchVersionRange(installed, rangeStr string) bool {
	parts := strings.Split(rangeStr, " - ")

	if len(parts) != 2 {
		return false
	}

	minVersion, maxVersion := parts[0], parts[1]

	return compareVersions(installed, minVersion) >= 0 && compareVersions(installed, maxVersion) <= 0
}

// compareVersions COMPARES TWO VERSIONS AND RETURNS -1, 0, or 1 for less, equal, or greater.
func compareVersions(v1, v2 string) int {
	parts1 := parseDetailedSemver(v1)
	parts2 := parseDetailedSemver(v2)

	if parts1 == nil || parts2 == nil {
		// IF PARSING FAILS, FALL BACK TO SIMPLER COMPARISON.
		v1Parts, v2Parts := parseSemver(v1), parseSemver(v2)
		return compareVersionArrays(v1Parts, v2Parts)
	}

	// COMPARE MAJOR VERSIONS.
	if parts1.Major != parts2.Major {
		if parts1.Major > parts2.Major {
			return 1
		}

		return -1
	}

	// COMPARE MINOR VERSIONS.
	if parts1.Minor != parts2.Minor {
		if parts1.Minor > parts2.Minor {
			return 1
		}

		return -1
	}

	// COMPARE PATCH VERSIONS.
	if parts1.Patch != parts2.Patch {
		if parts1.Patch > parts2.Patch {
			return 1
		}

		return -1
	}

	// COMPARE PRERELEASE (PRERELEASE IS LOWER THAN NO PRERELEASE).
	if parts1.Prerelease == "" && parts2.Prerelease != "" {
		return 1
	}

	if parts1.Prerelease != "" && parts2.Prerelease == "" {
		return -1
	}

	if parts1.Prerelease != parts2.Prerelease {
		if parts1.Prerelease > parts2.Prerelease {
			return 1
		}

		return -1
	}

	// EVERYTHING MATCHES, VERSIONS ARE EQUAL.
	return 0
}

// compareVersionArrays COMPARES TWO VERSION ARRAYS.
func compareVersionArrays(v1Parts, v2Parts []int) int {
	maxLen := len(v1Parts)

	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	// EXTEND BOTH ARRAYS TO THE SAME LENGTH BY ADDING ZEROS.
	for len(v1Parts) < maxLen {
		v1Parts = append(v1Parts, 0)
	}

	for len(v2Parts) < maxLen {
		v2Parts = append(v2Parts, 0)
	}

	// COMPARE EACH PART.
	for i := 0; i < maxLen; i++ {
		if v1Parts[i] < v2Parts[i] {
			return -1
		} else if v1Parts[i] > v2Parts[i] {
			return 1
		}
	}

	return 0
}

// parseDetailedSemver PARSE A VERSION STRING INTO STRUCTURED COMPONENTS.
func parseDetailedSemver(version string) *VersionParts {
	matches := semverRegex.FindStringSubmatch(version)

	if len(matches) < 4 {
		return nil
	}

	result := &VersionParts{
		Major:      -1,
		Minor:      -1,
		Patch:      -1,
		Prerelease: "",
		Build:      "",
	}

	// PARSE MAJOR VERSION (MANDATORY).
	if major, err := strconv.Atoi(matches[1]); err == nil {
		result.Major = major
	} else {
		return nil
	}

	// PARSE MINOR VERSION (OPTIONAL)
	if matches[2] != "" {
		if minor, err := strconv.Atoi(matches[2]); err == nil {
			result.Minor = minor
		}
	}

	// PARSE PATCH VERSION (OPTIONAL)
	if matches[3] != "" {
		if patch, err := strconv.Atoi(matches[3]); err == nil {
			result.Patch = patch
		}
	}

	// STORE PRERELEASE AND BUILD METADATA IF AVAILABLE.
	if len(matches) >= 5 && matches[4] != "" {
		result.Prerelease = matches[4]
	}

	if len(matches) >= 6 && matches[5] != "" {
		result.Build = matches[5]
	}

	return result
}

// parseSemver PARSES A VERSION STRING INTO AN ARRAY OF INTEGERS.
func parseSemver(version string) []int {
	parts := numberRegex.FindAllString(version, -1)
	parsed := make([]int, 0, len(parts))

	for _, part := range parts {
		if num, err := strconv.Atoi(part); err == nil {
			parsed = append(parsed, num)
		}
	}

	return parsed
}
