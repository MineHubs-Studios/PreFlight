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

	if !ow.Println(utils.Bold + utils.Blue + "\n╭─────────────────────────────────────────╮" + utils.Reset) {
		return 0
	}

	if !ow.Println(utils.Bold + utils.Blue + "│" + utils.Cyan + utils.Bold + "  🚀 PreFlight Checker  " + utils.Reset) {
		return 0
	}

	if !ow.Println(utils.Bold + utils.Blue + "╰─────────────────────────────────────────╯" + utils.Reset) {
		return 0
	}

	if !ow.Println(utils.Bold + "\nProcessing modules.." + utils.Reset) {
		return 0
	}

	for _, module := range modules {
		select {
		case <-ctx.Done():
			if !ow.Println("\nChecks canceled...") {
				return 0
			}
			return 1
		default:
		}

		moduleStart := time.Now()

		if !ow.Printf("  %s %s %s", utils.Yellow+utils.TimeGlass+utils.Reset, utils.Bold+module.Name()+utils.Reset, utils.Yellow+"..."+utils.Reset) {
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
			statusColor = utils.Red
			statusSymbol = utils.CrossMark
		} else if len(warnings) > 0 {
			statusColor = utils.Yellow
			statusSymbol = utils.WarningSign
		} else {
			statusColor = utils.Green
			statusSymbol = utils.CheckMark
		}

		if !ow.Printf("\r  %s %s completed (%dms)\n", statusColor+statusSymbol+utils.Reset, utils.Bold+module.Name()+utils.Reset, moduleDuration.Milliseconds()) {
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

		sb.WriteString(utils.Reset)
		sb.WriteString(utils.Bold)
		sb.WriteString("\nScope: ")
		sb.WriteString(result.Scope)
		sb.WriteString(utils.Reset)

		if !ow.Println(sb.String()) {
			return
		}

		if len(result.Successes) > 0 {
			if !ow.Println(utils.Green + "  Successes:" + utils.Reset) {
				return
			}

			printMessages(ow, result.Successes, utils.Green, utils.CheckMark)

			if !ow.Println("") {
				return
			}
		}

		if len(result.Warnings) > 0 {
			if !ow.Println(utils.Yellow + "  Warnings:" + utils.Reset) {
				return
			}

			printMessages(ow, result.Warnings, utils.Yellow, utils.WarningSign)

			if !ow.Println("") {
				return
			}
		}

		if len(result.Errors) > 0 {
			if !ow.Println(utils.Red + "  Errors:" + utils.Reset) {
				return
			}

			printMessages(ow, result.Errors, utils.Red, utils.CrossMark)

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
			strings.Contains(msgLower, "installed package") ||
			strings.Contains(msgLower, "missing package") ||
			strings.Contains(msgLower, "installed extension") ||
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
		statusIcon = utils.CrossMark
		statusColor = utils.Red
		statusText = "Check completed, please resolve."
		exitCode = 1
	} else if totalWarnings > 0 {
		statusIcon = utils.WarningSign
		statusColor = utils.Yellow
		statusText = "Check completed with warnings, please review."
		exitCode = 0
	} else {
		statusIcon = utils.CheckMark
		statusColor = utils.Green
		statusText = "Check completed successfully!"
		exitCode = 0
	}

	currentTime := time.Now().Format("02-01-2006 15:04:05")

	fmt.Println(utils.Bold + utils.Blue + "\n╭────────────────────────────────────────────────────────────────╮" + utils.Reset)
	fmt.Println(utils.Bold + utils.Blue + "│ " + statusColor + statusIcon + " Status: " + statusText + utils.Reset)
	fmt.Println(utils.Bold + utils.Blue + "│ " + utils.Dim + utils.Clock + " Ended: " + currentTime + utils.Reset)
	fmt.Println(utils.Bold + utils.Blue + "╰────────────────────────────────────────────────────────────────╯" + utils.Reset)

	return exitCode
}
