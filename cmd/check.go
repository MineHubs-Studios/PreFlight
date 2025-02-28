package cmd

import (
	"PreFlight/core"
	"PreFlight/modules"
	"context"
	"github.com/spf13/cobra"
	"time"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks if all required dependencies are installed",
	Run: func(cmd *cobra.Command, args []string) {
		err := core.RegisterModule(modules.PhpModule{})

		if err != nil {
			return
		}

		err = core.RegisterModule(modules.ComposerModule{})

		if err != nil {
			return
		}

		err = core.RegisterModule(modules.NpmModule{})

		if err != nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		core.RunChecks(ctx)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
