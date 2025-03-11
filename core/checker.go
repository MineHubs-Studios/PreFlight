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

func RunChecks(ctx context.Context) int {
	modules := SortModules(GetModules())
	categorizedResults := make([]CheckResult, 0, len(modules))
	ow := utils.NewOutputWriter()

	if !ow.Println(Bold + Blue + "\n╭─────────────────────────────────────────╮" + Reset) {
		return 0
	}

	if !ow.Println(Bold + Blue + "│" + Cyan + Bold + "  🚀 PreFlight Checker  " + Reset) {
		return 0
	}

	if !ow.Println(Bold + Blue + "╰─────────────────────────────────────────╯" + Reset) {
		return 0
	}

	if !ow.Println(Bold + "\nProcessing modules.." + Reset) {
		return 0
	}

	for _, module := range modules {
		select {
		case <-ctx.Done():
			if !ow.Println("\nChecks cancelled...") {
				return 0
			}
			return 1
		default:
		}

		moduleStart := time.Now()

		if !ow.Printf("  %s %s %s", Yellow+TimeGlass+Reset, Bold+module.Name()+Reset, Yellow+"..."+Reset) {
			return 0
		}

		errors, warnings, successes := module.CheckRequirements(ctx)

		moduleDuration := time.Since(moduleStart)

		if len(errors) == 0 && len(warnings) == 0 && len(successes) == 0 {
			if !ow.Printf("\r%s\r", strings.Repeat(" ", 50)) {
				return 0
			}

			continue
		}

		var statusColor, statusSymbol string

		if len(errors) > 0 {
			statusColor = Red
			statusSymbol = CrossMark
		} else if len(warnings) > 0 {
			statusColor = Yellow
			statusSymbol = WarningSign
		} else {
			statusColor = Green
			statusSymbol = CheckMark
		}

		if !ow.Printf("\r  %s %s completed (%dms)\n", statusColor+statusSymbol+Reset, Bold+module.Name()+Reset, moduleDuration.Milliseconds()) {
			return 0
		}

		result := CheckResult{
			Scope:     module.Name(),
			Errors:    errors,
			Warnings:  warnings,
			Successes: successes,
		}

		categorizedResults = append(categorizedResults, result)
	}

	if !ow.PrintNewLines(1) {
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

	var statusIcon, statusColor, statusText string
	var exitCode int

	if totalErrors > 0 {
		statusIcon = CrossMark
		statusColor = Red
		statusText = "Check completed, please resolve."
		exitCode = 1
	} else if totalWarnings > 0 {
		statusIcon = WarningSign
		statusColor = Yellow
		statusText = "Check completed with warnings, please review."
		exitCode = 0
	} else {
		statusIcon = CheckMark
		statusColor = Green
		statusText = "Check completed successfully!"
		exitCode = 0
	}

	currentTime := time.Now().Format("02-01-2006 15:04:05")

	fmt.Println(Bold + Blue + "\n╭────────────────────────────────────────────────────────────────╮" + Reset)
	fmt.Printf(Bold + Blue + "│ " + statusColor + statusIcon + " Status: " + statusText + Reset + "\n")
	fmt.Printf(Bold + Blue + "│ " + Dim + Clock + " Ended: " + currentTime + Reset + "\n")
	fmt.Println(Bold + Blue + "╰────────────────────────────────────────────────────────────────╯" + Reset)

	return exitCode
}
