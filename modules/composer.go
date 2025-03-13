package modules

import (
	"PreFlight/config"
	"context"
	"fmt"
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

	composerConfig := config.LoadComposerConfig()

	// IF composer.json OR composer.lock IS NOT FOUND, THEN SKIP.
	if !composerConfig.HasJSON && !composerConfig.HasLock {
		return nil, nil, nil
	}

	composerVersion, err := GetComposerVersion(ctx)

	if err != nil {
		errors = append(errors, "Composer is not installed or not available in path.")
		return errors, warnings, successes
	}

	successes = append(successes, fmt.Sprintf("Composer is installed with version %s.", composerVersion))

	if !composerConfig.HasJSON && composerConfig.HasLock {
		warnings = append(warnings, "composer.lock exists without composer.json. Consider including composer.json.")
		return errors, warnings, successes
	}

	if composerConfig.Error != nil {
		errors = append(errors, fmt.Sprintf("Error reading composer.json: %v", composerConfig.Error))
		return errors, warnings, successes
	}

	if composerConfig.HasJSON {
		successes = append(successes, "composer.json found.")
	}

	composerDeps := append(composerConfig.Dependencies, composerConfig.DevDependencies...)

	for _, dep := range composerDeps {
		if installed, err := getInstalledPackage(ctx, dep); !installed {
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

// GetComposerVersion RETURNS THE INSTALLED Composer VERSION OR AN ERROR.
func GetComposerVersion(ctx context.Context) (string, error) {
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

// getInstalledPackage CHECK IF A SPECIFIC COMPOSER PACKAGE IS INSTALLED.
func getInstalledPackage(ctx context.Context, packageName string) (bool, error) {
	cmd := exec.CommandContext(ctx, "composer", "show", packageName)
	err := cmd.Run()

	if err != nil {
		return false, err
	}

	return true, nil
}
