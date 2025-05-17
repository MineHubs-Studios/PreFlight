package modules

import (
	"PreFlight/pm"
	"PreFlight/utils"
	"context"
	"fmt"
	"strings"
)

type NodeModule struct{}

func (n NodeModule) Name() string {
	return "Node"
}

// CheckRequirements verifies Node.js configurations and dependencies.
func (n NodeModule) CheckRequirements(ctx context.Context) (errors []string, warnings []string, successes []string) {
	// Check if context is canceled.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	nodeVersion, err := getNodeVersion(ctx)

	// Skip this module if Node.js is not installed.
	if err != nil {
		return nil, nil, nil
	}

	packageConfig := pm.LoadPackageConfig()

	if packageConfig.Error != nil {
		warnings = append(warnings, packageConfig.Error.Error())
		return errors, warnings, successes
	}

	// Validate Node.js version.
	if packageConfig.NodeVersion != "" {
		isValid, _ := utils.ValidateVersion(nodeVersion, packageConfig.NodeVersion)

		eolVersions := map[string]bool{
			"10": true, "12": true, "14": true, "15": true,
			"16": true, "17": true, "18": true,
		}

		feedback := fmt.Sprintf("Installed %sNode.js (%s ⟶ required %s).", utils.Reset, nodeVersion, packageConfig.NodeVersion)
		versionPrefix := strings.TrimPrefix(strings.Split(nodeVersion, ".")[0], "v")

		if eolVersions[versionPrefix] {
			warnings = append(warnings, fmt.Sprintf("Installed %sNode.js (%s ⟶ End-of-Life), consider upgrading!", utils.Reset, nodeVersion))

			if isValid {
				warnings = append(warnings, feedback)
			}
		} else if !isValid {
			errors = append(errors, feedback)
		} else {
			successes = append(successes, feedback)
		}
	}

	return errors, warnings, successes
}

// getNodeVersion retrieves the installed Node.js version.
func getNodeVersion(ctx context.Context) (string, error) {
	output, err := utils.RunCommand(ctx, "node", "--version")

	if err != nil {
		return "", fmt.Errorf("failed to run node --version: %w", err)
	}

	return strings.TrimSpace(output), nil
}
