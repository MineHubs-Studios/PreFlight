package cmd

import (
	"PreFlight/core"
	"PreFlight/utils"
	"github.com/spf13/cobra"
	"runtime"
)

var (
	Version = "1.2.0"
)

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows PreFlight version information",
	Long:  `Shows detailed information about the PreFlight version including version number and build date.`,
	Run: func(_ *cobra.Command, _ []string) {
		ow := utils.NewOutputWriter()
		versionData, done := core.GetVersionInfo(Version, runtime.Version(), runtime.GOOS+"/"+runtime.GOARCH)

		ow.PrintNewLines(1)
		ow.Println(utils.Bold + utils.Cyan + "PreFlight - Version Information" + utils.Reset + utils.Bold)
		ow.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Wait for async tag fetch.
		<-done

		ow.Printf("Version:         %s\n", versionData.Version)

		if versionData.Error != nil {
			ow.Println("Latest version:  Unable to check")
		} else if versionData.Version == "development" || versionData.HasUpdate {
			ow.Printf("Latest version:  %s\n", versionData.LatestVersion)
		}

		ow.Printf("Go version:      %s\n", versionData.GoVersion)
		ow.Printf("Platform:        %s\n", versionData.Platform)
		ow.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
