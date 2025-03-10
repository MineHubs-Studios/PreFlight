package modules

import (
	"PreFlight/utils"
	"context"
	"fmt"
	"os"
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

	// IF go IS NOT INSTALLED, THEN SKIP.
	if err != nil {
		return nil, nil, nil
	}

	successes = append(successes, fmt.Sprintf("Go is installed with version %s.", goVersion))

	// IF go.mod OR composer.lock IS NOT FOUND, THEN SKIP.
	if _, errJson := os.Stat("go.mod"); os.IsNotExist(errJson) {
		return errors, warnings, successes
	}

	successes = append(successes, "go.mod found.")

	requiredGoVersion := getGoVersionRequirement()

	// CHECK go VERSION.
	if requiredGoVersion != "" {
		isValid, feedback := utils.ValidateVersion(goVersion, requiredGoVersion)

		if !isValid {
			errors = append(errors, feedback)
		}
	} else {
		warnings = append(warnings, "Go version requirement not specified in go.mod.")
	}

	// READ go.mod TO FIND REQUIREMENTS.
	requiredModules, err := GetRequiredGoModules()

	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Could not parse dependencies: %v", err)) // SILENT THIS AND ONLY CHECK DEPENDENCIES IF go.mod is found!
	}

	// CHECK IF REQUIRED MODULES ARE INSTALLED.
	for _, module := range requiredModules {
		if isGoModuleInstalled(ctx, module) {
			successes = append(successes, fmt.Sprintf("Go module %s is installed.", module))
		} else {
			errors = append(errors, fmt.Sprintf("Go module %s is missing. Run 'go get %s'.", module, module))
		}
	}

	return errors, warnings, successes
}

// getGoVersion RETURNS THE INSTALLED GO VERSION OR AN ERROR.
func getGoVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "version")
	output, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("failed to run 'go version': %w", err)
	}

	// EXTRACT VERSION FROM OUTPUT (FORMAT: "go version go1.18.3 darwin/amd64").
	versionOutput := strings.TrimSpace(string(output))
	parts := strings.Split(versionOutput, " ")

	if len(parts) >= 3 {
		return strings.TrimPrefix(parts[2], "go"), nil
	}

	return versionOutput, nil
}

// getGoVersionRequirement RETURNS THE GO VERSION REQUIREMENT. dd
func getGoVersionRequirement() string {
	// READ go.mod FILE.
	content, err := os.ReadFile("go.mod")

	if err != nil {
		fmt.Println("Could not read go.mod file:", err)

		return ""
	}

	fileContent := string(content)
	lines := strings.Split(fileContent, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "go ") {
			version := strings.TrimPrefix(line, "go ")

			return version
		}
	}

	return ""
}

// GetRequiredGoModules RETURNS A LIST OF REQUIRED GO MODULES.
func GetRequiredGoModules() ([]string, error) {
	// RUN 'go list -m all' TO GET A LIST OF ALL DEPENDENCIES.
	cmd := exec.Command("go", "list", "-m", "all")
	output, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("failed to run 'go list -m all'")
	}

	lines := strings.Split(string(output), "\n")
	var modules []string

	// START FROM 1 TO SKIP THE CURRENT MODULE.
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if line == "" {
			continue
		}

		// EXTRACT MODULE NAME (BEFORE ANY VERSION NUMBERS).
		parts := strings.Fields(line)

		if len(parts) > 0 && parts[0] != "" {
			modules = append(modules, parts[0])
		}
	}

	return modules, nil
}

// isGoModuleInstalled CHECKS IF A SPECIFIC MODULE IS INSTALLED.
func isGoModuleInstalled(ctx context.Context, moduleName string) bool {
	if moduleName == "" {
		return false
	}

	cmd := exec.CommandContext(ctx, "go", "list", "-m", moduleName)
	err := cmd.Run()

	return err == nil
}
