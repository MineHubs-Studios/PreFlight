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

func (p PhpModule) IsApplicable(ctx context.Context) bool {
	if ctx.Err() != nil {
		return false
	}

	// CHECK IF PHP IS INSTALLED.
	_, err := getPhpVersion(ctx)

	if err == nil {
		return true
	}

	return false
}

func (p PhpModule) CheckRequirements(ctx context.Context, params map[string]interface{}) (errors []string, warnings []string, successes []string) {
	// CHECK IF CONTEXT IS CANCELED.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	phpVersion, err := getPhpVersion(ctx)
	successes = append(successes, fmt.Sprintf("PHP is installed with version: %s.", phpVersion))

	// READ PHP REQUIREMENTS FROM composer.json.
	phpVersionRequirement, requiredExtensions, _, _ := utils.ReadComposerJSON()

	// CHECK PHP VERSION AGAINST REQUIREMENT.
	if phpVersionRequirement != "" {
		isValid, feedback := utils.ValidateVersion(phpVersion, phpVersionRequirement)

		if isValid && feedback != "" {
			successes = append(successes, feedback)
		} else if !isValid {
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
