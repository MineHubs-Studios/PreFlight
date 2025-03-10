package modules

import (
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

	// CHECK IF A SPECIFIC Node.js VERSION IS REQUIRED.
	requiredVersion, _, found := utils.ReadPackageJSON()

	if found && requiredVersion != "" {
		if isValid, feedback := utils.ValidateVersion(nodeVersion, requiredVersion); isValid {
			successes = append(successes, feedback)
		} else {
			errors = append(errors, feedback)
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
