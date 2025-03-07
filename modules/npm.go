package modules

import (
	"PreFlight/utils"
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

type PackageManager struct {
	Command  string // Command TO RUN (npm, pnpm, yarn)
	LockFile string // ASSOCIATED LOCK FILE.
}

func DeterminePackageManager() PackageManager {
	if _, err := os.Stat("pnpm-lock.yaml"); err == nil {
		return PackageManager{Command: "pnpm", LockFile: "pnpm-lock.yaml"}
	}

	if _, err := os.Stat("yarn.lock"); err == nil {
		return PackageManager{Command: "yarn", LockFile: "yarn.lock"}
	}

	if _, err := os.Stat("package-lock.json"); err == nil {
		return PackageManager{Command: "npm", LockFile: "package-lock.json"}
	}

	// DEFAULT TO NPM WITH NO LOCK FILE.
	return PackageManager{Command: "npm", LockFile: ""}
}

// CheckRequirements CHECK THE REQUIREMENTS FOR THE NPM MODULE.
func (n NpmModule) CheckRequirements(ctx context.Context, params map[string]interface{}) (errors []string, warnings []string, successes []string) {
	// CHECK IF CONTEXT IS CANCELED.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	// READ package.json TO EXTRACT DEPENDENCIES.
	_, requiredDeps, found := utils.ReadPackageJSON()

	// DETERMINE WHICH PACKAGE MANAGER TO USE.
	pm := DeterminePackageManager()

	// HANDLE MISSING package.json.
	if !found {
		errors = append(errors, "package.json not found.")

		if pm.LockFile != "" {
			warnings = append(warnings, fmt.Sprintf("package.json not found, but %s exists. Ensure package.json is included in your project.", pm.LockFile))
		} else {
			warnings = append(warnings, "Neither package.json nor lock files (package-lock.json, yarn.lock, pnpm-lock.yaml) are found.")
		}

		return errors, warnings, successes
	}

	successes = append(successes, "package.json found.")

	// GET ALL INSTALLED PACKAGES.
	installedPackages, err := getInstalledPackages(ctx, pm.Command)

	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Error getting installed packages: %v", err))
	}

	// CHECK REQUIRED PACKAGES.
	for _, dep := range requiredDeps {
		if isInstalled, exists := installedPackages[dep]; exists && isInstalled {
			successes = append(successes, fmt.Sprintf("NPM package %s is installed.", dep))
		} else {
			errors = append(errors, fmt.Sprintf("NPM package %s is missing. Run `%s install %s`.", dep, pm.Command, dep))
		}
	}

	return errors, warnings, successes
}

// getInstalledPackages RETURNS A MAP OF ALL INSTALLED PACKAGES.
func getInstalledPackages(ctx context.Context, pmCommand string) (map[string]bool, error) {
	cmd := exec.CommandContext(ctx, pmCommand, "list", "--depth=0", "--json")
	output, err := cmd.Output()

	installedPackages := make(map[string]bool)

	if err != nil {
		if len(output) == 0 {
			return installedPackages, fmt.Errorf("failed to list installed packages: %w", err)
		}
	}

	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" || !strings.Contains(line, "node_modules/") {
			continue
		}

		parts := strings.Split(line, "node_modules/")

		if len(parts) > 1 {
			packageName := strings.TrimSpace(parts[1])

			if strings.Contains(packageName, "/") && !strings.HasPrefix(packageName, "@") {
				packageName = strings.Split(packageName, "/")[0]
			}

			installedPackages[packageName] = true
		}
	}

	return installedPackages, nil
}
