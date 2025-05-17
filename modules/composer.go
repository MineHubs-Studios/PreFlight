package modules

import (
	"PreFlight/pm"
	"PreFlight/utils"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

type ComposerModule struct{}

func (c ComposerModule) Name() string {
	return "Composer"
}

// CheckRequirements VERIFIES Composer CONFIGURATIONS AND DEPENDENCIES.
func (c ComposerModule) CheckRequirements(ctx context.Context) (errors []string, warnings []string, successes []string) {
	// CHECK IF CONTEXT IS CANCELED.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	composerConfig := pm.LoadComposerConfig()
	pm := composerConfig.PackageManager

	if pm.LockFile == "" && !composerConfig.HasConfig {
		return nil, nil, nil
	}

	if !composerConfig.HasConfig {
		warnings = append(warnings, "composer.json not found.")

		if pm.LockFile != "" {
			warnings = append(warnings, fmt.Sprintf("composer.json not found, but %s exists. Ensure composer.json is included in your project.", pm.LockFile))
		}

		return errors, warnings, successes
	}

	composerVersion, err := GetComposerVersion(ctx)

	if err != nil {
		errors = append(errors, "Composer is not installed or not available in path.")
		return errors, warnings, successes
	}

	successes = append(successes, fmt.Sprintf("Installed %sComposer (%s).", utils.Reset, composerVersion))

	if !composerConfig.HasConfig && composerConfig.HasLock {
		warnings = append(warnings, "composer.lock exists without composer.json. Consider including composer.json.")
		return errors, warnings, successes
	}

	if composerConfig.Error != nil {
		errors = append(errors, fmt.Sprintf("Error reading composer.json: %v", composerConfig.Error))
		return errors, warnings, successes
	}

	successes = append(successes, "composer.json found.")
	installedDependencies := GetInstalledDependencies(ctx, composerConfig.Dependencies, composerConfig.DevDependencies)

	for _, dep := range append(composerConfig.Dependencies, composerConfig.DevDependencies...) {
		if version, exists := installedDependencies[dep]; exists {
			successes = append(successes, fmt.Sprintf("Installed dependency %s%s (%s).", utils.Reset, dep, version))
		} else {
			errors = append(errors, fmt.Sprintf("Missing dependency %s%s, Run `composer require %s`.", utils.Reset, dep, dep))
		}
	}

	return errors, warnings, successes
}

// GetComposerVersion RETRIEVES THE INSTALLED Composer VERSION.
func GetComposerVersion(ctx context.Context) (string, error) {
	output, err := utils.RunCommand(ctx, "composer", "--version")

	if err != nil {
		return "", err
	}

	parts := strings.Fields(strings.TrimSpace(output))

	if len(parts) >= 3 {
		return parts[2], nil
	}

	return "", fmt.Errorf("unexpected composer version format: %s", output)
}

// GetInstalledDependencies RETRIEVES THE INSTALLED Composer DEPENDENCIES.
func GetInstalledDependencies(ctx context.Context, dependencies, devDependencies []string) map[string]string {
	installedDependencies := make(map[string]string)
	allDeps := append(dependencies, devDependencies...)

	var wg sync.WaitGroup
	var mu sync.Mutex

	output, err := utils.RunCommand(ctx, "composer", "show", "--format=json")

	if err == nil {
		var data struct {
			Dependencies []struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"installed"`
		}

		if json.Unmarshal([]byte(output), &data) == nil {
			for _, dependency := range data.Dependencies {
				installedDependencies[dependency.Name] = dependency.Version
			}
		}
	}

	for _, dep := range allDeps {
		if _, exists := installedDependencies[dep]; exists {
			continue
		}

		wg.Add(1)

		go func(dep string) {
			defer wg.Done()
			output, err := utils.RunCommand(ctx, "composer", "show", dep)

			if err == nil {
				for _, line := range strings.Split(output, "\n") {
					line = strings.TrimSpace(line)

					if strings.HasPrefix(line, "versions :") || strings.HasPrefix(line, "version :") {
						parts := strings.SplitN(line, ":", 2)

						if len(parts) > 1 {
							version := strings.TrimSpace(strings.TrimPrefix(parts[1], "* "))

							mu.Lock()
							installedDependencies[dep] = version
							mu.Unlock()

							return
						}
					}
				}
			}

			mu.Lock()
			installedDependencies[dep] = "version unknown"
			mu.Unlock()
		}(dep)
	}

	wg.Wait()

	return installedDependencies
}
