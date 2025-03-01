package cmd

import (
	"PreFlight/core"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

var (
	packageManagersList string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all required dependencies for this project\n",
	Run: func(cmd *cobra.Command, args []string) {
		var pms []string

		if packageManagersList != "" {
			pms = strings.Split(packageManagersList, ",")
			hasNodePM := false

			for _, pm := range pms {
				if pm == "npm" || pm == "pnpm" {
					if hasNodePM {
						fmt.Println(core.Red + "Error: You can't use npm and pnpm at the same time.")
						return
					}

					hasNodePM = true
				}
			}
		}

		dependencies := core.GetAllDependencies(pms)
		core.PrintDependencies(dependencies)
	},
}

func init() {
	listCmd.Flags().StringVar(
		&packageManagersList,
		"pm",
		"",
		"Comma-separated list of package managers to list (composer,npm,pnpm)",
	)

	rootCmd.AddCommand(listCmd)
}
