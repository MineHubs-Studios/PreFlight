package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "PreFlight",
	Short: "PreFlight is a CLI tool for checking project dependencies",
	Long: `PreFlight helps developers check, install, and manage
project dependencies dynamically based on configuration files.

It supports multiple package managers such as composer, npm, pnpm and yarn.`,
	Example:       "preflight check --pm=npm,composer",
	Version:       fmt.Sprintf("%s (built: %s)", "1.0.0", "Unknown"),
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Error during execution: %s\n", err)

		if err != nil {
			return err
		}

		os.Exit(1)
	}

	return nil
}

func init() {
	// Enable shell completion.
	rootCmd.CompletionOptions.DisableDefaultCmd = false
}
