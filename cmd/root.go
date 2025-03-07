package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd REPRESENTS THE BASE COMMAND WHEN THE PROGRAM IS EXECUTED WITHOUT SUBCOMMANDS.
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

// Execute ADDS ALL CHILD COMMANDS TO THE ROOT COMMAND AND SETS FLAGS APPROPRIATELY.
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

// init REGISTERS FLAGS FOR THE ROOT COMMAND.
func init() {
	// ENABLE SHELL COMPLETION.
	rootCmd.CompletionOptions.DisableDefaultCmd = false
}
