package cmd

import (
	"PreFlight/core"
	"github.com/spf13/cobra"
	"strings"
)

var listPackageManagers string

// listCmd REPRESENTS THE LIST COMMAND THAT DISPLAYS DEPENDENCIES.
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all required dependencies for this project",
	Long:    `Lists all dependencies required by this project based on package manager configuration files.`,
	Example: "preflight list --pm=composer,go",
	Aliases: []string{"dependencies", "deps"},
	Run: func(_ *cobra.Command, _ []string) {
		var selectedPMs []string

		if listPackageManagers != "" {
			parts := strings.Split(listPackageManagers, ",")

			for _, p := range parts {
				name := strings.ToLower(strings.TrimSpace(p))

				switch name {
				case "npm", "pnpm", "yarn", "bun":
					name = "package"
				}

				if name != "" {
					selectedPMs = append(selectedPMs, name)
				}
			}
		}

		dependencies := core.GetAllDependencies(selectedPMs...)
		core.PrintDependencies(dependencies)
	},
}

func init() {
	listCmd.Flags().StringVar(
		&listPackageManagers,
		"pm",
		"",
		"Comma-separated list of package managers to list (composer,package,go,npm,yarn,pnpm,bun)",
	)

	rootCmd.AddCommand(listCmd)
}
