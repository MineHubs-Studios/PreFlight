package modules

import (
	"PreFlight/utils"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type PhpModule struct{}

func (p PhpModule) Name() string {
	return "PHP"
}

func (p PhpModule) CheckRequirements(ctx context.Context, params map[string]interface{}) (errors []string, warnings []string, successes []string) {
	select {
	case <-ctx.Done():
		return nil, nil, nil
	default:
	}

	// CHECK IF PHP IS INSTALLED.
	phpInstalled, phpVersion := isPhpInstalled(ctx)

	if phpInstalled {
		successes = append(successes, fmt.Sprintf("PHP is installed with version: %s.", phpVersion))
	} else {
		errors = append(errors, "PHP is not installed. Please install PHP.")
		return errors, warnings, successes
	}

	// READ PHP REQUIREMENTS FROM composer.json.
	phpVersionRequirement, requiredExtensions, _, found := utils.ReadComposerJSON()

	if !found {
		warnings = append(warnings, "composer.json not found. PHP requirements cannot be dynamically determined.")
		return errors, warnings, successes
	}

	// CHECK PHP VERSION AGAINST REQUIREMENT.
	if phpVersionRequirement != "" {
		if isValid, feedback := utils.ValidateVersion(phpVersion, phpVersionRequirement); isValid {
			successes = append(successes, fmt.Sprintf("Installed PHP version matches the required version: %s.", phpVersionRequirement))
		} else {
			errors = append(errors, feedback)
		}
	}

	// CHECK REQUIRED PHP EXTENSIONS.
	for _, ext := range requiredExtensions {
		if !CheckPHPExtension(ctx, ext) {
			errors = append(errors, fmt.Sprintf("PHP extension %s is missing. Please enable it.", ext))
		} else {
			successes = append(successes, fmt.Sprintf("PHP extension %s is installed.", ext))
		}
	}

	return errors, warnings, successes
}

// CHECK IF PHP IS INSTALLED AND RETURN THE VERSION IF AVAILABLE.
func isPhpInstalled(ctx context.Context) (bool, string) {
	cmd := exec.CommandContext(ctx, "php", "--version")
	output, err := cmd.Output()

	if err != nil {

		return false, ""
	}

	lines := strings.Split(string(output), "\n")

	if len(lines) > 0 {
		versionLine := lines[0]
		return true, versionLine
	}

	return true, ""
}

// CheckPHPExtension CHECK IF A SPECIFIC PHP EXTENSION IS INSTALLED.
func CheckPHPExtension(ctx context.Context, extension string) bool {
	cmd := exec.CommandContext(ctx, "php", "-m")
	output, err := cmd.Output()

	if err != nil {
		return false
	}

	for _, ext := range strings.Split(string(output), "\n") {
		if strings.TrimSpace(ext) == extension {
			return true
		}
	}

	return false
}
