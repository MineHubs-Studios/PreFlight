package core

import (
	"PreFlight/utils"
	"fmt"
	"os"
	"strings"
	"time"
)

type CheckResult struct {
	Scope     string
	Errors    []string
	Warnings  []string
	Successes []string
}

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
	var categorizedResults []CheckResult

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

		result := CheckResult{
			Scope:     utils.CapitalizeWords(module.Name()),
			Errors:    errors,
			Warnings:  warnings,
			Successes: successes,
		}

		categorizedResults = append(categorizedResults, result)
	}

	fmt.Println()

	printResults(categorizedResults)
	finalMessage(categorizedResults)
}

func printResults(results []CheckResult) {
	for _, result := range results {
		fmt.Println(Bold + "\nScope: " + result.Scope + Reset)

		if len(result.Successes) > 0 {
			fmt.Println(Green + "  Successes:" + Reset)
			for _, msg := range result.Successes {
				fmt.Println(Green + "    " + CheckMark + " " + msg + Reset)
			}

			fmt.Println()
		}

		if len(result.Warnings) > 0 {
			fmt.Println(Yellow + "  Warnings:" + Reset)

			for _, msg := range result.Warnings {
				fmt.Println(Yellow + "    " + WarningSign + " " + msg + Reset)
			}

			fmt.Println()
		}

		if len(result.Errors) > 0 {
			fmt.Println(Red + "  Errors:" + Reset)

			for _, msg := range result.Errors {
				fmt.Println(Red + "    " + CrossMark + " " + msg + Reset)
			}

			fmt.Println()
		}
	}
}

func finalMessage(results []CheckResult) {
	var totalErrors, totalWarnings int

	for _, result := range results {
		totalErrors += len(result.Errors)
		totalWarnings += len(result.Warnings)
	}

	var finalMessage string
	var exitCode int

	if totalErrors > 0 {
		finalMessage = Bold + Red + "System setup check completed. Resolve the above issues before proceeding." + Reset
		exitCode = 1
	} else if totalWarnings > 0 {
		finalMessage = Bold + Yellow + "System setup check completed with warnings. Review them before proceeding." + Reset
		exitCode = 0
	} else {
		finalMessage = Bold + Green + "System setup check completed successfully! All required tools and configurations are in place." + Reset
		exitCode = 0
	}

	fmt.Println("\n" + finalMessage)
	os.Exit(exitCode)
}
