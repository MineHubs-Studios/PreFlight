package utils

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// RunCommand runs a CLI command and returns its trimmed stdout output or an error.
func RunCommand(ctx context.Context, name string, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run '%s %s': %w — %s", name, strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}
