package cmd

import (
	"PreFlight/core"
	"PreFlight/modules"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks if all required dependencies are installed",
	Run: func(cmd *cobra.Command, args []string) {
		core.RegisterModule(modules.ComposerModule{})
		core.RegisterModule(modules.NpmModule{})

		core.RunChecks()
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
