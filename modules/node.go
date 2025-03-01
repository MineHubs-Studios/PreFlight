package modules

import (
	"PreFlight/utils"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type NodeModule struct{}

func (n NodeModule) Name() string {
	return "Node"
}

func (n NodeModule) CheckRequirements(ctx context.Context, params map[string]interface{}) (errors []string, warnings []string, successes []string) {
	select {
	case <-ctx.Done():
		return nil, nil, nil
	default:
	}

	// CHECK IF NODE IS INSTALLED.
	nodeVersionOutput := isNodeInstalled(ctx, &errors, &successes)

	// VALIDATE NODE VERSION IF SPECIFIC VERSION IS REQUIRED.
	nodeVersion, _, found := utils.ReadPackageJSON()

	if found && nodeVersion != "" {
		if isValid, feedback := utils.ValidateVersion(nodeVersionOutput, nodeVersion); isValid {
			successes = append(successes, feedback)
		} else {
			errors = append(errors, feedback)
		}
	}

	return errors, warnings, successes
}

// VALIDATE NODE INSTALLATION AND OBTAIN INSTALLED NODE VERSION.
func isNodeInstalled(ctx context.Context, errors *[]string, successes *[]string) string {
	cmd := exec.CommandContext(ctx, "node", "--version")
	var outBuffer bytes.Buffer
	cmd.Stdout = &outBuffer

	err := cmd.Run()

	if err != nil {
		*errors = append(*errors, "Node.js is not installed. Please install Node.js to use NPM.")
		return ""
	}

	installedVersion := strings.TrimSpace(outBuffer.String())
	*successes = append(*successes, fmt.Sprintf("Node.js is installed with version %s.", installedVersion))

	return installedVersion
}
