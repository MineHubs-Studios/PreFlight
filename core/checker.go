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

	bar := fmt.Sprintf("\r    [%s%s] %s%d%%%s",
		strings.Repeat(BlockFull, fullBlocks),
		strings.Repeat(BlockEmpty, emptyBlocks),
		Green,
		percent,
		Reset,
	)

	fmt.Printf("\r%s", bar)
}

func RunChecks() {
	var categorizedResults []CheckResult

	fmt.Println(Bold + "🚀 Running system setup checks...")

	for _, module := range RegisteredModules {
		fmt.Printf(Bold+"\n🔍 Running checks for module: %s\n", module.Name())

		for progress := 0; progress <= 100; progress += 25 {
			showProgress(progress)
			time.Sleep(200 * time.Millisecond)
		}

		showProgress(100)

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
	fmt.Println()

	printResults(categorizedResults)
	finalMessage(categorizedResults)
}

func printResults(results []CheckResult) {
	for _, result := range results {
		fmt.Println(Reset + Bold + "\nScope: " + result.Scope + Reset)

		if len(result.Successes) > 0 {
			fmt.Println(Green + "  Successes:" + Reset)
			printIndentedMessages(result.Successes, Green, CheckMark)
			fmt.Println()
		}

		if len(result.Warnings) > 0 {
			fmt.Println(Yellow + "  Warnings:" + Reset)
			printIndentedMessages(result.Warnings, Yellow, WarningSign)
			fmt.Println()
		}

		if len(result.Errors) > 0 {
			fmt.Println(Red + "  Errors:" + Reset)
			printIndentedMessages(result.Errors, Red, CrossMark)
			fmt.Println()
		}
	}
}

// printIndentedMessages håndterer indrykning af beskeder baseret på deres type.
func printIndentedMessages(messages []string, color string, symbol string) {
	for _, msg := range messages {
		indentLevel := 4

		if strings.Contains(strings.ToLower(msg), "composer package") || strings.Contains(strings.ToLower(msg), "npm package") {
			indentLevel = 6
		}

		fmt.Printf("%s%s %s %s\n", color, strings.Repeat(" ", indentLevel), symbol, msg)
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
