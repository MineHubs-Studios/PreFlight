package modules

import (
	"PreFlight/utils"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type NpmModule struct{}

func (n NpmModule) Name() string {
	return "NPM"
}

// CheckRequirements CHECK THE REQUIREMENTS FOR THE NPM MODULE.
func (n NpmModule) CheckRequirements(context map[string]interface{}) (errors []string, warnings []string, successes []string) {
	// CHECK IF NODE IS INSTALLED.
	nodeVersionOutput := isNodeInstalled(&errors, &successes)

	// READ package.json TO EXTRACT REQUIRED NODE VERSION AND DEPENDENCIES.
	requiredNodeVersion, packageFound, requiredDeps := utils.ReadPackageJSON()

	// HANDLE MISSING package.json.
	if !packageFound {
		errors = append(errors, "package.json not found.")
		handleLockFileWarnings(&warnings)
		return errors, warnings, successes
	}

	successes = append(successes, "package.json found.")

	// VALIDATE NODE VERSION IF SPECIFIC VERSION IS REQUIRED.
	if requiredNodeVersion != "" {
		if isValid, feedback := validateVersion(nodeVersionOutput, requiredNodeVersion); isValid {
			successes = append(successes, feedback)
		} else {
			errors = append(errors, feedback)
		}
	} else {
		successes = append(successes, "No specific Node.js version is required.")
	}

	// CHECK ALL NPM PACKAGES DEFINED IN package.json.
	for _, dep := range requiredDeps {
		if !checkNpmPackage(dep) {
			errors = append(errors, fmt.Sprintf("NPM package %s is missing. Run `npm install %s`.", dep, dep))
		} else {
			successes = append(successes, fmt.Sprintf("NPM package %s is installed.", dep))
		}
	}

	return errors, warnings, successes
}

// VALIDATE NODE INSTALLATION AND OBTAIN INSTALLED NODE VERSION.
func isNodeInstalled(errors *[]string, successes *[]string) string {
	cmd := exec.Command("node", "--version")
	var outBuffer bytes.Buffer
	cmd.Stdout = &outBuffer

	err := cmd.Run()

	if err != nil {
		*errors = append(*errors, "Node.js is not installed. Please install Node.js to use NPM.")
		return ""
	}

	installedVersion := strings.TrimSpace(outBuffer.String())
	*successes = append(*successes, fmt.Sprintf("Node.js is installed with version %s.", installedVersion))

	return installedVersion
}

// HANDLE MISSING package.json BY CHECKING FOR LOCK FILES.
func handleLockFileWarnings(warnings *[]string) {
	lockFiles := []string{"package-lock.json", "pnpm-lock.yaml"}
	found := false

	for _, file := range lockFiles {
		if _, err := os.Stat(file); err == nil {
			found = true
			break
		}
	}

	if found {
		*warnings = append(*warnings, "package.json not found, but a lock file exists. Ensure package.json is included in your project.")
	} else {
		*warnings = append(*warnings, "Neither package.json nor lock files are found.")
	}
}

// VALIDATE IF INSTALLED NODE VERSION MATCHES THE REQUIRED VERSION.
func validateVersion(installedVersion, requiredVersion string) (bool, string) {
	installedVersion = strings.TrimPrefix(installedVersion, "v")

	if !matchVersionConstraint(installedVersion, requiredVersion) {
		return false, fmt.Sprintf("Node.js version %s is required, but version %s is installed.", requiredVersion, installedVersion)
	}

	return true, fmt.Sprintf("Required Node.js version %s is installed.", requiredVersion)
}

// MATCH NODE VERSION CONSTRAINTS LIKE >=, >, <=, < AND ^.
func matchVersionConstraint(installed, required string) bool {
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

// CHECK IF AN NPM PACKAGE IS INSTALLED BY RUNNING `npm list`.
func checkNpmPackage(packageName string) bool {
	cmd := exec.Command("npm", "list", packageName, "--depth=0")
	var outBuffer, errBuffer bytes.Buffer
	cmd.Stdout, cmd.Stderr = &outBuffer, &errBuffer

	if err := cmd.Run(); err != nil || strings.Contains(errBuffer.String(), "missing") || strings.TrimSpace(outBuffer.String()) == "" {
		return false
	}

	return true
}
