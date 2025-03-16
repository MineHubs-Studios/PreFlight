package modules

import (
	"PreFlight/config"
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

func (g GoModule) CheckRequirements(ctx context.Context) (errors []string, warnings []string, successes []string) {
	// CHECK IF CONTEXT IS CANCELED.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	goVersion, err := getGoVersion(ctx)

	// IF Go IS NOT INSTALLED, THEN SKIP.
	if err != nil {
		return nil, nil, nil
	}

	goConfig := config.LoadGoConfig()

	if !goConfig.HasMod {
		return errors, warnings, successes
	}

	if goConfig.Error != nil {
		errors = append(errors, fmt.Sprintf("Error parsing go.mod: %v", goConfig.Error))
		return errors, warnings, successes
	}

	successes = append(successes, "go.mod found.")

	if goConfig.GoVersion != "" {
		isValid, _ := utils.ValidateVersion(goVersion, goConfig.GoVersion)

		if isValid {
			eolVersions := []string{"1.12", "1.13", "1.14", "1.15", "1.16", "1.17", "1.18", "1.19", "1.20", "1.21", "1.22"}
			successes = append(successes, fmt.Sprintf("Installed %sGo (%s ⟶ required %s).", utils.Reset, goVersion, goConfig.GoVersion))

			for _, eolVersion := range eolVersions {
				if strings.HasPrefix(goVersion, eolVersion) {
					warnings = append(warnings, fmt.Sprintf("Installed %sGo (%s ⟶ End-of-Life), consider upgrading!", utils.Reset, goVersion))
				}
			}
		} else {
			errors = append(errors, fmt.Sprintf("Installed %sGo (%s ⟶ required %s).", utils.Reset, goVersion, goConfig.GoVersion))
		}
	} else {
		warnings = append(warnings, "Go version requirement not specified in go.mod.")
	}

	for _, mod := range goConfig.Modules {
		if getInstalledModules(ctx, mod) {
			successes = append(successes, fmt.Sprintf("Installed module %s%s", utils.Reset, mod))
		} else {
			errors = append(errors, fmt.Sprintf("Missing module %s , Run 'go get %s'.", mod, mod))
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

// getInstalledModules CHECKS IF A SPECIFIC MODULE IS INSTALLED.
func getInstalledModules(ctx context.Context, moduleName string) bool {
	cmd := exec.CommandContext(ctx, "go", "list", "-m", moduleName)
	err := cmd.Run()

	return err == nil
}
