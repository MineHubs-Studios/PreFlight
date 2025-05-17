package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Pre-compiled regular expressions for better performance.
var (
	phpVersionRegex = regexp.MustCompile(`PHP (\d+\.\d+\.\d+)`)
	semverRegex     = regexp.MustCompile(`(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([\w.-]+))?(?:\+([\w.-]+))?`)
	numberRegex     = regexp.MustCompile(`[0-9]+`)
)

// VersionParts represents a semantic version split into components.
type VersionParts struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
}

// compareInts compares two integers and returns -1, 0, or 1.
func compareInts(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// ValidateVersion check if an installed version matches the requirements.
func ValidateVersion(installedVersion, requiredVersion string) (bool, string) {
	// Remove any "v" prefix and PHP-specific formatting.
	if strings.Contains(installedVersion, "PHP") {
		installedVersion = extractPHPVersion(installedVersion)
	} else {
		installedVersion = strings.TrimPrefix(installedVersion, "v")
	}

	if !MatchVersionConstraint(installedVersion, requiredVersion) {
		return false, fmt.Sprintf("Version %s is required, but version %s is installed.", requiredVersion, installedVersion)
	}

	return true, ""
}

// extractPHPVersion extracts the PHP version from a string.
func extractPHPVersion(phpVersionString string) string {
	matches := phpVersionRegex.FindStringSubmatch(phpVersionString)

	if len(matches) >= 2 {
		return matches[1]
	}

	return phpVersionString
}

// MatchVersionConstraint checks if the installed version matches the required version.
// Supported operators, >=, >, <=, <, ^, and ~).
func MatchVersionConstraint(installed, required string) bool {
	// If required is empty, accept any installed version.
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
		// Check for version ranges with spaces.
		if strings.Contains(required, " - ") {
			return matchVersionRange(installed, required)
		}

		// Direct comparison.
		return compareVersions(installed, required) == 0
	}
}

// matchCaretRange implements the ^ operator from NPM's semver.
// ^1.2.3 -> >=1.2.3 <2.0.0
func matchCaretRange(installed, required string) bool {
	installedParts := parseDetailedSemver(installed)
	requiredParts := parseDetailedSemver(required)

	// If parsing fails, fallback to direct comparison.
	if installedParts == nil || requiredParts == nil {
		return installed == required
	}

	// Versions must have at least the same major version.
	if installedParts.Major != requiredParts.Major {
		// For version 0.x.x, ^ means changes only allowed in patch level.
		if requiredParts.Major == 0 && requiredParts.Minor > 0 {
			if installedParts.Major > 0 || installedParts.Minor > requiredParts.Minor {
				return false
			}

			return installedParts.Minor == requiredParts.Minor && installedParts.Patch >= requiredParts.Patch
		}
		// For versions 0.0.x, ^ means no changes allowed.
		if requiredParts.Major == 0 && requiredParts.Minor == 0 {
			return installedParts.Major == 0 && installedParts.Minor == 0 &&
				installedParts.Patch == requiredParts.Patch
		}

		return false
	}

	// For non-zero major version, ^ allows minor and patch changes.
	if requiredParts.Major > 0 {
		return compareVersions(installed, required) >= 0 && installedParts.Major == requiredParts.Major
	}

	// For 0.y.z versions, ^ means the same as ~
	return matchTildeRange(installed, required)
}

// matchTildeRange implements the ~ operator from NPM's semver.
// ~1.2.3 -> >=1.2.3 <1.3.0
// ~1.2 -> >=1.2.0 <1.3.0
func matchTildeRange(installed, required string) bool {
	requiredParts := parseDetailedSemver(required)
	installedParts := parseDetailedSemver(installed)

	if requiredParts == nil || installedParts == nil {
		return installed == required
	}

	// Major version must always match.
	if installedParts.Major != requiredParts.Major {
		return false
	}

	// If a minor version is specified, it must also mach.
	if requiredParts.Minor != -1 && installedParts.Minor != requiredParts.Minor {
		return false
	}

	// If both major and minor match, a patch version must be >= required.
	if requiredParts.Patch != -1 {
		if installedParts.Minor == requiredParts.Minor {
			return installedParts.Patch >= requiredParts.Patch
		}

		// If minor is higher, any patch version is ok.
		return installedParts.Minor > requiredParts.Minor
	}

	return true
}

// matchVersionRange checks if the installed version is within a specified range.
func matchVersionRange(installed, rangeStr string) bool {
	minVersion, maxVersion, found := strings.Cut(rangeStr, " - ")

	if !found {
		return false
	}

	return compareVersions(installed, minVersion) >= 0 &&
		compareVersions(installed, maxVersion) <= 0
}

// compareVersions compares two version strings and returns -1, 0, or 1.
func compareVersions(v1, v2 string) int {
	parts1 := parseDetailedSemver(v1)
	parts2 := parseDetailedSemver(v2)

	if parts1 == nil || parts2 == nil {
		// If parsing fails, fallback to simpler comparison.
		v1Parts, v2Parts := parseSemver(v1), parseSemver(v2)
		return compareVersionArrays(v1Parts, v2Parts)
	}

	// Compare major versions.
	if parts1.Major != parts2.Major {
		return compareInts(parts1.Major, parts2.Major)
	}

	// Compare minor versions.
	if parts1.Minor != parts2.Minor {
		if parts1.Minor > parts2.Minor {
			return 1
		}

		return -1
	}

	// Compare patch versions.
	if parts1.Patch != parts2.Patch {
		if parts1.Patch > parts2.Patch {
			return 1
		}

		return -1
	}

	// Compare prerelease (prerelease is lower than no prerelease).
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

	// Everything matches, versions are equal.
	return 0
}

// compareVersionArrays Compares two version arrays.
func compareVersionArrays(v1Parts, v2Parts []int) int {
	maxLen := max(len(v1Parts), len(v2Parts))

	// Create copies to avoid modifying the originals.
	v1Extended := make([]int, maxLen)
	v2Extended := make([]int, maxLen)

	copy(v1Extended, v1Parts)
	copy(v2Extended, v2Parts)

	// Compare each part.
	for i := range v1Extended {
		if v1Extended[i] != v2Extended[i] {
			return compareInts(v1Extended[i], v2Extended[i])
		}
	}

	return 0
}

// parseDetailedSemver parses a semantic version string into its components.
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

	// Parse major version (mandatory).
	if major, err := strconv.Atoi(matches[1]); err == nil {
		result.Major = major
	} else {
		return nil
	}

	// Parse minor version (optional).
	if matches[2] != "" {
		if minor, err := strconv.Atoi(matches[2]); err == nil {
			result.Minor = minor
		}
	}

	// Parse path version (optional).
	if matches[3] != "" {
		if patch, err := strconv.Atoi(matches[3]); err == nil {
			result.Patch = patch
		}
	}

	// Store prerelease and build metadata if available.
	if len(matches) >= 5 && matches[4] != "" {
		result.Prerelease = matches[4]
	}

	if len(matches) >= 6 && matches[5] != "" {
		result.Build = matches[5]
	}

	return result
}

// parseSemver parses a semantic version string into an array of integers.
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
