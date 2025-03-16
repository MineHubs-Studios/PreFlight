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

	packageConfig := config.LoadPackageConfig()

	if packageConfig.Error != nil {
		warnings = append(warnings, packageConfig.Error.Error())
		return errors, warnings, successes
	}

	if packageConfig.NodeVersion != "" {
		isValid, _ := utils.ValidateVersion(nodeVersion, packageConfig.NodeVersion)

		if isValid {
			eolVersions := []string{"10.", "12.", "14.", "15.", "16.", "17.", "18."}
			successes = append(successes, fmt.Sprintf("Installed %sNode.js (%s ⟶ required %s).", utils.Reset, nodeVersion, packageConfig.NodeVersion))

			for _, eolVersion := range eolVersions {
				if strings.HasPrefix(nodeVersion, "v"+eolVersion) {
					warnings = append(warnings, fmt.Sprintf("Installed %sNode.js (%s ⟶ End-of-Life), Consider upgrading!", utils.Reset, nodeVersion))
				}
			}
		} else {
			errors = append(errors, fmt.Sprintf("Installed %sNode.js (%s ⟶ required %s).", utils.Reset, nodeVersion, packageConfig.NodeVersion))
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
