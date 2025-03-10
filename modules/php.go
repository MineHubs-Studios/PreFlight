package modules

import (
	"PreFlight/config"
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

func (p PhpModule) CheckRequirements(ctx context.Context) (errors []string, warnings []string, successes []string) {
	// CHECK IF CONTEXT IS CANCELED.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	phpVersion, err := getPhpVersion(ctx)

	// IF PHP IS NOT INSTALLED, THEN SKIP.
	if err != nil {
		return nil, nil, nil
	}

	successes = append(successes, fmt.Sprintf("PHP is installed with version: %s.", phpVersion))

	composerConfig := config.LoadComposerConfig()

	// IF composer.json IS NOT FOUND, THEN SKIP.
	if !composerConfig.HasJSON {
		warnings = append(warnings, "composer.json file not found.")
		return errors, warnings, successes
	}

	if composerConfig.Error != nil {
		errors = append(errors, fmt.Sprintf("Failed to read composer.json: %v", composerConfig.Error))
		return errors, warnings, successes
	}

	// CHECK PHP VERSION AGAINST REQUIREMENT.
	if composerConfig.PHPVersion != "" {
		isValid, feedback := utils.ValidateVersion(phpVersion, composerConfig.PHPVersion)

		if isValid && feedback != "" {
			successes = append(successes, feedback)
		} else if !isValid {
			errors = append(errors, feedback)
		}
	}

	if len(composerConfig.PHPExtensions) > 0 {
		// GET ALL INSTALLED PHP EXTENSIONS.
		PHPExtensions, err := getPhpExtensions(ctx)

		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to check PHP extensions: %v", err))
			return errors, warnings, successes
		}

		// CHECK REQUIRED PHP EXTENSIONS.
		for _, ext := range composerConfig.PHPExtensions {
			if _, exists := PHPExtensions[ext]; !exists {
				errors = append(errors, fmt.Sprintf("PHP extension %s is missing. Please enable it.", ext))
			} else {
				successes = append(successes, fmt.Sprintf("PHP extension %s is installed.", ext))
			}
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

	regex := regexp.MustCompile(`PHP (\d+\.\d+\.\d+)`)
	matches := regex.FindStringSubmatch(lines[0])

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

	PHPExtensions := make(map[string]struct{})

	for _, ext := range strings.Split(string(output), "\n") {
		if trimmed := strings.TrimSpace(ext); trimmed != "" {
			PHPExtensions[trimmed] = struct{}{}
		}
	}

	return PHPExtensions, nil
}
