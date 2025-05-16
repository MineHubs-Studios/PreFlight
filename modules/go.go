package modules

import (
	"PreFlight/pm"
	"PreFlight/utils"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type GoModule struct{}

func (g GoModule) Name() string {
	return "Go"
}

// CheckRequirements verifies Go configurations and dependencies.
func (g GoModule) CheckRequirements(ctx context.Context) (errors []string, warnings []string, successes []string) {
	// Check if context is canceled.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	goVersion, err := getGoVersion(ctx)

	// Skip this module if Go is not installed.
	if err != nil {
		return nil, nil, nil
	}

	goConfig := pm.LoadGoConfig()

	if !goConfig.HasMod {
		return errors, warnings, successes
	}

	if goConfig.Error != nil {
		errors = append(errors, fmt.Sprintf("Error parsing go.mod: %v", goConfig.Error))
		return errors, warnings, successes
	}

	successes = append(successes, "go.mod found.")

	// VALIDATE Go VERSION.
	if goConfig.GoVersion != "" {
		isValid, _ := utils.ValidateVersion(goVersion, goConfig.GoVersion)

		eolVersions := map[string]bool{
			"1.12": true, "1.13": true, "1.14": true, "1.15": true,
			"1.16": true, "1.17": true, "1.18": true, "1.19": true,
			"1.20": true, "1.21": true, "1.22": true,
		}

		feedback := fmt.Sprintf("Installed %sGo (%s ⟶ required %s).", utils.Reset, goVersion, goConfig.GoVersion)
		versionPrefix := strings.Split(goVersion, ".")[0] + "." + strings.Split(goVersion, ".")[1]

		if eolVersions[versionPrefix] {
			warnings = append(warnings, fmt.Sprintf("Installed %sGo (%s ⟶ End-of-Life), consider upgrading!", utils.Reset, goVersion))

			if isValid {
				warnings = append(warnings, feedback)
			}
		} else if !isValid {
			errors = append(errors, feedback)
		} else {
			successes = append(successes, feedback)
		}
	} else {
		warnings = append(warnings, "Go version requirement not specified in go.mod.")
	}

	installedModules := getInstalledModules(ctx)

	for _, mod := range goConfig.Modules {
		if _, exists := installedModules[mod]; exists {
			successes = append(successes, fmt.Sprintf("Installed module %s%s.", utils.Reset, mod))
		} else {
			errors = append(errors, fmt.Sprintf("Missing module %s, Run 'go get %s'.", utils.Reset, mod))
		}
	}

	return errors, warnings, successes
}

// getGoVersion RETRIEVES THE INSTALLED Go VERSION.
func getGoVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "version")
	output, err := cmd.Output()

	if err != nil {
		return "", err
	}

	versionOutput := strings.TrimSpace(string(output))
	parts := strings.Split(versionOutput, " ")

	if len(parts) >= 3 {
		return strings.TrimPrefix(parts[2], "go"), nil
	}

	return "", fmt.Errorf("unexpected go version format: %s", versionOutput)
}

// getInstalledModules RETRIEVES THE INSTALLED Go MODULES.
func getInstalledModules(ctx context.Context) map[string]struct{} {
	modules := make(map[string]struct{})

	cmd := exec.CommandContext(ctx, "go", "list", "-m", "all")
	output, err := cmd.Output()

	if err != nil {
		return modules
	}

	for _, line := range strings.Split(string(output), "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			fields := strings.Fields(trimmed)

			if len(fields) > 0 {
				modules[fields[0]] = struct{}{}
			}
		}
	}

	return modules
}
