package core

import (
	"PreFlight/utils"
	"context"
	"fmt"
	"strings"
	"time"
)

type CheckResult struct {
	Scope     string
	Errors    []string
	Warnings  []string
	Successes []string
}

const (
	progressIncrement = 25
	progressSleep     = 200 * time.Millisecond
)

func showProgress(percent int) {
	fullBlocks := percent / 5
	emptyBlocks := 20 - fullBlocks

	var sb strings.Builder
	sb.Grow(50)

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

func RunChecks(ctx context.Context) int {
	modules := SortModules(GetModules())
	categorizedResults := make([]CheckResult, 0, len(modules))
	ow := utils.NewOutputWriter()

	if !ow.Println(Bold + "🚀 Running system setup checks...") {
		return 0
	}

	for _, module := range modules {
		// CHECK FOR CANCELLATION.
		select {
		case <-ctx.Done():
			if !ow.Println("\nChecks cancelled...") {
				return 0
			}

			return 1
		default:
		}

		moduleStart := time.Now()

		errors, warnings, successes := module.CheckRequirements(ctx)

		if len(errors) == 0 && len(warnings) == 0 && len(successes) == 0 {
			continue
		}

		if !ow.Printf(Bold+"\n🔍 Running checks for module: %s\n", module.Name()) {
			return 0
		}

		for progress := 0; progress <= 100; progress += progressIncrement {
			select {
			case <-ctx.Done():
				if !ow.Flush() {
					return 0
				}

				return 1
			default:
				showProgress(progress)
				time.Sleep(progressSleep)
			}
		}

		showProgress(100)

		result := CheckResult{
			Scope:     module.Name(),
			Errors:    errors,
			Warnings:  warnings,
			Successes: successes,
		}

		categorizedResults = append(categorizedResults, result)

		moduleDuration := time.Since(moduleStart)

		if !ow.Printf("\n      %s⏱ Completed in: %dms%s\n", Yellow, moduleDuration.Milliseconds(), Reset) {
			return 0
		}
	}

	if !ow.PrintNewLines(2) {
		return 0
	}

	printResults(categorizedResults)
	return finalMessage(categorizedResults)
}

func printResults(results []CheckResult) {
	ow := utils.NewOutputWriter()

	for _, result := range results {
		var sb strings.Builder

		sb.WriteString(Reset)
		sb.WriteString(Bold)
		sb.WriteString("\nScope: ")
		sb.WriteString(result.Scope)
		sb.WriteString(Reset)

		if !ow.Println(sb.String()) {
			return
		}

		if len(result.Successes) > 0 {
			if !ow.Println(Green + "  Successes:" + Reset) {
				return
			}

			printMessages(ow, result.Successes, Green, CheckMark)

			if !ow.Println("") {
				return
			}
		}

		if len(result.Warnings) > 0 {
			if !ow.Println(Yellow + "  Warnings:" + Reset) {
				return
			}

			printMessages(ow, result.Warnings, Yellow, WarningSign)

			if !ow.Println("") {
				return
			}
		}

		if len(result.Errors) > 0 {
			if !ow.Println(Red + "  Errors:" + Reset) {
				return
			}

			printMessages(ow, result.Errors, Red, CrossMark)

			if !ow.Println("") {
				return
			}
		}
	}
}

// printMessages
func printMessages(ow *utils.OutputWriter, messages []string, color string, symbol string) {
	var isUnderVersionMatch bool

	for _, msg := range messages {
		indentLevel := 2
		msgLower := strings.ToLower(msg)

		if strings.Contains(msgLower, "version matches") ||
			(strings.Contains(msgLower, "installed") &&
				(strings.Contains(msgLower, "php") ||
					strings.Contains(msgLower, "composer") ||
					strings.Contains(msgLower, "node") ||
					strings.Contains(msgLower, "go"))) {
			isUnderVersionMatch = true
			indentLevel = 4
		} else if strings.Contains(msg, "Scope:") {
			isUnderVersionMatch = false
			indentLevel = 2
		} else if strings.Contains(msg, ".json found") ||
			strings.Contains(msg, "go.mod found") {
			isUnderVersionMatch = true
			indentLevel = 4
		} else if !strings.Contains(msgLower, "version") &&
			!strings.Contains(msgLower, "installed") &&
			!strings.Contains(msgLower, ".json") &&
			!strings.Contains(msgLower, "go.mod") {
			isUnderVersionMatch = false
		}

		if isUnderVersionMatch && (strings.Contains(msgLower, "composer package") ||
			strings.Contains(msgLower, "npm package") ||
			strings.Contains(msgLower, "php extension") ||
			strings.Contains(msgLower, "go module")) {
			indentLevel = 6
		} else if !strings.Contains(msg, "Scope:") {
			indentLevel = 4
		}

		ow.Printf("%s%s %s %s\n", color, strings.Repeat(" ", indentLevel), symbol, msg)
	}
}

func finalMessage(results []CheckResult) int {
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
	return exitCode
}
