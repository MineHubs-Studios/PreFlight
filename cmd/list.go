package cmd

import (
	"PreFlight/core"
	"fmt"
	"github.com/spf13/cobra"
)

// listCmd REPRESENTS THE LIST COMMAND.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all required dependencies for this project\n",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(core.Bold + "🔍 Scanning project for dependencies...")

		dependencies := core.GetAllDependencies()

		fmt.Println(" ")
		fmt.Println(core.Bold + "Composer Dependencies:" + core.Reset)

		if len(dependencies.ComposerDeps) > 0 {
			for _, dep := range dependencies.ComposerDeps {
				fmt.Printf(core.Green+" "+core.CheckMark+" %s\n", dep+core.Reset)
			}
		} else {
			fmt.Println(core.Red + " " + core.CrossMark + " No Composer dependencies found!" + core.Reset)
		}

		fmt.Println("\n" + core.Bold + "NPM Dependencies:" + core.Reset)

		if len(dependencies.NpmDeps) > 0 {
			for _, dep := range dependencies.NpmDeps {
				fmt.Printf(core.Green+" "+core.CheckMark+" %s\n", dep+core.Reset)
			}
		} else {
			fmt.Println(core.Red + " " + core.CrossMark + " No NPM dependencies found!" + core.Reset)
		}

		fmt.Print(" ")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
