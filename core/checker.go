package core

import (
	"fmt"
	"os"
	"strings"
	"time"
)

var RegisteredModules []Module

func RegisterModule(module Module) {
	RegisteredModules = append(RegisteredModules, module)
}

func showProgress(percent int) {
	fullBlocks := percent / 5
	emptyBlocks := 20 - fullBlocks

	bar := fmt.Sprintf("\r[%s%s] %d%%",
		strings.Repeat(BlockFull, fullBlocks),
		strings.Repeat(BlockEmpty, emptyBlocks),
		percent,
	)

	fmt.Printf("\r%s", bar)
}

func RunChecks() {
	var totalErrors, totalWarnings, totalSuccesses []string

	fmt.Println("Running system setup checks...")
	fmt.Println()

	for _, module := range RegisteredModules {
		fmt.Printf(Bold+"\n🔍 Running checks for module: %s\n", module.Name())

		for progress := 0; progress <= 100; progress += 25 {
			showProgress(progress)
			time.Sleep(200 * time.Millisecond)
		}

		showProgress(100)
		fmt.Println()

		errors, warnings, successes := module.CheckRequirements(map[string]interface{}{
			"environment": "production",
		})

		totalErrors = append(totalErrors, errors...)
		totalWarnings = append(totalWarnings, warnings...)
		totalSuccesses = append(totalSuccesses, successes...)
	}

	fmt.Println()

	printResults(totalErrors, totalWarnings, totalSuccesses)

	finalMessageAndExit(totalErrors, totalWarnings)
}

func printResults(errors []string, warnings []string, successes []string) {
	if len(successes) > 0 {
		fmt.Println("\n" + Bold + Green + "Successes:" + Reset)

		for _, msg := range successes {
			fmt.Println(Green + "  " + CheckMark + " " + msg + Reset)
		}
	}

	if len(warnings) > 0 {
		fmt.Println("\n" + Bold + Yellow + "Warnings:" + Reset)

		for _, msg := range warnings {
			fmt.Println(Yellow + "  " + WarningSign + " " + msg + Reset)
		}
	}

	if len(errors) > 0 {
		fmt.Println("\n" + Bold + Red + "Errors:" + Reset)

		for _, msg := range errors {
			fmt.Println(Red + "  " + CrossMark + " " + msg + Reset)
		}
	}
}

func finalMessageAndExit(errors []string, warnings []string) {
	var finalMessage string
	var exitCode int

	if len(errors) > 0 {
		finalMessage = Bold + Red + "System setup check completed. Resolve the above issues before proceeding." + Reset
		exitCode = 1
	} else if len(warnings) > 0 {
		finalMessage = Bold + Yellow + "System setup check completed with warnings. Review them before proceeding." + Reset
		exitCode = 0
	} else {
		finalMessage = Bold + Green + "System setup check completed successfully! All required tools and configurations are in place." + Reset
		exitCode = 0
	}

	fmt.Println("\n" + finalMessage)
	os.Exit(exitCode)
}
