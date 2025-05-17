package cmd

import (
	"PreFlight/core"
	"context"
	"github.com/spf13/cobra"
)

var forceFix bool

// fixCmd represents the fix command.
var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Fix missing dependencies (Composer & npm)",
	Run: func(_ *cobra.Command, _ []string) {
		ctx := context.Background()
		core.FixDependencies(ctx, forceFix)
	},
}

func init() {
	fixCmd.Flags().BoolVarP(&forceFix, "force", "f", false, "Force reinstall dependencies")
	rootCmd.AddCommand(fixCmd)
}
