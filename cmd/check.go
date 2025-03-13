package cmd

import (
	"PreFlight/core"
	"PreFlight/modules"
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

var (
	packageManagers string
	timeoutSeconds  uint
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks if all required dependencies are installed",
	Run: func(_ *cobra.Command, args []string) {
		// REGISTER ALL AVAILABLE MODULES.
		availableModules := map[string]core.Module{
			"php":      modules.PhpModule{},
			"composer": modules.ComposerModule{},
			"node":     modules.NodeModule{},
			"npm":      modules.NpmModule{},
			"go":       modules.GoModule{},
		}

		for name, module := range availableModules {
			core.RegisterAvailableModule(name, module)
		}

		// PROCESS REQUESTED MODULES.
		var moduleNames []string

		if packageManagers != "" {
			// SPLIT, TRIM AND VALIDATE MODULE NAMES.
			for _, name := range strings.Split(packageManagers, ",") {
				name = strings.TrimSpace(strings.ToLower(name))

				if name != "" {
					moduleNames = append(moduleNames, name)
				}
			}
		}

		// REGISTER REQUESTED MODULES.
		if err := core.RegisterModule(nil, moduleNames...); err != nil {
			fmt.Printf(core.Red+"Failed to register modules: %v\n", err)
			return
		}

		// SETUP CONTEXT WITH TIMEOUT FROM FLAG.
		timeout := time.Duration(timeoutSeconds) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// RUN THE CHECKS.
		core.RunChecks(ctx)
	},
}

func init() {
	// DEFINE FLAGS FOR CHECK COMMAND.
	checkCmd.Flags().StringVar(
		&packageManagers,
		"pm",
		"",
		"Comma-separated list of package managers to check (php,composer,node,npm)",
	)

	checkCmd.Flags().UintVar(
		&timeoutSeconds,
		"timeout",
		300,
		"Timeout in seconds for all checks to complete",
	)

	rootCmd.AddCommand(checkCmd)
}
