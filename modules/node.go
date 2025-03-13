package modules

import (
	"PreFlight/config"
	"PreFlight/utils"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type NodeModule struct{}

func (n NodeModule) Name() string {
	return "Node"
}

func (n NodeModule) CheckRequirements(ctx context.Context) (errors []string, warnings []string, successes []string) {
	// CHECK IF CONTEXT IS CANCELED.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	nodeVersion, err := getNodeVersion(ctx)

	// IF node.js IS NOT INSTALLED, THEN SKIP.
	if err != nil {
		return nil, nil, nil
	}

	successes = append(successes, fmt.Sprintf("Node.js is installed with version %s.", nodeVersion))

	packageConfig := config.LoadPackageConfig()

	if packageConfig.Error != nil {
		warnings = append(warnings, packageConfig.Error.Error())
		return errors, warnings, successes
	}

	if packageConfig.NodeVersion != "" {
		isValid, feedback := utils.ValidateVersion(nodeVersion, packageConfig.NodeVersion)

		if isValid {
			successes = append(successes, feedback)
		} else {
			errors = append(errors, feedback)
		}
	}

	// CHECK FOR EOL NODE VERSIONS.
	eolVersions := []string{"10.", "12.", "14.", "15.", "16.", "17.", "18."}

	for _, eolVersion := range eolVersions {
		if strings.HasPrefix(nodeVersion, "v"+eolVersion) {
			warnings = append(warnings, fmt.Sprintf("Detected Node.js version %s is End-of-Life (EOL). Consider upgrading!", nodeVersion))
		}
	}

	return errors, warnings, successes
}

// getNodeVersion RETURNS THE INSTALLED Node.js VERSION OR AN ERROR.
func getNodeVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "node", "--version")
	output, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("failed to run node --version: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
