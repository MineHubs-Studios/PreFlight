package cmd

import (
	"PreFlight/core"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	version   = "0.1.0"
	buildDate = "28-02-2025"
)

// versionCmd REPRESENTS THE VERSION COMMAND.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		fmt.Println(core.Bold + core.Yellow + "PreFlight - Version Information" + core.Reset + core.Bold)
		fmt.Println("-----------------------------")
		fmt.Printf("Version:    %s\n", version)
		fmt.Printf("Build dato: %s\n", buildDate)
		fmt.Println("-----------------------------" + core.Reset)
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
