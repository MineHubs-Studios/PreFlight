package cmd

import (
	"PreFlight/core"
	"PreFlight/modules"
	"PreFlight/utils"
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

// checkCmd represents the check command.
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks if all required dependencies are installed",
	Run: func(_ *cobra.Command, _ []string) {
		// Register available modules.
		availableModules := map[string]core.Module{
			"php":      modules.PhpModule{},
			"composer": modules.ComposerModule{},
			"node":     modules.NodeModule{},
			"package":  modules.PackageModule{},
			"go":       modules.GoModule{},
		}

		for name, module := range availableModules {
			core.RegisterAvailableModule(name, module)
		}

		aliasMap := map[string]string{
			"npm":  "package",
			"pnpm": "package",
			"yarn": "package",
			"bun":  "package",
		}

		// Process requested modules.
		var moduleNames []string

		if packageManagers != "" {
			for _, name := range strings.Split(packageManagers, ",") {
				normalized := strings.TrimSpace(strings.ToLower(name))

				if alias, ok := aliasMap[normalized]; ok {
					normalized = alias
				}

				if normalized != "" {
					moduleNames = append(moduleNames, normalized)
				}
			}
		}

		// Register requested modules.
		if err := core.RegisterModule(nil, moduleNames...); err != nil {
			fmt.Printf(utils.Red+"Failed to register modules: %v\n", err)
			return
		}

		// Setup context with timeout from a flag.
		timeout := time.Duration(timeoutSeconds) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		core.RunChecks(ctx)
	},
}

func init() {
	checkCmd.Flags().StringVar(
		&packageManagers,
		"pm",
		"",
		"Comma-separated list of package managers to check (php,composer,node,bun,pnpm,npm,yarn)",
	)

	checkCmd.Flags().UintVar(
		&timeoutSeconds,
		"timeout",
		300,
		"Timeout in seconds for all checks to complete",
	)

	rootCmd.AddCommand(checkCmd)
}
