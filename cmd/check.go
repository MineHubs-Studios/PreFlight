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
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks if all required dependencies are installed",
	Run: func(cmd *cobra.Command, args []string) {
		core.RegisterAvailableModule("php", modules.PhpModule{})
		core.RegisterAvailableModule("composer", modules.ComposerModule{})
		core.RegisterAvailableModule("node", modules.NodeModule{})
		core.RegisterAvailableModule("npm", modules.NpmModule{})

		var moduleNames []string

		if packageManagers != "" {
			moduleNames = strings.Split(packageManagers, ",")
		}

		if err := core.RegisterModule(nil, moduleNames...); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		core.RunChecks(ctx)
	},
}

func init() {
	checkCmd.Flags().StringVar(
		&packageManagers,
		"pm",
		"",
		"Comma-separated list of package managers to check (php,composer,npm)",
	)
	rootCmd.AddCommand(checkCmd)
}
