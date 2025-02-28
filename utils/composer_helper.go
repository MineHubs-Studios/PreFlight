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
	var phpExtensions []string
	var composerDeps []string

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

	// EXTRACT "require" AND "require-dev" SECTIONS.
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

	if requireDev, ok := data["require-dev"].(map[string]interface{}); ok {
		for dep := range requireDev {
			composerDeps = append(composerDeps, dep)
		}
	}

	return phpVersion, phpExtensions, composerDeps, true
}
