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

func (c ComposerModule) IsApplicable(ctx context.Context) bool {
	if ctx.Err() != nil {
		return false
	}

	// CHECK IF COMPOSER IS INSTALLED.
	_, err := getComposerVersion(ctx)

	if err == nil {
		return true
	}

	// CHECK IF COMPOSER.JSON OR COMPOSER.LOCK EXISTS.
	if _, err := os.Stat("composer.json"); err == nil {
		return true
	}

	if _, err := os.Stat("composer.lock"); err == nil {
		return true
	}

	return false
}

func (c ComposerModule) CheckRequirements(ctx context.Context, params map[string]interface{}) (errors []string, warnings []string, successes []string) {
	// CHECK IF CONTEXT IS CANCELED.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	// CHECK IF Composer.js IS INSTALLED AND GET THE VERSION.
	composerVersion, _ := getComposerVersion(ctx)

	successes = append(successes, fmt.Sprintf("Composer is installed with version %s.", composerVersion))

	// READ composer.json TO EXTRACT REQUIRED NODE VERSION AND DEPENDENCIES.
	_, _, composerDeps, found := utils.ReadComposerJSON()

	// HANDLE MISSING composer.json.
	if !found {
		errors = append(errors, "composer.json not found.")

		// CHECK FOR composer.lock IF composer.json IS MISSING.
		if _, err := os.Stat("composer.lock"); err == nil {
			warnings = append(warnings, "composer.lock exists. Ensure composer.json is included.")
		} else {
			warnings = append(warnings, "No composer.lock found.")
		}

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
