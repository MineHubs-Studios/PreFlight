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

func determineNpmPackageManager() (command string, lockFile string) {
	if _, err := os.Stat("pnpm-lock.yaml"); err == nil {
		return "pnpm", "pnpm-lock.yaml"
	}

	if _, err := os.Stat("package-lock.json"); err == nil {
		return "npm", "package-lock.json"
	}

	return "npm", ""
}

// CheckRequirements CHECK THE REQUIREMENTS FOR THE NPM MODULE.
func (n NpmModule) CheckRequirements(ctx context.Context, params map[string]interface{}) (errors []string, warnings []string, successes []string) {
	select {
	case <-ctx.Done():
		return nil, nil, nil
	default:
	}

	// READ package.json TO EXTRACT DEPENDENCIES.
	_, found, requiredDeps := utils.ReadPackageJSON()

	pm, lockFile := determineNpmPackageManager()

	// HANDLE MISSING package.json.
	if !found {
		errors = append(errors, "package.json not found.")

		if lockFile != "" {
			warnings = append(warnings, fmt.Sprintf("package.json not found, but %s exists. Ensure package.json is included in your project.", lockFile))
		} else {
			warnings = append(warnings, "Neither package.json nor lock files (package-lock.json, pnpm-lock.yaml) are found.")
		}

		return errors, warnings, successes
	}

	successes = append(successes, "package.json found.")

	// CHECK ALL NPM PACKAGES DEFINED IN package.json.
	for _, dep := range requiredDeps {
		//pm := determineNpmPackageManager()

		if !checkNpmPackage(ctx, pm, dep) {
			errors = append(errors, fmt.Sprintf("NPM package %s is missing. Run `%s install %s`.", dep, pm, dep))
		} else {
			successes = append(successes, fmt.Sprintf("NPM package %s is installed.", dep))
		}
	}

	return errors, warnings, successes
}

// CHECK IF AN NPM PACKAGE IS INSTALLED BY RUNNING `npm list`.
func checkNpmPackage(ctx context.Context, pm string, packageName string) bool {
	cmd := exec.CommandContext(ctx, pm, "list", packageName, "--depth=0")

	var outBuffer, errBuffer bytes.Buffer
	cmd.Stdout, cmd.Stderr = &outBuffer, &errBuffer

	if cmd.Run() == nil && !strings.Contains(errBuffer.String(), "missing") && strings.TrimSpace(outBuffer.String()) != "" {
		return true
	}

	return false
}
