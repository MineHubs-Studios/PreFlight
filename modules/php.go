package modules

import (
	"PreFlight/utils"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type PhpModule struct{}

func (p PhpModule) Name() string {
	return "PHP"
}

func (p PhpModule) CheckRequirements(ctx context.Context, params map[string]interface{}) (errors []string, warnings []string, successes []string) {
	// CHECK IF CONTEXT IS CANCELED.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	// CHECK IF PHP IS INSTALLED.
	phpVersion, err := getPhpVersion(ctx)

	if err != nil {
		errors = append(errors, "PHP is not installed. Please install PHP.")
		return errors, warnings, successes
	}

	successes = append(successes, fmt.Sprintf("PHP is installed with version: %s.", phpVersion))

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

	// GET ALL INSTALLED PHP EXTENSIONS.
	installedExtensions, err := getPhpExtensions(ctx)

	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to check PHP extensions: %v", err))
		return errors, warnings, successes
	}

	// CHECK REQUIRED PHP EXTENSIONS.
	for _, ext := range requiredExtensions {
		if _, exists := installedExtensions[ext]; !exists {
			errors = append(errors, fmt.Sprintf("PHP extension %s is missing. Please enable it.", ext))
		} else {
			successes = append(successes, fmt.Sprintf("PHP extension %s is installed.", ext))
		}
	}

	return errors, warnings, successes
}

// getPhpVersion RETURNS THE INSTALLED PHP VERSION OR AN ERROR.
func getPhpVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "php", "--version")
	output, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("failed to run php --version: %w", err)
	}

	lines := strings.Split(string(output), "\n")

	if len(lines) == 0 {
		return "", fmt.Errorf("unexpected output from php --version")
	}

	re := regexp.MustCompile(`PHP (\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(lines[0])

	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse PHP version from: %s", lines[0])
	}

	return matches[1], nil

	// TODO - DO WE WANT TO GET THIS DATA FOR USERS? (built: Nov 20 2024 11:13:22) (NTS Visual C++ 2022 x64)
}

// getPhpExtensions RETURNS A MAP OF ALL INSTALLED PHP EXTENSIONS.
func getPhpExtensions(ctx context.Context) (map[string]struct{}, error) {
	cmd := exec.CommandContext(ctx, "php", "-m")
	output, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("failed to run php -m: %w", err)
	}

	extensions := make(map[string]struct{})

	for _, ext := range strings.Split(string(output), "\n") {
		if trimmed := strings.TrimSpace(ext); trimmed != "" {
			extensions[trimmed] = struct{}{}
		}
	}

	return extensions, nil
}
