package cmd

import (
	"PreFlight/core"
	"context"
	"github.com/spf13/cobra"
)

var forceFix bool

var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Fix missing dependencies (Composer & npm)",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		core.FixDependencies(ctx, forceFix)
	},
}

func init() {
	fixCmd.Flags().BoolVarP(&forceFix, "force", "f", false, "Force reinstall dependencies")
	rootCmd.AddCommand(fixCmd)
}
