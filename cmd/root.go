package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "PreFlight",
	Short: "PreFlight is a CLI tool for checking project dependencies",
	Long: `PreFlight helps developers check, install, and manage 
project dependencies dynamically based on configuration files.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {}
