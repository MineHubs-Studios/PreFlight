package modules

import (
	"PreFlight/pm"
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

// CheckRequirements VERIFIES Node.js CONFIGURATIONS.
func (n NodeModule) CheckRequirements(ctx context.Context) (errors []string, warnings []string, successes []string) {
	// CHECK IF CONTEXT IS CANCELED.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	nodeVersion, err := getNodeVersion(ctx)

	// SKIP MODULE IF Node.js IS NOT INSTALLED.
	if err != nil {
		return nil, nil, nil
	}

	packageConfig := pm.LoadPackageConfig()

	if packageConfig.Error != nil {
		warnings = append(warnings, packageConfig.Error.Error())
		return errors, warnings, successes
	}

	// VALIDATE Node.js VERSION.
	if packageConfig.NodeVersion != "" {
		isValid, _ := utils.ValidateVersion(nodeVersion, packageConfig.NodeVersion)

		// use map for O(1) lookup instead of slice iteration
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

// getNodeVersion RETRIEVES THE INSTALLED Node.js VERSION.
func getNodeVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "node", "--version")
	output, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("failed to run node --version: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
