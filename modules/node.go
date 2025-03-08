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

func (n NodeModule) IsApplicable(ctx context.Context) bool {
	if ctx.Err() != nil {
		return false
	}

	_, err := getNodeVersion(ctx)

	if err == nil {
		return true
	}

	return false
}

func (n NodeModule) CheckRequirements(ctx context.Context, params map[string]interface{}) (errors []string, warnings []string, successes []string) {
	// CHECK IF CONTEXT IS CANCELED.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	// CHECK IF Node.js IS INSTALLED AND GET THE VERSION.
	nodeVersion, _ := getNodeVersion(ctx)

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
