package cmd

import (
	"PreFlight/core"
	"PreFlight/utils"
	"github.com/spf13/cobra"
	"runtime"
)

var (
	// Version SPECIFY THE CURRENT VERSION OF PreFlight.
	Version = "2.0.0-beta2"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows PreFlight version information",
	Long:  `Shows detailed information about the PreFlight version including version number and build date.`,
	Run: func(_ *cobra.Command, _ []string) {
		ow := utils.NewOutputWriter()

		// GET VERSION DATA.
		versionData, done := core.GetVersionInfo(
			Version,
			runtime.Version(),
			runtime.GOOS+"/"+runtime.GOARCH,
		)

		ow.PrintNewLines(1)
		ow.Println(core.Bold + core.Cyan + "PreFlight - Version Information" + core.Reset + core.Bold)
		ow.Println(core.Border)

		// WAIT FOR THE ASYNC OPERATION TO COMPLETE.
		<-done

		if versionData.Version == "development" {
			ow.Printf("Version:         %s\n", versionData.Version)

			if versionData.Error == nil {
				ow.Printf("Latest version:  %s (GitHub)\n", versionData.LatestVersion)
			} else {
				ow.Printf("Latest version:  Unable to check\n")
			}
		} else {
			ow.Printf("Version:         %s\n", versionData.Version)

			if versionData.HasUpdate {
				ow.Printf("Latest version:  %s\n", versionData.LatestVersion)
			}
		}

		// ALWAYS SHOW Go VERSION AND PLATFORM.
		ow.Printf("Go version:      %s\n", versionData.GoVersion)
		ow.Printf("Platform:        %s\n", versionData.Platform)
		ow.Println(core.Border + core.Reset)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
