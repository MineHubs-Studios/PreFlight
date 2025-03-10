package modules

import (
	"PreFlight/utils"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type ComposerModule struct{}

func (c ComposerModule) Name() string {
	return "Composer"
}

func (c ComposerModule) CheckRequirements(ctx context.Context) (errors []string, warnings []string, successes []string) {
	// CHECK IF CONTEXT IS CANCELED.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	// IF composer.json OR composer.lock IS NOT FOUND, THEN SKIP.
	if _, errJson := os.Stat("composer.json"); os.IsNotExist(errJson) {
		if _, errLock := os.Stat("composer.lock"); os.IsNotExist(errLock) {
			return nil, nil, nil
		}
	}

	// CHECK IF COMPOSER IS INSTALLED.
	composerVersion, err := getComposerVersion(ctx)

	if err != nil {
		errors = append(errors, "Composer is not installed or not available in path.")
		return errors, warnings, successes
	}

	successes = append(successes, fmt.Sprintf("Composer is installed with version %s.", composerVersion))

	// CHECK IF composer.json EXISTS.
	if _, err := os.Stat("composer.json"); os.IsNotExist(err) {
		warnings = append(warnings, "composer.lock exists without composer.json. Consider including composer.json.")
		return errors, warnings, successes
	}

	// READ composer.json TO EXTRACT REQUIRED NODE VERSION AND DEPENDENCIES.
	_, _, composerDeps, found := utils.ReadComposerJSON()

	if !found {
		errors = append(errors, "composer.json exists but could not be read. Please check file format.")
		return errors, warnings, successes
	}

	successes = append(successes, "composer.json found.")

	// CHECK Composer DEPENDENCIES.
	for _, dep := range composerDeps {
		if installed, err := isComposerPackageInstalled(ctx, dep); !installed {
			errorMsg := fmt.Sprintf("Composer package %s is missing. Run `composer require %s`.", dep, dep)

			if err != nil {
				errorMsg += fmt.Sprintf(" Error: %v", err)
			}

			errors = append(errors, errorMsg)
		} else {
			successes = append(successes, fmt.Sprintf("Composer package %s is installed.", dep))
		}
	}

	return errors, warnings, successes
}

// getComposerVersion RETURNS THE INSTALLED Composer VERSION OR AN ERROR.
func getComposerVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "composer", "--version")
	output, err := cmd.Output()

	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(output))
	versionParts := strings.Split(version, " ")

	if len(versionParts) >= 3 {
		return versionParts[2], nil
	}

	return "", fmt.Errorf("unexpected composer version format: %s", version)
}

// isComposerPackageInstalled CHECK IF A SPECIFIC COMPOSER PACKAGE IS INSTALLED.
func isComposerPackageInstalled(ctx context.Context, packageName string) (bool, error) {
	cmd := exec.CommandContext(ctx, "composer", "show", packageName)
	err := cmd.Run()

	if err != nil {
		return false, err
	}

	return true, nil
}
