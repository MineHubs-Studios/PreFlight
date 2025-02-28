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

type ComposerModule struct{}

func (c ComposerModule) Name() string {
	return "Composer"
}

func (c ComposerModule) CheckRequirements(ctx context.Context, params map[string]interface{}) (errors []string, warnings []string, successes []string) {
	select {
	case <-ctx.Done():
		return nil, nil, nil
	default:
	}

	// CHECK IF COMPOSER IS INSTALLED.
	_ = isComposerInstalled(ctx, &errors, &successes)

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

	// CHECK COMPOSER DEPENDENCIES.
	for _, dep := range composerDeps {
		if !CheckComposerPackage(ctx, dep) {
			errors = append(errors, fmt.Sprintf("Composer package %s is missing. Run `composer require %s`.", dep, dep))
		} else {
			successes = append(successes, fmt.Sprintf("Composer package %s is installed.", dep))
		}
	}

	return errors, warnings, successes
}

func isComposerInstalled(ctx context.Context, errors *[]string, successes *[]string) string {
	cmd := exec.CommandContext(ctx, "composer", "--version")
	var outBuffer bytes.Buffer
	cmd.Stdout = &outBuffer

	err := cmd.Run()

	if err != nil {
		*errors = append(*errors, "Composer is not installed. Please install Composer.")
		return ""
	}

	version := strings.TrimSpace(outBuffer.String())
	versionParts := strings.Split(version, " ")

	if len(versionParts) >= 3 {
		composerVersion := versionParts[2]
		*successes = append(*successes, fmt.Sprintf("Composer is installed with version %s.", composerVersion))
		return composerVersion
	}

	return ""
}

// CheckComposerPackage CHECK IF A SPECIFIC COMPOSER PACKAGE IS INSTALLED.
func CheckComposerPackage(ctx context.Context, packageName string) bool {
	cmd := exec.CommandContext(ctx, "composer", "show", packageName)
	err := cmd.Run()

	if err != nil {
		return false
	}

	return true
}
