package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ReadComposerJSON READ composer.json, PARSE REQUIRED PHP VERSION, EXTENSIONS, AND DEPENDENCIES.
func ReadComposerJSON() (string, []string, []string, bool) {
	var phpVersion string

	// PRE-ALLOCATES SLICES WITH SMALL CAPACITY TO AVOID REALLOCATIONS IN COMMON CASES.
	phpExtensions := make([]string, 0, 5)
	composerDeps := make([]string, 0, 10)

	// CHECK IF composer.json EXISTS.
	if _, err := os.Stat("composer.json"); os.IsNotExist(err) {
		return "", phpExtensions, composerDeps, false
	}

	// READ THE composer.json FILE.
	file, err := os.ReadFile("composer.json")

	if err != nil {
		return "", phpExtensions, composerDeps, false
	}

	// PARSE JSON CONTENT FROM composer.json.
	var data map[string]interface{}

	if err := json.Unmarshal(file, &data); err != nil {
		return "", phpExtensions, composerDeps, false
	}

	// CALCULATE THE TOTAL CAPACITY NEEDED FOR DEPENDENCIES TO MINIMIZE ALLOCATIONS.
	totalDeps := 0

	if require, ok := data["require"].(map[string]interface{}); ok {
		for dep := range require {
			if dep != "php" && !strings.HasPrefix(dep, "ext-") {
				totalDeps++
			}
		}
	}

	if requireDev, ok := data["require-dev"].(map[string]interface{}); ok {
		totalDeps += len(requireDev)
	}

	// RE-ALLOCATE WITH EXACT CAPACITY IF WE KNOW THE SIZE.
	if totalDeps > 0 {
		composerDeps = make([]string, 0, totalDeps)
	}

	// CALCULATE EXTENSION CAPACITY AND PRE-ALLOCATE IF POSSIBLE.
	extCount := 0

	if require, ok := data["require"].(map[string]interface{}); ok {
		for dep := range require {
			if strings.HasPrefix(dep, "ext-") {
				extCount++
			}
		}

		if extCount > 0 {
			phpExtensions = make([]string, 0, extCount)
		}
	}

	// EXTRACT DEPENDENCIES.
	if require, ok := data["require"].(map[string]interface{}); ok {
		for dep, version := range require {
			if dep == "php" {
				phpVersion = fmt.Sprintf("%v", version)
			} else if strings.HasPrefix(dep, "ext-") {
				phpExtensions = append(phpExtensions, strings.TrimPrefix(dep, "ext-"))
			} else {
				composerDeps = append(composerDeps, dep)
			}
		}
	}

	// EXTRACT DEV DEPENDENCIES.
	if requireDev, ok := data["require-dev"].(map[string]interface{}); ok {
		for dep := range requireDev {
			composerDeps = append(composerDeps, dep)
		}
	}

	return phpVersion, phpExtensions, composerDeps, true
}
