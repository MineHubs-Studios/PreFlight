package core

import (
	"context"
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

func showProgress(percent int) {
	fullBlocks := percent / 5
	emptyBlocks := 20 - fullBlocks

	var sb strings.Builder

	sb.WriteString("\r    [")
	sb.WriteString(strings.Repeat(BlockFull, fullBlocks))
	sb.WriteString(strings.Repeat(BlockEmpty, emptyBlocks))
	sb.WriteString("] ")
	sb.WriteString(Green)
	sb.WriteString(fmt.Sprintf("%d", percent))
	sb.WriteString("%")
	sb.WriteString(Reset)

	fmt.Print(sb.String())
}

func RunChecks(ctx context.Context) {
	var categorizedResults []CheckResult

	fmt.Println(Bold + "🚀 Running system setup checks...")

	for _, module := range GetModules() {
		// CHECK FOR CANCELLATION.
		select {
		case <-ctx.Done():
			fmt.Println("\nChecks cancelled...")
			return
		default:
		}

		moduleStart := time.Now()
		fmt.Printf(Bold+"\n🔍 Running checks for module: %s\n", module.Name())

		for progress := 0; progress <= 100; progress += 25 {
			select {
			case <-ctx.Done():
				return
			default:
				showProgress(progress)
				time.Sleep(200 * time.Millisecond)
			}
		}

		showProgress(100)

		errors, warnings, successes := module.CheckRequirements(ctx, map[string]interface{}{
			"environment": "production",
		})

		result := CheckResult{
			Scope:     module.Name(),
			Errors:    errors,
			Warnings:  warnings,
			Successes: successes,
		}

		categorizedResults = append(categorizedResults, result)

		moduleDuration := time.Since(moduleStart)
		fmt.Printf("\n      %s⏱ Completed in: %dms%s\n", Yellow, moduleDuration.Milliseconds(), Reset)
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
			printMessages(result.Successes, Green, CheckMark)
			fmt.Println()
		}

		if len(result.Warnings) > 0 {
			fmt.Println(Yellow + "  Warnings:" + Reset)
			printMessages(result.Warnings, Yellow, WarningSign)
			fmt.Println()
		}

		if len(result.Errors) > 0 {
			fmt.Println(Red + "  Errors:" + Reset)
			printMessages(result.Errors, Red, CrossMark)
			fmt.Println()
		}
	}
}

// printMessages
func printMessages(messages []string, color string, symbol string) {
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
		finalMessage = Bold + Red + "System setup check completed, resolve the above issues before proceeding." + Reset
		exitCode = 1
	} else if totalWarnings > 0 {
		finalMessage = Bold + Yellow + "System setup check completed with warnings, review them before proceeding." + Reset
		exitCode = 0
	} else {
		finalMessage = Bold + Green + "System setup check completed successfully! All required tools and configurations are in place." + Reset
		exitCode = 0
	}

	currentTime := time.Now().Format("02-01-2006 15:04:05")

	fmt.Printf("\n%s (Completed at: %s)\n", finalMessage, currentTime)
	os.Exit(exitCode)
}
