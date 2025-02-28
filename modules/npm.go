package modules

import (
	"PreFlight/utils"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type NpmModule struct{}

func (n NpmModule) Name() string {
	return "NPM"
}

// CheckRequirements CHECK THE REQUIREMENTS FOR THE NPM MODULE.
func (n NpmModule) CheckRequirements(ctx context.Context, params map[string]interface{}) (errors []string, warnings []string, successes []string) {
	select {
	case <-ctx.Done():
		return nil, nil, nil
	default:
	}

	// CHECK IF NODE IS INSTALLED.
	nodeVersionOutput := isNodeInstalled(ctx, &errors, &successes)

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
		if isValid, feedback := utils.ValidateVersion(nodeVersionOutput, requiredNodeVersion); isValid {
			successes = append(successes, feedback)
		} else {
			errors = append(errors, feedback)
		}
	} else {
		successes = append(successes, "No specific Node.js version is required.")
	}

	// CHECK ALL NPM PACKAGES DEFINED IN package.json.
	for _, dep := range requiredDeps {
		if !checkNpmPackage(ctx, dep) {
			errors = append(errors, fmt.Sprintf("NPM package %s is missing. Run `npm install %s`.", dep, dep))
		} else {
			successes = append(successes, fmt.Sprintf("NPM package %s is installed.", dep))
		}
	}

	return errors, warnings, successes
}

// VALIDATE NODE INSTALLATION AND OBTAIN INSTALLED NODE VERSION.
func isNodeInstalled(ctx context.Context, errors *[]string, successes *[]string) string {
	cmd := exec.CommandContext(ctx, "node", "--version")
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

// CHECK IF AN NPM PACKAGE IS INSTALLED BY RUNNING `npm list`.
func checkNpmPackage(ctx context.Context, packageName string) bool {
	cmd := exec.CommandContext(ctx, "npm", "list", packageName, "--depth=0")
	var outBuffer, errBuffer bytes.Buffer
	cmd.Stdout, cmd.Stderr = &outBuffer, &errBuffer

	if err := cmd.Run(); err != nil || strings.Contains(errBuffer.String(), "missing") || strings.TrimSpace(outBuffer.String()) == "" {
		return false
	}

	return true
}
