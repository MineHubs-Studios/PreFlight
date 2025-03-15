package cmd

import (
	"PreFlight/core"
	"github.com/spf13/cobra"
)

// listCmd REPRESENTS THE LIST COMMAND THAT DISPLAYS DEPENDENCIES.
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all required dependencies for this project",
	Long:    `Lists all dependencies required by this project based on package manager configuration files.`,
	Example: "preflight list",
	Aliases: []string{"dependencies", "deps"},
	Run: func(_ *cobra.Command, _ []string) {
		// GET AND PRINT DEPENDENCIES.
		dependencies := core.GetAllDependencies()
		core.PrintDependencies(dependencies)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
