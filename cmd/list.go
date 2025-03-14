package cmd

import (
	"PreFlight/core"
	"PreFlight/utils"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

// COMMAND-LINE FLAGS.
var (
	packageManagersList string
)

// listCmd REPRESENTS THE LIST COMMAND THAT DISPLAYS DEPENDENCIES.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all required dependencies for this project",
	Long: `Lists all dependencies required by this project based on package manager configuration files.
Multiple package managers can be specified except that npm and pnpm cannot be used simultaneously.`,
	Example: "preflight list --pm composer,npm",
	Aliases: []string{"dependencies", "deps"},
	Run: func(_ *cobra.Command, _ []string) {
		var packageManagers []string

		if packageManagersList != "" {
			// PROCESS AND VALIDATE PACKAGE MANAGER NAMES.
			hasNodePM := false

			for _, pm := range strings.Split(packageManagersList, ",") {
				pm = strings.TrimSpace(strings.ToLower(pm))

				if pm == "" {
					continue
				}

				// CHECK FOR NPM AND PNPM CONFLICT.
				if pm == "npm" || pm == "pnpm" || pm == "yarn" {
					if hasNodePM {
						fmt.Printf(utils.Red+"%sError: You can't use npm, pnpm and yarn at the same time.%s\n",
							utils.Red, utils.Reset)
						return
					}

					hasNodePM = true
				}

				packageManagers = append(packageManagers, pm)
			}
		}

		// GET AND PRINT DEPENDENCIES.
		dependencies := core.GetAllDependencies(packageManagers)

		core.PrintDependencies(dependencies)
	},
}

func init() {
	listCmd.Flags().StringVar(
		&packageManagersList,
		"pm",
		"",
		"Comma-separated list of package managers to list (composer,npm,pnpm)",
	)

	rootCmd.AddCommand(listCmd)
}
